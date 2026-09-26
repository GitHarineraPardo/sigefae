package documento_radicado

import (
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"gorm.io/gorm"

	"sigefae/internal/db"
)

type Service struct {
	db *gorm.DB
}

func New(database *gorm.DB) *Service {
	return &Service{db: database}
}

func (s *Service) Create(dto CreateDTO, usuarioID uint) (*db.DocumentoRadicado, error) {
	var radicado db.DocumentoRadicado

	err := s.db.Transaction(func(tx *gorm.DB) error {
		// ── 1. Validar que el documento comercial existe ──
		var docCom db.DocumentoComercial
		if err := tx.First(&docCom, dto.DocumentoComercialID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("el documento comercial no existe")
			}
			return err
		}

		// ── 2. Validar que NO esté ya radicado ──
		var existing db.DocumentoRadicado
		if err := tx.Where("documento_comercial_id = ?", dto.DocumentoComercialID).First(&existing).Error; err == nil {
			return errors.New("el documento ya fue radicado anteriormente")
		}

		// ── 3. Validar catálogos ──
		var tipoRad db.TipoRadicacion
		if err := tx.First(&tipoRad, dto.TipoRadicacionID).Error; err != nil {
			return errors.New("el tipo de radicación no existe")
		}

		var ruta db.Ruta
		if err := tx.First(&ruta, dto.RutaID).Error; err != nil {
			return errors.New("la ruta no existe")
		}

		var metodoPago db.MetodoPago
		if err := tx.First(&metodoPago, dto.MetodoPagoID).Error; err != nil {
			return errors.New("el método de pago no existe")
		}

		// ── 4. Obtener primer paso de la ruta ──
		var primerPaso db.PasoRuta
		var primerPasoID *uint
		responsableID := &usuarioID
		var cargoActualID *uint
		estadoPosesion := "Libre"

		err := tx.Where("ruta_id = ?", dto.RutaID).Order("orden ASC, id ASC").First(&primerPaso).Error
		if err == nil {
			id := primerPaso.ID
			primerPasoID = &id
			if primerPaso.UsuarioID != nil || primerPaso.CargoID != nil {
				responsableID = primerPaso.UsuarioID
				cargoActualID = primerPaso.CargoID
			}
			estadoPosesion = "EnProceso"
		}

		// ── 5. Obtener estado inicial ──
		var estadoInicial db.EstadoDocumentoRadicado
		if err := tx.Order("id ASC").First(&estadoInicial).Error; err != nil {
			return errors.New("no hay estados de radicación configurados")
		}

		// ── 6. Generar número de radicado ──
		numeroRadicado := dto.NumeroRadicado
		if numeroRadicado == "" {
			year := time.Now().Year()
			var count int64
			tx.Model(&db.DocumentoRadicado{}).Where("YEAR(fecha_radicacion) = ?", year).Count(&count)
			prefix := "BU"
			if dto.EsMalambo {
				prefix = "MB"
			}
			numeroRadicado = fmt.Sprintf("%s-%d-%05d", prefix, year, count+1)
		}

		var existingNum db.DocumentoRadicado
		if err := tx.Where("numero_radicado = ?", numeroRadicado).First(&existingNum).Error; err == nil {
			return errors.New("el número de radicado ya existe")
		}

		// ── 7. Crear QR ──
		qr := db.CodigoQr{
			Url:    fmt.Sprintf("http://localhost:5173/radicado/%s", numeroRadicado),
			Activo: true,
		}
		if err := tx.Create(&qr).Error; err != nil {
			return err
		}

		// ── 8. Crear radicado ──
		radicado = db.DocumentoRadicado{
			DocumentoComercialID: dto.DocumentoComercialID,
			TipoRadicacionID:     dto.TipoRadicacionID,
			RutaID:               dto.RutaID,
			NumeroRadicado:       numeroRadicado,
			FechaRadicacion:      time.Now(),
			UsuarioActualID:      responsableID,
			CargoActualID:        cargoActualID,
			EstadoPosesion:       estadoPosesion,
			PasoActualID:         primerPasoID,
			EstadoID:             estadoInicial.ID,
			MetodoPagoID:         dto.MetodoPagoID,
			QrID:                 qr.ID,
		}

		if err := tx.Create(&radicado).Error; err != nil {
			return err
		}

		// ── 9. Generar tareas desde la ruta ──
		if err := generarTareasDesdeRuta(tx, &radicado, dto.RutaID, dto.DocumentoComercialID, dto.EsMalambo); err != nil {
			return err
		}

		// ── 10. Adjuntar archivos automáticamente desde el correo origen ──
		// Consultamos explícitamente correo_id para no depender del mapeo del struct
		var correoID *uint
		tx.Model(&db.DocumentoComercial{}).
			Select("correo_id").
			Where("id = ?", dto.DocumentoComercialID).
			Scan(&correoID)

		if correoID != nil && *correoID != 0 {
			if err := adjuntarArchivosDesdeCorreo(tx, &radicado, *correoID, usuarioID); err != nil {
				fmt.Printf("[RADICADO %s] warning: no se pudieron adjuntar archivos del correo: %v\n", radicado.NumeroRadicado, err)
			}
		} else {
			fmt.Printf("[RADICADO %s] info: documento comercial %d no tiene correo_id\n", radicado.NumeroRadicado, dto.DocumentoComercialID)
		}

		// ── 11. Validar total porcentajes normas + activos ──
		totalPctNormas := 0.0
		for _, n := range dto.NormasReparto {
			totalPctNormas += n.Porcentaje
		}
		totalPctActivos := 0.0
		for _, a := range dto.ActivosFijos {
			totalPctActivos += a.Porcentaje
		}
		if totalPctNormas+totalPctActivos > 100.01 {
			return fmt.Errorf("la suma de normas de reparto (%.2f%%) y activos fijos (%.2f%%) no puede superar el 100%%", totalPctNormas, totalPctActivos)
		}

		// ── 12. Crear normas de reparto (priorizar DTO, si no, buscar en proveedor) ──
		if len(dto.NormasReparto) > 0 {
			for _, n := range dto.NormasReparto {
				nr := db.RadicadoNormaReparto{
					DocumentoRadicadoID: radicado.ID,
					NormaRepartoID:      n.NormaRepartoID,
					Porcentaje:          n.Porcentaje,
					Proyecto:            n.Proyecto,
					Descripcion:         n.Descripcion,
					CreadoPorID:         usuarioID,
				}
				if err := tx.Create(&nr).Error; err != nil {
					return err
				}
			}
		} else {
			// Buscar normas por defecto del proveedor para esta ruta
			if docCom.IDProveedor != 0 {
				var proveedorNormas []db.ProveedorNormaReparto
				if err := tx.Where("proveedor_id = ? AND ruta_id = ?", docCom.IDProveedor, dto.RutaID).Find(&proveedorNormas).Error; err == nil && len(proveedorNormas) > 0 {
					for _, pn := range proveedorNormas {
						nr := db.RadicadoNormaReparto{
							DocumentoRadicadoID: radicado.ID,
							NormaRepartoID:      pn.NormaRepartoID,
							Porcentaje:          pn.Porcentaje,
							CreadoPorID:         usuarioID,
						}
						if err := tx.Create(&nr).Error; err != nil {
							return err
						}
					}
				}
			}
		}

		// ── 13. Crear activos fijos si vienen en DTO ──
		if len(dto.ActivosFijos) > 0 {
			for _, a := range dto.ActivosFijos {
				af := db.RadicadoActivoFijo{
					DocumentoRadicadoID: radicado.ID,
					Nombre:              a.Nombre,
					Sucursal:            a.Sucursal,
					Proyecto:            a.Proyecto,
					Porcentaje:          a.Porcentaje,
					Descripcion:         a.Descripcion,
					CreadoPorID:         usuarioID,
				}
				if err := tx.Create(&af).Error; err != nil {
					return err
				}
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// ── 11. Preload completo fuera de la transacción ──
	if err := s.db.
		Preload("DocumentoComercial").
		Preload("DocumentoComercial.Proveedor").
		Preload("DocumentoComercial.Receptor").
		Preload("DocumentoComercial.Area").
		Preload("DocumentoComercial.Moneda").
		Preload("DocumentoComercial.Detalles").
		Preload("TipoRadicacion").
		Preload("Ruta").
		Preload("UsuarioActual").
		Preload("CargoActual").
		Preload("PasoActual").
		Preload("PasoActual.Usuario").
		Preload("Estado").
		Preload("MetodoPago").
		Preload("Qr").
		Preload("Archivos").
		Preload("Archivos.Origen").
		Preload("NormasReparto").
		Preload("NormasReparto.NormaReparto").
		First(&radicado, radicado.ID).Error; err != nil {
		return nil, err
	}

	return &radicado, nil
}

func (s *Service) List() ([]db.DocumentoRadicado, error) {
	var radicados []db.DocumentoRadicado
	if err := s.db.
		Preload("DocumentoComercial").
		Preload("DocumentoComercial.Proveedor").
		Preload("DocumentoComercial.Receptor").
		Preload("TipoRadicacion").
		Preload("Ruta").
		Preload("MetodoPago").
		Preload("Estado").
		Preload("PasoActual").
		Preload("UsuarioActual").
		Preload("CargoActual").
		Preload("Archivos").
		Preload("Qr").
		Preload("NormasReparto").
		Preload("NormasReparto.NormaReparto").
		Order("fecha_radicacion DESC").
		Find(&radicados).Error; err != nil {
		return nil, err
	}
	return radicados, nil
}

func (s *Service) GetByID(id uint) (*db.DocumentoRadicado, error) {
	var radicado db.DocumentoRadicado
	if err := s.db.
		Preload("DocumentoComercial").
		Preload("DocumentoComercial.Proveedor").
		Preload("DocumentoComercial.Receptor").
		Preload("TipoRadicacion").
		Preload("Ruta").
		Preload("MetodoPago").
		Preload("Estado").
		Preload("PasoActual").
		Preload("UsuarioActual").
		Preload("CargoActual").
		Preload("Archivos").
		Preload("Archivos.CreadoPor").
		Preload("Qr").
		Preload("NormasReparto").
		Preload("NormasReparto.NormaReparto").
		First(&radicado, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("documento radicado no encontrado")
		}
		return nil, err
	}
	return &radicado, nil
}

func (s *Service) Update(id uint, dto UpdateDTO) (*db.DocumentoRadicado, error) {
	var radicado db.DocumentoRadicado
	if err := s.db.First(&radicado, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("documento radicado no encontrado")
		}
		return nil, err
	}

	// Si el estado_posesion fue marcado como 'Devuelto', marcar en trazabilidad
	if radicado.EstadoPosesion == "Devuelto" {
		t := time.Now()
		s.db.Create(&db.Trazabilidad{
			DocumentoRadicadoID: radicado.ID,
			UsuarioID:           0,
			Accion:              "Rechazado",
			Descripcion:         "Documento marcado como rechazado",
			Fecha:               t,
		})
	}

	updates := map[string]any{}
	if dto.TipoRadicacionID != 0 {
		updates["tipo_radicacion_id"] = dto.TipoRadicacionID
	}
	if dto.RutaID != 0 {
		updates["ruta_id"] = dto.RutaID
	}
	if dto.MetodoPagoID != 0 {
		updates["metodo_pago_id"] = dto.MetodoPagoID
	}
	if dto.NumeroRadicado != "" {
		updates["numero_radicado"] = dto.NumeroRadicado
	}
	if dto.UsuarioActualID != 0 {
		updates["usuario_actual_id"] = dto.UsuarioActualID
	}
	if dto.EstadoPosesion != "" {
		updates["estado_posesion"] = dto.EstadoPosesion
	}
	if dto.PasoActualID != 0 {
		updates["paso_actual_id"] = dto.PasoActualID
	}
	if dto.EstadoID != 0 {
		updates["estado_id"] = dto.EstadoID
	}

	if len(updates) > 0 {
		if err := s.db.Model(&radicado).Updates(updates).Error; err != nil {
			return nil, err
		}
	}

	if err := s.db.
		Preload("DocumentoComercial").
		Preload("DocumentoComercial.Proveedor").
		Preload("DocumentoComercial.Receptor").
		Preload("TipoRadicacion").
		Preload("Ruta").
		Preload("MetodoPago").
		Preload("Estado").
		Preload("PasoActual").
		Preload("UsuarioActual").
		Preload("CargoActual").
		Preload("Archivos").
		Preload("Qr").
		Preload("NormasReparto").
		Preload("NormasReparto.NormaReparto").
		First(&radicado, id).Error; err != nil {
		return nil, err
	}
	return &radicado, nil
}

// ─────────────────────────────────────────────────────────────
// generarTareasDesdeRuta
// ─────────────────────────────────────────────────────────────
func generarTareasDesdeRuta(tx *gorm.DB, radicado *db.DocumentoRadicado, rutaID uint, docComercialID uint, esMalambo bool) error {
	// ── 1. Pasos base de la ruta ──
	var pasos []db.PasoRuta
	if err := tx.Where("ruta_id = ? AND activo = ?", rutaID, true).Order("orden asc, id asc").Find(&pasos).Error; err != nil {
		return err
	}
	if len(pasos) == 0 {
		return errors.New("la ruta no tiene pasos configurados")
	}

	// ── 2. Documento comercial (monto, área, moneda) ──
	var docCom db.DocumentoComercial
	if err := tx.First(&docCom, docComercialID).Error; err != nil {
		return err
	}

	// ── 3. Verificar si hay reglas de monto activas aplicables ──
	var countReglas int64
	tx.Model(&db.ReglaMontoRuta{}).
		Where("activo = ?", true).
		Count(&countReglas)

	var reglas []db.ReglaMontoRuta

	if countReglas > 0 {
		// Obtener SMMLV solo si hay reglas que evaluar
		var smmlv db.SalarioMinimo
		if err := tx.Where("activo = ?", true).Order("ano desc").First(&smmlv).Error; err != nil {
			return errors.New("no hay un salario mínimo vigente configurado en el sistema")
		}
		if smmlv.Valor <= 0 {
			return errors.New("el salario mínimo vigente debe ser mayor a 0")
		}
		docTotalEnSmmlv := docCom.Total / smmlv.Valor

		// ── 4. Obtener reglas de monto aplicables por SMMLV ──
		query := tx.Where("activo = ? AND COALESCE(monto_minimo_smmlv, 0) <= ?", true, docTotalEnSmmlv).
			Where("(COALESCE(monto_maximo_smmlv, 0) = 0 OR monto_maximo_smmlv >= ?)", docTotalEnSmmlv).
			Preload("UsuarioAprobador").
			Preload("RolAprobador").
			Order("monto_minimo_smmlv desc")

		if err := query.Find(&reglas).Error; err != nil {
			return err
		}
	}

	// ── 5. Clasificar reglas por posición ──
	var alInicio, antesDelFinal, alFinal []db.ReglaMontoRuta
	for _, r := range reglas {
		switch r.PosicionInsercion {
		case "PRIMERO":
			alInicio = append(alInicio, r)
		case "ANTES_FINAL":
			antesDelFinal = append(antesDelFinal, r)
		default: // ULTIMO
			alFinal = append(alFinal, r)
		}
	}

	// Ordenar cada grupo por prioridad (menor = va primero)
	sort.Slice(alInicio, func(i, j int) bool { return alInicio[i].Prioridad < alInicio[j].Prioridad })
	sort.Slice(antesDelFinal, func(i, j int) bool { return antesDelFinal[i].Prioridad < antesDelFinal[j].Prioridad })
	sort.Slice(alFinal, func(i, j int) bool { return alFinal[i].Prioridad < alFinal[j].Prioridad })

	// ── 5. Helper para resolver asignación de una regla ──
	resolverAsignacion := func(r db.ReglaMontoRuta) (*uint, *uint) {
		if r.CargoAprobadorID != nil {
			return nil, r.CargoAprobadorID
		}
		if r.UsuarioAprobadorID != nil {
			return r.UsuarioAprobadorID, nil
		}
		if r.RolAprobadorID != nil {
			var usr db.Usuario
			if err := tx.Where("id_rol = ? AND activo = ?", *r.RolAprobadorID, true).First(&usr).Error; err == nil {
				return &usr.ID, nil
			}
		}
		return nil, nil
	}

	formatoRegla := func(r db.ReglaMontoRuta) string {
		if r.MontoMaximoSmmlv > 0 {
			return fmt.Sprintf("Aprobación por monto (%.2f - %.2f SMMLV)", r.MontoMinimoSmmlv, r.MontoMaximoSmmlv)
		}
		return fmt.Sprintf("Aprobación por monto (≥ %.2f SMMLV)", r.MontoMinimoSmmlv)
	}

	type pasoFinal struct {
		Nombre    string
		UsuarioID *uint
		CargoID   *uint
		EsRegla   bool
	}

	var flujo []pasoFinal

	// 5.1 PRIMERO
	for _, r := range alInicio {
		uid, cargoID := resolverAsignacion(r)
		flujo = append(flujo, pasoFinal{
			Nombre:    formatoRegla(r),
			UsuarioID: uid,
			CargoID:   cargoID,
			EsRegla:   true,
		})
	}

	// 5.2 Pasos base (todos menos el último)
	for i := 0; i < len(pasos)-1; i++ {
		p := pasos[i]
		flujo = append(flujo, pasoFinal{
			Nombre:    p.Nombre,
			UsuarioID: p.UsuarioID,
			CargoID:   p.CargoID,
			EsRegla:   false,
		})
	}

	// 5.3 ANTES_FINAL: justo antes del cierre del flujo base
	for _, r := range antesDelFinal {
		uid, cargoID := resolverAsignacion(r)
		flujo = append(flujo, pasoFinal{
			Nombre:    formatoRegla(r),
			UsuarioID: uid,
			CargoID:   cargoID,
			EsRegla:   true,
		})
	}

	// 5.4 Último paso base (el cierre)
	ultimoPaso := pasos[len(pasos)-1]
	flujo = append(flujo, pasoFinal{
		Nombre:    ultimoPaso.Nombre,
		UsuarioID: ultimoPaso.UsuarioID,
		CargoID:   ultimoPaso.CargoID,
		EsRegla:   false,
	})

	// 5.5 ULTIMO: después de todo
	for _, r := range alFinal {
		uid, cargoID := resolverAsignacion(r)
		flujo = append(flujo, pasoFinal{
			Nombre:    formatoRegla(r),
			UsuarioID: uid,
			CargoID:   cargoID,
			EsRegla:   true,
		})
	}

	// ── 5.6 Obtener IDs de Cargos fijos ──
	var cargoAuxOriente, cargoAuxNorte, cargoAnalistaNorte, cargoAnalistaOriente db.Cargo
	tx.Where("nombre = ?", "Auxiliar Compras Oriente").First(&cargoAuxOriente)
	tx.Where("nombre = ?", "Auxiliar Compras Norte").First(&cargoAuxNorte)
	tx.Where("nombre = ?", "Analista Compras Norte").First(&cargoAnalistaNorte)
	tx.Where("nombre = ?", "Analista Compras Oriente").First(&cargoAnalistaOriente)

	// ── 5.7 PREFIJO Y SUFIJO POR ZONA ──
	// BU: [Aux Compras Oriente (Jennyfer)] → [Analista Compras Oriente] → pasos → [Aux Compras Oriente]
	// MB: [Aux Compras Oriente (Jennyfer)] → [Aux Compras Norte] → [Analista Compras Norte] → pasos → [Aux Compras Oriente]
	var prefijoFlujo []pasoFinal
	var sufijo1Flujo []pasoFinal

	// ── Jennyfer (Aux Compras Oriente) siempre va de PRIMERO en ambas zonas ──
	prefijoFlujo = append(prefijoFlujo, pasoFinal{
		Nombre:  "Revisión Auxiliar Compras Oriente",
		CargoID: &cargoAuxOriente.ID,
		EsRegla: false,
	})

	if esMalambo {
		// MB: después de Jennyfer va Aux Compras Norte, luego Analista Compras Norte, luego pasos
		prefijoFlujo = append(prefijoFlujo, pasoFinal{
			Nombre:  "Revisión Auxiliar Compras Norte",
			CargoID: &cargoAuxNorte.ID,
			EsRegla: false,
		})
		prefijoFlujo = append(prefijoFlujo, pasoFinal{
			Nombre:  "Revisión Analista Compras Norte",
			CargoID: &cargoAnalistaNorte.ID,
			EsRegla: false,
		})
	} else {
		// BU: después de Jennyfer va Analista Compras Oriente, luego pasos
		prefijoFlujo = append(prefijoFlujo, pasoFinal{
			Nombre:  "Revisión Analista Compras Oriente",
			CargoID: &cargoAnalistaOriente.ID,
			EsRegla: false,
		})
	}

	flujo = append(prefijoFlujo, flujo...)
	flujo = append(flujo, sufijo1Flujo...)

	// ── 5.8 PASO FINAL ABSOLUTO ──
	flujo = append(flujo, pasoFinal{
		Nombre:  "Revisión Auxiliar Compras Oriente",
		CargoID: &cargoAuxOriente.ID,
		EsRegla: false,
	})

	// ── 6. Estados base ──
	var estadoPendiente db.EstadoTarea
	if err := tx.Where("nombre = ?", "Pendiente").First(&estadoPendiente).Error; err != nil {
		return errors.New("no se encontró el estado 'Pendiente' para tareas")
	}
	var estadoEnProceso db.EstadoTarea
	if err := tx.Where("nombre = ?", "En Proceso").First(&estadoEnProceso).Error; err != nil {
		return errors.New("no se encontró el estado 'En Proceso' para tareas")
	}

	now := time.Now()

	// ── 7. Crear tareas en orden ──
	for i, pf := range flujo {
		var uid *uint
		cargoID := pf.CargoID

		// Si no tiene asignación específica, heredar del radicado (creador)
		if pf.UsuarioID == nil && pf.CargoID == nil && i == 0 {
			uid = radicado.UsuarioActualID
			cargoID = radicado.CargoActualID
		} else {
			uid = pf.UsuarioID
		}

		tarea := db.Tarea{
			DocumentoRadicadoID: radicado.ID,
			UsuarioAsignadoID:   uid,
			CargoAsignadoID:     cargoID,
			EstadoID:            estadoPendiente.ID,
			Descripcion:         pf.Nombre,
			FechaAsignacion:     now,
		}
		if i == 0 {
			tarea.EstadoID = estadoEnProceso.ID
			tarea.FechaInicio = &now
			radicado.UsuarioActualID = uid
			radicado.CargoActualID = cargoID
		}
		if err := tx.Create(&tarea).Error; err != nil {
			return err
		}
	}

	return tx.Save(radicado).Error
}

// ─────────────────────────────────────────────────────────────
// adjuntarArchivosDesdeCorreo
// ─────────────────────────────────────────────────────────────
func adjuntarArchivosDesdeCorreo(tx *gorm.DB, radicado *db.DocumentoRadicado, correoID uint, usuarioID uint) error {
	// 1. Cargar el correo para obtener id_mensaje
	var correo db.Correo
	if err := tx.First(&correo, correoID).Error; err != nil {
		return fmt.Errorf("correo origen no encontrado (id=%d): %w", correoID, err)
	}
	if correo.IDMensaje == "" {
		return errors.New("correo no tiene id_mensaje definido")
	}

	// 2. Buscar o crear origen "Sistema"
	var origen db.ArchivoOrigen
	if err := tx.Where("nombre = ?", "Sistema").First(&origen).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			origen = db.ArchivoOrigen{Nombre: "Sistema", Activo: true}
			if err := tx.Create(&origen).Error; err != nil {
				return fmt.Errorf("no se pudo crear origen 'Sistema': %w", err)
			}
		} else {
			return fmt.Errorf("error buscando origen 'Sistema': %w", err)
		}
	}
	origenID := origen.ID

	// 3. Directorio fuente
	srcDir := filepath.Join("storage", "mails", correo.IDMensaje)
	entries, err := os.ReadDir(srcDir)
	if err != nil {
		return fmt.Errorf("no se pudo leer directorio origen %s: %w", srcDir, err)
	}

	// 4. Directorio destino
	dstDir := filepath.Join("storage", "radicados", fmt.Sprintf("%d", radicado.ID))
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return fmt.Errorf("no se pudo crear directorio destino %s: %w", dstDir, err)
	}

	// 5. Copiar cada archivo (solo PDF)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		ext := strings.ToLower(filepath.Ext(name))

		// ← Solo archivos PDF
		if ext != ".pdf" {
			continue
		}

		srcPath := filepath.Join(srcDir, name)
		dstPath := filepath.Join(dstDir, name)

		if err := copyFile(srcPath, dstPath); err != nil {
			fmt.Printf("[RADICADO %d] ERROR copiando %s: %v\n", radicado.ID, name, err)
			continue
		}

		info, _ := os.Stat(dstPath)
		peso := int64(0)
		if info != nil {
			peso = info.Size()
		}

		archivo := db.Archivo{
			DocumentoRadicadoID: radicado.ID,
			Nombre:              name,
			Extension:           strings.TrimPrefix(ext, "."),
			Ruta:                dstPath,
			Peso:                peso,
			OrigenID:            origenID,
			CreadoPorID:         usuarioID,
		}
		if err := tx.Create(&archivo).Error; err != nil {
			fmt.Printf("[RADICADO %d] ERROR guardando registro BD de %s: %v\n", radicado.ID, name, err)
			continue
		}
	}

	return nil
}

func copyFile(src, dst string) error {
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)
	return err
}

func (s *Service) GetByNumeroRadicado(numero string) (*db.DocumentoRadicado, error) {
	var radicado db.DocumentoRadicado
	if err := s.db.
		Preload("DocumentoComercial").
		Preload("DocumentoComercial.Proveedor").
		Preload("DocumentoComercial.Receptor").
		Preload("DocumentoComercial.Moneda").
		Preload("TipoRadicacion").
		Preload("Ruta").
		Preload("MetodoPago").
		Preload("Estado").
		Preload("PasoActual").
		Preload("UsuarioActual").
		Preload("CargoActual").
		Preload("Qr").
		Preload("NormasReparto").
		Preload("NormasReparto.NormaReparto").
		Where("numero_radicado = ?", numero).
		First(&radicado).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("documento no encontrado")
		}
		return nil, err
	}
	return &radicado, nil
}

func (s *Service) GetNormasReparto(radicadoID uint) ([]db.RadicadoNormaReparto, error) {
	var normas []db.RadicadoNormaReparto
	if err := s.db.Where("documento_radicado_id = ?", radicadoID).
		Preload("NormaReparto").
		Find(&normas).Error; err != nil {
		return nil, err
	}
	return normas, nil
}

func (s *Service) AsignarNormasReparto(radicadoID uint, dtos []NormaRepartoInputDTO, usuarioID uint, isAdmin bool) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var existentes []db.RadicadoNormaReparto
		if err := tx.Where("documento_radicado_id = ?", radicadoID).Find(&existentes).Error; err != nil {
			return err
		}

		// Calcular total normas DTO
		totalNormas := 0.0
		for _, d := range dtos {
			totalNormas += d.Porcentaje
		}

		// Calcular total activos fijos existentes
		var activosExistentes []db.RadicadoActivoFijo
		if err := tx.Where("documento_radicado_id = ?", radicadoID).Find(&activosExistentes).Error; err != nil {
			return err
		}
		totalActivos := 0.0
		for _, a := range activosExistentes {
			totalActivos += a.Porcentaje
		}

		gran_total := totalNormas + totalActivos
		if gran_total > 100.01 {
			return fmt.Errorf("la suma de normas de reparto (%.2f%%) + activos fijos (%.2f%%) no puede superar el 100%%, actualmente sumaría %.2f%%", totalNormas, totalActivos, gran_total)
		}

		porNorma := map[uint]db.RadicadoNormaReparto{}
		for _, e := range existentes {
			porNorma[e.NormaRepartoID] = e
		}

		dtoMap := map[uint]NormaRepartoInputDTO{}
		for _, d := range dtos {
			dtoMap[d.NormaRepartoID] = d
		}

		// Validar que no-admins no modifiquen normas ajenas
		if !isAdmin {
			for _, e := range existentes {
				if e.CreadoPorID != 0 && e.CreadoPorID != usuarioID {
					if dto, ok := dtoMap[e.NormaRepartoID]; ok && math.Abs(dto.Porcentaje-e.Porcentaje) > 0.0001 {
						return errors.New("No puedes modificar una norma que no creaste. Solicita permiso al propietario.")
					}
				}
			}
		}

		// Crear o actualizar normas del DTO
		for _, d := range dtos {
			existente, ok := porNorma[d.NormaRepartoID]
			if ok {
				if isAdmin || existente.CreadoPorID == 0 || existente.CreadoPorID == usuarioID {
					existente.Porcentaje = d.Porcentaje
					existente.Proyecto = d.Proyecto
					existente.Descripcion = d.Descripcion
					if existente.CreadoPorID == 0 {
						existente.CreadoPorID = usuarioID
					}
					if err := tx.Save(&existente).Error; err != nil {
						return err
					}
				}
				continue
			}

			nr := db.RadicadoNormaReparto{
				DocumentoRadicadoID: radicadoID,
				NormaRepartoID:      d.NormaRepartoID,
				Porcentaje:          d.Porcentaje,
				Proyecto:            d.Proyecto,
				Descripcion:         d.Descripcion,
				CreadoPorID:         usuarioID,
			}
			if err := tx.Create(&nr).Error; err != nil {
				return err
			}
		}

		// Eliminar normas que ya no están en el DTO
		// Admins pueden eliminar cualquier norma; no-admins solo las propias
		for _, e := range existentes {
			if _, exists := dtoMap[e.NormaRepartoID]; !exists {
				if isAdmin || e.CreadoPorID == 0 || e.CreadoPorID == usuarioID {
					if err := tx.Delete(&e).Error; err != nil {
						return err
					}
				}
			}
		}

		return nil
	})
}

// ─────────────────────────────────────────────────────────────
// MemorizarNormasProveedorRuta
// ─────────────────────────────────────────────────────────────
// Auto-aprendizaje: Lee las normas finales de un documento y reemplaza la plantilla del proveedor + ruta
func (s *Service) MemorizarNormasProveedorRuta(radicadoID uint) error {
	var radicado db.DocumentoRadicado
	if err := s.db.Preload("DocumentoComercial").Preload("NormasReparto").First(&radicado, radicadoID).Error; err != nil {
		return err
	}

	proveedorID := radicado.DocumentoComercial.IDProveedor
	rutaID := radicado.RutaID

	if proveedorID == 0 || rutaID == 0 {
		return nil // No se puede memorizar sin proveedor o ruta
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		// 1. Eliminar la plantilla anterior para ese proveedor y ruta
		if err := tx.Where("proveedor_id = ? AND ruta_id = ?", proveedorID, rutaID).Delete(&db.ProveedorNormaReparto{}).Error; err != nil {
			return err
		}

		// 2. Insertar las normas actuales del documento como la nueva plantilla
		for _, nr := range radicado.NormasReparto {
			pnr := db.ProveedorNormaReparto{
				ProveedorID:    proveedorID,
				RutaID:         rutaID,
				NormaRepartoID: nr.NormaRepartoID,
				Porcentaje:     nr.Porcentaje,
			}
			if err := tx.Create(&pnr).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// ─────────────────────────────────────────────────────────────
// Activos Fijos
// ─────────────────────────────────────────────────────────────

type ActivoFijoInputDTO struct {
	ID          uint    `json:"id"`
	Nombre      string  `json:"nombre" binding:"required"`
	Sucursal    string  `json:"sucursal" binding:"required"`
	Proyecto    string  `json:"proyecto"`
	Porcentaje  float64 `json:"porcentaje" binding:"required"`
	Descripcion string  `json:"descripcion"`
}

func (s *Service) GetActivosFijos(radicadoID uint) ([]db.RadicadoActivoFijo, error) {
	var items []db.RadicadoActivoFijo
	if err := s.db.Where("documento_radicado_id = ?", radicadoID).
		Preload("CreadoPor").
		Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (s *Service) AsignarActivosFijos(radicadoID uint, dtos []ActivoFijoInputDTO, usuarioID uint, isAdmin bool) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// Calcular total normas de reparto para este radicado
		var normas []db.RadicadoNormaReparto
		if err := tx.Where("documento_radicado_id = ?", radicadoID).Find(&normas).Error; err != nil {
			return err
		}
		totalNormas := 0.0
		for _, n := range normas {
			totalNormas += n.Porcentaje
		}

		// Calcular total activos fijos del DTO
		totalActivos := 0.0
		for _, d := range dtos {
			totalActivos += d.Porcentaje
		}

		// Validar que la suma total no supere el 100%
		gran_total := totalNormas + totalActivos
		if gran_total > 100.01 {
			return fmt.Errorf("la suma de normas de reparto (%.2f%%) + activos fijos (%.2f%%) no puede superar el 100%%, actualmente sumaría %.2f%%", totalNormas, totalActivos, gran_total)
		}

		// Obtener existentes
		var existentes []db.RadicadoActivoFijo
		if err := tx.Where("documento_radicado_id = ?", radicadoID).Find(&existentes).Error; err != nil {
			return err
		}

		// Mapa por ID para updates
		porID := map[uint]db.RadicadoActivoFijo{}
		for _, e := range existentes {
			porID[e.ID] = e
		}

		// IDs del DTO para saber cuáles eliminar
		dtoIDs := map[uint]bool{}
		for _, d := range dtos {
			if d.ID > 0 {
				dtoIDs[d.ID] = true
			}
		}

		// Crear o actualizar
		for _, d := range dtos {
			if d.ID > 0 {
				existente, ok := porID[d.ID]
				if ok {
					if isAdmin || existente.CreadoPorID == 0 || existente.CreadoPorID == usuarioID {
						existente.Nombre = d.Nombre
						existente.Sucursal = d.Sucursal
						existente.Proyecto = d.Proyecto
						existente.Porcentaje = d.Porcentaje
						existente.Descripcion = d.Descripcion
						if err := tx.Save(&existente).Error; err != nil {
							return err
						}
					}
					continue
				}
			}
			// Crear nuevo
			af := db.RadicadoActivoFijo{
				DocumentoRadicadoID: radicadoID,
				Nombre:              d.Nombre,
				Sucursal:            d.Sucursal,
				Proyecto:            d.Proyecto,
				Porcentaje:          d.Porcentaje,
				Descripcion:         d.Descripcion,
				CreadoPorID:         usuarioID,
			}
			if err := tx.Create(&af).Error; err != nil {
				return err
			}
		}

		// Eliminar los que ya no están en el DTO
		for _, e := range existentes {
			if !dtoIDs[e.ID] {
				if isAdmin || e.CreadoPorID == 0 || e.CreadoPorID == usuarioID {
					if err := tx.Delete(&e).Error; err != nil {
						return err
					}
				}
			}
		}

		return nil
	})
}
