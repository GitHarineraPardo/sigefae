package tarea

import (
	"fmt"
	"math" // ← AGREGAR
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"sigefae/internal/db"
	"sigefae/internal/documento_radicado"
	"sigefae/internal/notificacion"
)

type Handler struct {
	service  *Service
	notifSvc *notificacion.Service
	db       *gorm.DB
}

func NewHandler(service *Service, notifSvc *notificacion.Service, database *gorm.DB) *Handler {
	return &Handler{
		service:  service,
		notifSvc: notifSvc,
		db:       database,
	}
}

func (h *Handler) ListByRadicado(c *gin.Context) {
	id := c.Param("id")
	radicadoID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id inválido"})
		return
	}

	var tareas []db.Tarea
	if err := h.db.Where("documento_radicado_id = ?", radicadoID).
		Preload("Estado").
		Preload("UsuarioAsignado").
		Order("id asc").
		Find(&tareas).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tareas)
}
func (h *Handler) Completar(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id inválido"})
		return
	}

	user := c.MustGet("user").(db.Usuario)

	var tarea db.Tarea
	if err := h.db.Preload("DocumentoRadicado").First(&tarea, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tarea no encontrada"})
		return
	}

	if tarea.UsuarioAsignadoID != user.ID && user.Rol.Nombre != "Superadministrador" && user.Rol.Nombre != "Contabilidad" {
		c.JSON(http.StatusForbidden, gin.H{"error": "No estás asignado a esta tarea"})
		return
	}

	now := time.Now()

	var estadoCompletado db.EstadoTarea
	h.db.Where("nombre = ?", "Completada").First(&estadoCompletado)
	if estadoCompletado.ID == 0 {
		estadoCompletado.ID = 3
	}

	if tarea.EstadoID == estadoCompletado.ID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "La tarea ya está completada"})
		return
	}

	// ═══════════════════════════════════════════════════════════════
	// 1. DETERMINAR QUÉ VA A PASAR DESPUÉS (sin modificar nada aún)
	// ═══════════════════════════════════════════════════════════════

	var siguiente db.Tarea
	var haySiguiente bool
	var haySaltoDirecto bool

	// 1.1 Verificar si hay retorno directo pendiente
	var tareaRetornoID *uint
	if tarea.DocumentoRadicado != nil {
		tareaRetornoID = tarea.DocumentoRadicado.TareaPendienteRetornoID
	}
	if tareaRetornoID != nil && *tareaRetornoID != 0 {
		if err := h.db.First(&siguiente, *tareaRetornoID).Error; err == nil {
			haySaltoDirecto = true
		}
	}

	// 1.2 Si no hay salto directo, buscar siguiente tarea en el flujo
	if !haySaltoDirecto {
		if err := h.db.Where("documento_radicado_id = ? AND id > ? AND estado_id != ?",
			tarea.DocumentoRadicadoID, tarea.ID, estadoCompletado.ID).
			Order("id asc").First(&siguiente).Error; err == nil {
			haySiguiente = true
		}
	}

	// ═══════════════════════════════════════════════════════════════
	// 2. VALIDAR NORMAS DE REPARTO Y ACTIVOS FIJOS SOLO SI ES LA ÚLTIMA TAREA
	// ═══════════════════════════════════════════════════════════════
	if !haySaltoDirecto && !haySiguiente {
		var totalPorcentajeNormas float64
		h.db.Model(&db.RadicadoNormaReparto{}).
			Where("documento_radicado_id = ?", tarea.DocumentoRadicadoID).
			Select("COALESCE(SUM(porcentaje),0)").
			Scan(&totalPorcentajeNormas)

		var totalPorcentajeActivos float64
		h.db.Model(&db.RadicadoActivoFijo{}).
			Where("documento_radicado_id = ?", tarea.DocumentoRadicadoID).
			Select("COALESCE(SUM(porcentaje),0)").
			Scan(&totalPorcentajeActivos)

		totalPorcentaje := totalPorcentajeNormas + totalPorcentajeActivos

		if math.Abs(totalPorcentaje-100.0) > 0.01 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("Las normas de reparto y activos fijos suman %.2f%%. Deben sumar exactamente 100%% antes de finalizar el proceso.", totalPorcentaje),
			})
			return
		}
	}

	// ═══════════════════════════════════════════════════════════════
	// 3. AHORA SÍ: marcar tarea como completada
	// ═══════════════════════════════════════════════════════════════
	if err := h.db.Model(&tarea).Updates(map[string]any{
		"estado_id":          estadoCompletado.ID,
		"fecha_finalizacion": now,
	}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error actualizando tarea"})
		return
	}

	// ═══════════════════════════════════════════════════════════════
	// 4. EJECUTAR LA TRANSICIÓN (salto, siguiente o finalizar)
	// ═══════════════════════════════════════════════════════════════
	var notificarA uint = 0
	var numRadicado string = ""
	if tarea.DocumentoRadicado != nil {
		numRadicado = tarea.DocumentoRadicado.NumeroRadicado
	}

	if haySaltoDirecto {
		var estadoEnProceso db.EstadoTarea
		h.db.Where("nombre = ?", "En Proceso").First(&estadoEnProceso)
		if estadoEnProceso.ID == 0 {
			estadoEnProceso.ID = 2
		}

		h.db.Model(&siguiente).Updates(map[string]any{
			"estado_id":          estadoEnProceso.ID,
			"fecha_inicio":       now,
			"fecha_finalizacion": gorm.Expr("NULL"),
		})

		h.db.Model(&db.DocumentoRadicado{}).Where("id = ?", tarea.DocumentoRadicadoID).Updates(map[string]any{
			"usuario_actual_id":          siguiente.UsuarioAsignadoID,
			"estado_posesion":            "EnProceso",
			"tarea_pendiente_retorno_id": gorm.Expr("NULL"),
		})
		notificarA = siguiente.UsuarioAsignadoID

	} else if haySiguiente {
		var estadoEnProceso db.EstadoTarea
		h.db.Where("nombre = ?", "En Proceso").First(&estadoEnProceso)
		if estadoEnProceso.ID == 0 {
			estadoEnProceso.ID = 2
		}

		h.db.Model(&siguiente).Updates(map[string]any{
			"estado_id":    estadoEnProceso.ID,
			"fecha_inicio": now,
		})

		h.db.Model(&db.DocumentoRadicado{}).Where("id = ?", tarea.DocumentoRadicadoID).Updates(map[string]any{
			"usuario_actual_id": siguiente.UsuarioAsignadoID,
			"estado_posesion":   "EnProceso",
		})
		notificarA = siguiente.UsuarioAsignadoID

	} else {
		// Última tarea → finalizar radicado
		h.db.Model(&db.DocumentoRadicado{}).Where("id = ?", tarea.DocumentoRadicadoID).Updates(map[string]any{
			"estado_posesion": "Completado",
		})
		
		// ── AUTO-APRENDIZAJE DE NORMAS DE REPARTO ──
		docSvc := documento_radicado.New(h.db)
		_ = docSvc.MemorizarNormasProveedorRuta(tarea.DocumentoRadicadoID)
	}

	// ═══════════════════════════════════════════════════════════════
	// 5. NOTIFICAR AL SIGUIENTE USUARIO
	// ═══════════════════════════════════════════════════════════════
	if h.notifSvc != nil && notificarA != 0 && notificarA != user.ID {
		docID := tarea.DocumentoRadicadoID
		h.notifSvc.CreateFromEvent(notificacion.CreateDTO{
			UsuarioID:           notificarA,
			DocumentoRadicadoID: &docID,
			Mensaje:             "Te asignaron el radicado " + numRadicado,
			Estado:              "Pendiente",
			Tipo:                "Asignacion",
			FechaCreacion:       time.Now(),
		})

		// Además notificar a los administradores para control
		var admins []db.Usuario
		if err := h.db.Joins("Rol").Where("Rol.nombre = ?", "Superadministrador").Find(&admins).Error; err == nil {
			
			var usuarioSiguiente db.Usuario
			h.db.First(&usuarioSiguiente, notificarA)
			nombreSiguiente := usuarioSiguiente.Nombre
			if nombreSiguiente == "" {
				nombreSiguiente = "Usuario Desconocido"
			}

			for _, a := range admins {
				// evitar notificar al responsable ya notificado
				if a.ID == uint(user.ID) || a.ID == notificarA {
					continue
				}
				copyDocID := tarea.DocumentoRadicadoID
				_, _ = h.notifSvc.CreateFromEvent(notificacion.CreateDTO{
					UsuarioID:           a.ID,
					DocumentoRadicadoID: &copyDocID,
					Mensaje:             fmt.Sprintf("El radicado %s fue reasignado a %s", numRadicado, nombreSiguiente),
					Estado:              "Pendiente",
					Tipo:                "Asignacion",
					FechaCreacion:       time.Now(),
				})
			}
		}
	} else if h.notifSvc != nil && notificarA == 0 {
		var admins []db.Usuario
		if err := h.db.Joins("Rol").Where("Rol.nombre = ?", "Superadministrador").Find(&admins).Error; err == nil {
			for _, a := range admins {
				if a.ID == uint(user.ID) {
					continue
				}
				copyDocID := tarea.DocumentoRadicadoID
				_, _ = h.notifSvc.CreateFromEvent(notificacion.CreateDTO{
					UsuarioID:           a.ID,
					DocumentoRadicadoID: &copyDocID,
					Mensaje:             fmt.Sprintf("El radicado %s ha finalizado su flujo automáticamente.", numRadicado),
					Estado:              "Pendiente",
					Tipo:                "Finalizado",
					FechaCreacion:       time.Now(),
				})
			}
		}
	}

	// ═══════════════════════════════════════════════════════════════
	// 6. REGISTRAR TRAZABILIDAD
	// ═══════════════════════════════════════════════════════════════
	var accionTrazabilidad string
	var descTrazabilidad string

	if notificarA != 0 {
		var usuarioSiguiente db.Usuario
		h.db.First(&usuarioSiguiente, notificarA)
		nombreSiguiente := usuarioSiguiente.Nombre
		if nombreSiguiente == "" {
			nombreSiguiente = "Usuario Desconocido"
		}
		accionTrazabilidad = "Paso Completado y Asignado"
		descTrazabilidad = fmt.Sprintf("El usuario completó su tarea. El documento fue asignado a: %s", nombreSiguiente)
	} else {
		accionTrazabilidad = "Proceso Finalizado"
		descTrazabilidad = "El usuario completó la última tarea. El proceso ha finalizado."
	}

	h.db.Create(&db.Trazabilidad{
		DocumentoRadicadoID: tarea.DocumentoRadicadoID,
		UsuarioID:           user.ID,
		Accion:              accionTrazabilidad,
		Descripcion:         descTrazabilidad,
		Fecha:               time.Now(),
	})

	c.JSON(http.StatusOK, gin.H{"message": "Tarea completada"})
}
func (h *Handler) Devolver(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id inválido"})
		return
	}

	var dto DevolverDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user := c.MustGet("user").(db.Usuario)

	var tarea db.Tarea
	if err := h.db.Preload("DocumentoRadicado").First(&tarea, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tarea actual no encontrada"})
		return
	}

	if tarea.UsuarioAsignadoID != user.ID && user.Rol.Nombre != "Superadministrador" && user.Rol.Nombre != "Contabilidad" {
		c.JSON(http.StatusForbidden, gin.H{"error": "No estás asignado a esta tarea"})
		return
	}

	var destino db.Tarea
	if err := h.db.First(&destino, dto.TareaDestinoID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tarea destino no encontrada"})
		return
	}

	// ===== VALIDACIONES QUE FALTAN =====
	// 1. Mismo documento
	if destino.DocumentoRadicadoID != tarea.DocumentoRadicadoID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "La tarea destino no pertenece a este documento"})
		return
	}

	// 2. No devolver a sí misma
	if destino.ID == tarea.ID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No puede devolver a la misma tarea"})
		return
	}

	// 3. Solo devolver a tareas COMPLETADAS (no a devueltas ni pendientes)
	var estadoDestino db.EstadoTarea
	if err := h.db.First(&estadoDestino, destino.EstadoID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error verificando estado de tarea destino"})
		return
	}
	if estadoDestino.Nombre != "Completada" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Solo puede devolver a una tarea que haya sido completada"})
		return
	}
	// ===================================

	// Estados
	var estadoDevuelta db.EstadoTarea
	h.db.Where("nombre = ?", "Devuelta").First(&estadoDevuelta)
	if estadoDevuelta.ID == 0 {
		estadoDevuelta.ID = 4
	}

	var estadoEnProceso db.EstadoTarea
	h.db.Where("nombre = ?", "En Proceso").First(&estadoEnProceso)
	if estadoEnProceso.ID == 0 {
		estadoEnProceso.ID = 2
	}

	now := time.Now()

	// Actualizar tarea actual a Devuelta
	if err := h.db.Model(&tarea).Updates(map[string]any{
		"estado_id":          estadoDevuelta.ID,
		"fecha_finalizacion": now,
	}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error actualizando tarea actual: " + err.Error()})
		return
	}

	// Reactivar tarea destino (limpiar fecha_finalizacion para que se muestre como activa)
	if err := h.db.Model(&destino).Updates(map[string]any{
		"estado_id":          estadoEnProceso.ID,
		"fecha_inicio":       now,
		"fecha_finalizacion": gorm.Expr("NULL"),
	}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error reactivando tarea destino: " + err.Error()})
		return
	}

	// Actualizar Radicado: mantener en proceso y cambiar responsable al destino
	radicadoUpdates := map[string]any{
		"usuario_actual_id": destino.UsuarioAsignadoID,
		"estado_posesion":   "EnProceso",
	}
	if dto.RetornoDirecto {
		radicadoUpdates["tarea_pendiente_retorno_id"] = tarea.ID
	} else {
		radicadoUpdates["tarea_pendiente_retorno_id"] = gorm.Expr("NULL")
	}

	if err := h.db.Model(&db.DocumentoRadicado{}).Where("id = ?", tarea.DocumentoRadicadoID).Updates(radicadoUpdates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error actualizando radicado: " + err.Error()})
		return
	}

	// Registrar Aprobacion, Comentario y Trazabilidad
	h.db.Create(&db.RegistroAprobacion{
		DocumentoRadicadoID: tarea.DocumentoRadicadoID,
		ResponsableID:       user.ID,
		RolID:               user.RolID,
		Estado:              "DEVUELTO",
		Observacion:         dto.Observacion,
		Fecha:               now,
	})

	h.db.Create(&db.Comentario{
		DocumentoRadicadoID: tarea.DocumentoRadicadoID,
		UsuarioID:           user.ID,
		Descripcion:         "DEVOLUCIÓN: " + dto.Observacion,
		Fecha:               now,
	})

	var usuarioDestino db.Usuario
	h.db.First(&usuarioDestino, destino.UsuarioAsignadoID)
	nombreDestino := usuarioDestino.Nombre
	if nombreDestino == "" {
		nombreDestino = "Usuario Desconocido"
	}

	h.db.Create(&db.Trazabilidad{
		DocumentoRadicadoID: tarea.DocumentoRadicadoID,
		UsuarioID:           user.ID,
		Accion:              "Tarea Devuelta",
		Descripcion:         fmt.Sprintf("El usuario devolvió el documento a %s. Motivo: %s", nombreDestino, dto.Observacion),
		Fecha:               now,
	})

	// Notificar al usuario destino
	if h.notifSvc != nil {
		docID := tarea.DocumentoRadicadoID
		numRad := ""
		if tarea.DocumentoRadicado != nil {
			numRad = tarea.DocumentoRadicado.NumeroRadicado
		}
		_, _ = h.notifSvc.CreateFromEvent(notificacion.CreateDTO{
			UsuarioID:           destino.UsuarioAsignadoID,
			DocumentoRadicadoID: &docID,
			Mensaje:             fmt.Sprintf("Te devolvieron el radicado %s para revisión. Motivo: %s", numRad, dto.Observacion),
			Estado:              "Pendiente",
			Tipo:                "Devolucion",
			FechaCreacion:       now,
		})

		// Notificar también a administradores
		var admins []db.Usuario
		if err := h.db.Joins("Rol").Where("Rol.nombre = ?", "Superadministrador").Find(&admins).Error; err == nil {
			for _, a := range admins {
				// evitar notificar al usuario destino (ya notificado)
				if a.ID == destino.UsuarioAsignadoID {
					continue
				}
				copyDocID := tarea.DocumentoRadicadoID
				_, _ = h.notifSvc.CreateFromEvent(notificacion.CreateDTO{
					UsuarioID:           a.ID,
					DocumentoRadicadoID: &copyDocID,
					Mensaje:             fmt.Sprintf("El radicado %s fue devuelto por %s a %s. Motivo: %s", numRad, user.Nombre, nombreDestino, dto.Observacion),
					Estado:              "Pendiente",
					Tipo:                "Devolucion",
					FechaCreacion:       now,
				})
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Tarea devuelta correctamente"})
}

func (h *Handler) Create(c *gin.Context) {
	var dto CreateDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	response, err := h.service.Create(dto)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, response)
}

func (h *Handler) List(c *gin.Context) {
	response, err := h.service.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, response)
}

func (h *Handler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id inválido"})
		return
	}
	var dto UpdateDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	response, err := h.service.Update(uint(id), dto)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, response)
}
