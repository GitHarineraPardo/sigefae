package gasto_proveedor

import (
	"math"
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

func (s *Service) GetGastos(user db.Usuario, filtroCargoID *uint, fechaInicio, fechaFin string) (*GastosResponse, error) {
	// Determinar el cargo a filtrar
	var cargoID *uint
	var cargoNombre string = "General"

	esSuperadmin := false
	if user.Rol != nil && user.Rol.Nombre == "Superadministrador" {
		esSuperadmin = true
	} else if user.RolID == 1 {
		esSuperadmin = true
	}

	if filtroCargoID != nil && *filtroCargoID > 0 {
		cargoID = filtroCargoID
	} else if !esSuperadmin {
		cargoID = user.CargoID
	}

	if cargoID != nil && *cargoID > 0 {
		var c db.Cargo
		if err := s.db.First(&c, *cargoID).Error; err == nil {
			cargoNombre = c.Nombre
		}
	} else if esSuperadmin {
		cargoNombre = "Todos los Cargos (Administración)"
	}

	// Consulta de radicados completados
	query := s.db.Model(&db.DocumentoRadicado{}).
		Preload("DocumentoComercial").
		Preload("DocumentoComercial.Proveedor").
		Preload("DocumentoComercial.Detalles", func(tx *gorm.DB) *gorm.DB {
			return tx.Where("activo = ?", true)
		}).
		Preload("NormasReparto").
		Preload("NormasReparto.NormaReparto").
		Where("documento_radicado.estado_posesion = ?", "Completado")

	// Filtro por aprobación del cargo o usuario
	if cargoID != nil && *cargoID > 0 {
		query = query.Where(`(
			EXISTS (SELECT 1 FROM tarea t WHERE t.documento_radicado_id = documento_radicado.id AND t.cargo_asignado_id = ? AND t.estado_id = 3)
			OR EXISTS (SELECT 1 FROM trazabilidad tz JOIN usuario u ON u.id = tz.usuario_id WHERE tz.documento_radicado_id = documento_radicado.id AND u.cargo_id = ?)
		)`, *cargoID, *cargoID)
	} else if !esSuperadmin {
		query = query.Where(`(
			EXISTS (SELECT 1 FROM tarea t WHERE t.documento_radicado_id = documento_radicado.id AND t.usuario_asignado_id = ? AND t.estado_id = 3)
			OR EXISTS (SELECT 1 FROM trazabilidad tz WHERE tz.documento_radicado_id = documento_radicado.id AND tz.usuario_id = ?)
		)`, user.ID, user.ID)
	}

	// Filtros por fecha
	if fechaInicio != "" {
		if tIni, err := time.Parse("2006-01-02", fechaInicio); err == nil {
			query = query.Where("documento_radicado.fecha_radicacion >= ?", tIni)
		}
	}
	if fechaFin != "" {
		if tFin, err := time.Parse("2006-01-02", fechaFin); err == nil {
			tFin = tFin.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
			query = query.Where("documento_radicado.fecha_radicacion <= ?", tFin)
		}
	}

	var radicados []db.DocumentoRadicado
	if err := query.Order("documento_radicado.fecha_radicacion DESC").Find(&radicados).Error; err != nil {
		return nil, err
	}

	// Procesar agregados
	resp := &GastosResponse{
		CargoID:     cargoID,
		CargoNombre: cargoNombre,
		Proveedores: make([]ProveedorGastoDTO, 0),
		Items:       make([]ItemGastoDTO, 0),
		Normas:      make([]NormaGastoDTO, 0),
		Radicados:   make([]RadicadoGastoDTO, 0),
	}

	proveedoresMap := make(map[uint]*ProveedorGastoDTO)
	itemsMap := make(map[string]*ItemGastoDTO)
	normasMap := make(map[string]*NormaGastoDTO)

	for _, rad := range radicados {
		if rad.DocumentoComercial == nil {
			continue
		}

		doc := rad.DocumentoComercial
		resp.TotalGeneral += doc.Total
		resp.SubtotalGeneral += doc.Subtotal
		resp.IvaGeneral += doc.Iva
		resp.CantidadFacturas++

		// Proveedor
		var provID uint = doc.IDProveedor
		var provRazonSocial string = "Proveedor Desconocido"
		var provNIT string = "S/N"
		if doc.Proveedor != nil {
			provID = doc.Proveedor.ID
			provRazonSocial = doc.Proveedor.RazonSocial
			provNIT = doc.Proveedor.NumeroDocumento
		}

		if p, ok := proveedoresMap[provID]; ok {
			p.TotalGastado += doc.Total
			p.TotalSubtotal += doc.Subtotal
			p.TotalIva += doc.Iva
			p.CantidadFacturas++
		} else {
			proveedoresMap[provID] = &ProveedorGastoDTO{
				ID:               provID,
				RazonSocial:      provRazonSocial,
				NIT:              provNIT,
				TotalGastado:     doc.Total,
				TotalSubtotal:    doc.Subtotal,
				TotalIva:         doc.Iva,
				CantidadFacturas: 1,
			}
		}

		// Ítems de la factura
		if len(doc.Detalles) > 0 {
			for _, det := range doc.Detalles {
				descNorm := strings.TrimSpace(det.Descripcion)
				if descNorm == "" {
					descNorm = "Ítem sin descripción"
				}
				itemKey := strings.ToUpper(descNorm)

				montoDetalle := det.Total
				if montoDetalle == 0 && det.ValorUnit > 0 {
					montoDetalle = det.ValorUnit * det.Cantidad
				}

				if item, ok := itemsMap[itemKey]; ok {
					item.CantidadTotal += det.Cantidad
					item.TotalGastado += montoDetalle
					item.CantidadFacturas++
				} else {
					itemsMap[itemKey] = &ItemGastoDTO{
						Descripcion:      descNorm,
						CantidadTotal:    det.Cantidad,
						TotalGastado:     montoDetalle,
						ProveedorID:      provID,
						ProveedorNombre:  provRazonSocial,
						CantidadFacturas: 1,
					}
				}
			}
		} else {
			// Si no hay líneas de detalle, usamos el Asunto o N° Factura
			descDoc := strings.TrimSpace(doc.Asunto)
			if descDoc == "" {
				descDoc = "Factura N° " + doc.NumeroDocumento
			}
			itemKey := strings.ToUpper(descDoc)

			if item, ok := itemsMap[itemKey]; ok {
				item.CantidadTotal += 1
				item.TotalGastado += doc.Total
				item.CantidadFacturas++
			} else {
				itemsMap[itemKey] = &ItemGastoDTO{
					Descripcion:      descDoc,
					CantidadTotal:    1,
					TotalGastado:     doc.Total,
					ProveedorID:      provID,
					ProveedorNombre:  provRazonSocial,
					CantidadFacturas: 1,
				}
			}
		}

		// Normas de reparto
		for _, nr := range rad.NormasReparto {
			if nr.NormaReparto != nil {
				codigo := strings.TrimSpace(nr.NormaReparto.Codigo)
				if codigo == "" {
					codigo = "S/C"
				}
				montoNorma := (nr.Porcentaje / 100.0) * doc.Total

				if n, ok := normasMap[codigo]; ok {
					n.TotalGastado += montoNorma
					n.CantidadFacturas++
				} else {
					normasMap[codigo] = &NormaGastoDTO{
						Codigo:           codigo,
						Nombre:           nr.NormaReparto.Nombre,
						Sucursal:         nr.NormaReparto.Sucursal,
						Departamento:     nr.NormaReparto.Departamento,
						TotalGastado:     montoNorma,
						CantidadFacturas: 1,
					}
				}
			}
		}

		// Radicado DTO
		resp.Radicados = append(resp.Radicados, RadicadoGastoDTO{
			ID:                   rad.ID,
			NumeroRadicado:       rad.NumeroRadicado,
			FechaRadicacion:      rad.FechaRadicacion,
			EstadoPosesion:       rad.EstadoPosesion,
			DocumentoComercialID: doc.ID,
			NumeroDocumento:      doc.NumeroDocumento,
			FechaDocumento:       doc.FechaDocumento,
			Total:                doc.Total,
			Subtotal:             doc.Subtotal,
			Iva:                  doc.Iva,
			ProveedorID:          provID,
			ProveedorRazonSocial: provRazonSocial,
			ProveedorNIT:         provNIT,
			Detalles:             doc.Detalles,
			NormasReparto:        rad.NormasReparto,
		})
	}

	resp.CantidadProveedores = len(proveedoresMap)
	if resp.CantidadFacturas > 0 {
		resp.PromedioFactura = math.Round((resp.TotalGeneral/float64(resp.CantidadFacturas))*100) / 100
	}

	// Convertir mapas a slices y calcular porcentajes
	for _, p := range proveedoresMap {
		if resp.TotalGeneral > 0 {
			p.PorcentajeGasto = math.Round((p.TotalGastado/resp.TotalGeneral)*10000) / 100
		}
		resp.Proveedores = append(resp.Proveedores, *p)
	}
	sort.Slice(resp.Proveedores, func(i, j int) bool {
		return resp.Proveedores[i].TotalGastado > resp.Proveedores[j].TotalGastado
	})

	for _, item := range itemsMap {
		if resp.TotalGeneral > 0 {
			item.PorcentajeGasto = math.Round((item.TotalGastado/resp.TotalGeneral)*10000) / 100
		}
		resp.Items = append(resp.Items, *item)
	}
	sort.Slice(resp.Items, func(i, j int) bool {
		return resp.Items[i].TotalGastado > resp.Items[j].TotalGastado
	})

	for _, n := range normasMap {
		if resp.TotalGeneral > 0 {
			n.PorcentajeGasto = math.Round((n.TotalGastado/resp.TotalGeneral)*10000) / 100
		}
		resp.Normas = append(resp.Normas, *n)
	}
	sort.Slice(resp.Normas, func(i, j int) bool {
		return resp.Normas[i].TotalGastado > resp.Normas[j].TotalGastado
	})

	return resp, nil
}
