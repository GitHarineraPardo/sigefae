package documento_radicado

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"sigefae/internal/db"
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

func (h *Handler) Create(c *gin.Context) {
	user := c.MustGet("user").(db.Usuario)

	var dto CreateDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.service.Create(dto, user.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// ── NOTIFICAR al primer responsable ──
	if h.notifSvc != nil && response.UsuarioActualID != 0 && response.UsuarioActualID != user.ID {
		docID := response.ID
		h.notifSvc.CreateFromEvent(notificacion.CreateDTO{
			UsuarioID:           response.UsuarioActualID,
			DocumentoRadicadoID: &docID,
			Mensaje:             "Nuevo radicado " + response.NumeroRadicado + " requiere tu revisión",
			Estado:              "Pendiente",
			Tipo:                "Asignacion",
			FechaCreacion:       time.Now(),
		})
	}

	// ── REGISTRAR Trazabilidad ──
	descTrazabilidad := "Documento radicado e iniciado en el flujo."
	if response.UsuarioActualID != 0 {
		var usuarioSiguiente db.Usuario
		h.db.First(&usuarioSiguiente, response.UsuarioActualID)
		nombreSiguiente := usuarioSiguiente.Nombre
		if nombreSiguiente == "" {
			nombreSiguiente = "Usuario Desconocido"
		}
		descTrazabilidad = fmt.Sprintf("Documento radicado. Asignado inicialmente a: %s", nombreSiguiente)
	}

	h.db.Create(&db.Trazabilidad{
		DocumentoRadicadoID: response.ID,
		UsuarioID:           user.ID,
		Accion:              "Radicación",
		Descripcion:         descTrazabilidad,
		Fecha:               time.Now(),
	})

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

func (h *Handler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id inválido"})
		return
	}
	response, err := h.service.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
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

func (h *Handler) VerificarPublico(c *gin.Context) {
	numero := c.Param("numero_radicado")

	radicado, err := h.service.GetByNumeroRadicado(numero)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"valido": false,
			"error":  "El documento no existe en nuestros registros",
		})
		return
	}

	// Preparar respuesta pública (no exponemos todo el modelo)
	response := gin.H{
		"valido":           true,
		"numero_radicado":  radicado.NumeroRadicado,
		"fecha_radicacion": radicado.FechaRadicacion,
		"tipo_radicacion":  radicado.TipoRadicacion.Nombre,
		"estado_posesion":  radicado.EstadoPosesion,
		"documento": gin.H{
			"tipo":          radicado.DocumentoComercial.Tipo,
			"numero":        radicado.DocumentoComercial.NumeroDocumento,
			"fecha_emision": radicado.DocumentoComercial.FechaDocumento,
			"total":         radicado.DocumentoComercial.Total,
			"moneda":        radicado.DocumentoComercial.Moneda.Nombre,
			"cufe":          radicado.DocumentoComercial.Cufe, // ← verifica que este campo exista en tu modelo
		},
		"proveedor": gin.H{
			"razon_social": radicado.DocumentoComercial.Proveedor.RazonSocial,
			"nit":          radicado.DocumentoComercial.Proveedor.NumeroDocumento,
		},
		"receptor": gin.H{
			"nombre": radicado.DocumentoComercial.Receptor.Nombre,
			"nit":    radicado.DocumentoComercial.Receptor.NumeroDocumento,
		},
		"qr_url": radicado.Qr.Url,
	}

	c.JSON(http.StatusOK, response)
}
func (h *Handler) GetNormasReparto(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id inválido"})
		return
	}
	normas, err := h.service.GetNormasReparto(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, normas)
}

func (h *Handler) AsignarNormasReparto(c *gin.Context) {
	user := c.MustGet("user").(db.Usuario)
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id inválido"})
		return
	}
	var dto struct {
		Normas []NormaRepartoInputDTO `json:"normas" binding:"required,dive"`
	}
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.AsignarNormasReparto(uint(id), dto.Normas, user.ID, user.Rol != nil && user.Rol.Nombre == "Superadministrador"); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	for _, n := range dto.Normas {
		if n.Porcentaje == 0 {
			if err := h.db.Where("documento_radicado_id = ? AND norma_reparto_id = ?", id, n.NormaRepartoID).Delete(&db.RadicadoNormaReparto{}).Error; err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "No se pudo eliminar la norma con valor 0"})
				return
			}
		}
	}
	c.JSON(http.StatusOK, gin.H{"message": "normas asignadas correctamente"})
}

func (h *Handler) GetActivosFijos(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id inválido"})
		return
	}
	items, err := h.service.GetActivosFijos(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, items)
}

func (h *Handler) AsignarActivosFijos(c *gin.Context) {
	user := c.MustGet("user").(db.Usuario)
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id inválido"})
		return
	}
	var dto struct {
		ActivosFijos []ActivoFijoInputDTO `json:"activos_fijos"`
	}
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	isAdmin := user.Rol != nil && user.Rol.Nombre == "Superadministrador"
	if err := h.service.AsignarActivosFijos(uint(id), dto.ActivosFijos, user.ID, isAdmin); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "activos fijos guardados correctamente"})
}

// SolicitarRechazo permite a un usuario solicitar que un administrador rechace
// el documento. Se crean notificaciones para todos los administradores.
func (h *Handler) SolicitarRechazo(c *gin.Context) {
	user := c.MustGet("user").(db.Usuario)

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id inválido"})
		return
	}

	var body struct {
		Mensaje string `json:"mensaje"`
	}
	_ = c.ShouldBindJSON(&body)

	// Buscar administradores por rol de forma segura
	var role db.Rol
	if err := h.db.Where("nombre = ?", "Superadministrador").First(&role).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "no se pudo localizar el rol de administrador"})
		return
	}

	// Crear registro de solicitud
	sol := db.SolicitudRechazo{
		DocumentoRadicadoID: uint(id),
		UsuarioID:           user.ID,
		Mensaje:             body.Mensaje,
		Estado:              "Pendiente",
		FechaCreacion:       time.Now(),
	}
	if err := h.db.Create(&sol).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Notificar administradores (como antes)
	var admins []db.Usuario
	if err := h.db.Joins("Rol").Where("Rol.nombre = ?", "Superadministrador").Find(&admins).Error; err != nil {
		// no crítico: continuar
		admins = []db.Usuario{}
	}
	docID := uint(id)
	for _, a := range admins {
		msg := user.Nombre + " solicita rechazo. "
		if body.Mensaje != "" {
			msg = msg + ": " + body.Mensaje
		}
		if h.notifSvc != nil {
			_, _ = h.notifSvc.CreateFromEvent(notificacion.CreateDTO{
				UsuarioID:           a.ID,
				DocumentoRadicadoID: &docID,
				Mensaje:             msg,
				Estado:              "Pendiente",
				Tipo:                "Rechazo",
				FechaCreacion:       time.Now(),
			})
		} else {
			n := db.Notificacion{
				UsuarioID:           a.ID,
				DocumentoRadicadoID: &docID,
				Mensaje:             msg,
				Estado:              "Pendiente",
				Tipo:                "Rechazo",
				FechaCreacion:       time.Now(),
			}
			_ = h.db.Create(&n).Error
		}
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Solicitud creada", "solicitud_id": sol.ID})
}

// ListMySolicitudes devuelve las solicitudes del usuario autenticado
func (h *Handler) ListMySolicitudes(c *gin.Context) {
	user := c.MustGet("user").(db.Usuario)
	var lista []db.SolicitudRechazo
	if err := h.db.Preload("DocumentoRadicado").Where("usuario_id = ?", user.ID).Order("fecha_creacion desc").Find(&lista).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, lista)
}

// ListSolicitudes devuelve las solicitudes pendientes (admin)
func (h *Handler) ListSolicitudes(c *gin.Context) {
	var lista []db.SolicitudRechazo
	if err := h.db.Preload("DocumentoRadicado").Preload("Usuario").Where("estado = ?", "Pendiente").Order("fecha_creacion desc").Find(&lista).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, lista)
}

// DecidirSolicitud permite al admin aceptar o rechazar una solicitud
func (h *Handler) DecidirSolicitud(c *gin.Context) {
	adminUser := c.MustGet("user").(db.Usuario)
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id inválido"})
		return
	}
	var body struct {
		Accept  bool   `json:"accept"`
		Mensaje string `json:"mensaje"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var sol db.SolicitudRechazo
	if err := h.db.First(&sol, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "solicitud no encontrada"})
		return
	}
	if sol.Estado != "Pendiente" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "solicitud ya procesada"})
		return
	}

	now := time.Now()
	estado := "Rechazada"
	if body.Accept {
		estado = "Aceptada"
	}
	sol.Estado = estado
	sol.ResueltoPorID = &adminUser.ID
	sol.Respuesta = body.Mensaje
	sol.FechaResolucion = &now
	if err := h.db.Save(&sol).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Si aceptada, marcar documento como `Rechazado` (estado final)
	if body.Accept {
		var radicado db.DocumentoRadicado
		if err := h.db.First(&radicado, sol.DocumentoRadicadoID).Error; err == nil {
			updates := map[string]any{
				"estado_posesion":   "Rechazado",
				"usuario_actual_id": 0,
				"paso_actual_id":    gorm.Expr("NULL"),
			}
			_ = h.db.Model(&db.DocumentoRadicado{}).Where("id = ?", radicado.ID).Updates(updates).Error
			h.db.Create(&db.Trazabilidad{
				DocumentoRadicadoID: radicado.ID,
				UsuarioID:           adminUser.ID,
				Accion:              "Rechazo",
				Descripcion:         body.Mensaje,
				Fecha:               time.Now(),
			})
		}
	}

	// Notificar al solicitante
	if h.notifSvc != nil {
		_, _ = h.notifSvc.CreateFromEvent(notificacion.CreateDTO{
			UsuarioID:           sol.UsuarioID,
			DocumentoRadicadoID: &sol.DocumentoRadicadoID,
			Mensaje:             "Su solicitud de rechazo fue procesada: " + estado + ". " + body.Mensaje,
			Estado:              "Pendiente",
			Tipo:                "Rechazo",
			FechaCreacion:       time.Now(),
		})
	} else {
		n := db.Notificacion{
			UsuarioID:           sol.UsuarioID,
			DocumentoRadicadoID: &sol.DocumentoRadicadoID,
			Mensaje:             "Su solicitud de rechazo fue procesada: " + estado + ". " + body.Mensaje,
			Estado:              "Pendiente",
			Tipo:                "Rechazo",
			FechaCreacion:       time.Now(),
		}
		_ = h.db.Create(&n).Error
	}

	c.JSON(http.StatusOK, gin.H{"message": "Solicitud procesada", "estado": estado})
}

// Rechazar permite al administrador o contabilidad marcar el documento como rechazado (estado final).
func (h *Handler) Rechazar(c *gin.Context) {
	adminUser := c.MustGet("user").(db.Usuario)

	// Validar permisos (Superadministrador o Contabilidad)
	if adminUser.Rol == nil || (adminUser.Rol.Nombre != "Superadministrador" && adminUser.Rol.Nombre != "Contabilidad") {
		c.JSON(http.StatusForbidden, gin.H{"error": "No tienes permisos para rechazar documentos"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id inválido"})
		return
	}

	var body struct {
		Mensaje      string `json:"mensaje"`
		NotifyUserID uint   `json:"notify_user_id"`
	}
	_ = c.ShouldBindJSON(&body)

	var radicado db.DocumentoRadicado
	if err := h.db.First(&radicado, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "documento no encontrado"})
		return
	}

	// Intentar marcar la tarea activa como "Devuelta" y registrar cierre de la misma
	now := time.Now()
	var estadoEnProceso db.EstadoTarea
	h.db.Where("nombre = ?", "En Proceso").First(&estadoEnProceso)
	var estadoDevuelta db.EstadoTarea
	h.db.Where("nombre = ?", "Devuelta").First(&estadoDevuelta)
	if estadoDevuelta.ID == 0 {
		estadoDevuelta.ID = 4
	}

	var tareaActiva db.Tarea
	if estadoEnProceso.ID != 0 {
		if err := h.db.Where("documento_radicado_id = ? AND estado_id = ?", radicado.ID, estadoEnProceso.ID).First(&tareaActiva).Error; err == nil {
			// Cerrar la tarea activa como Devuelta
			_ = h.db.Model(&tareaActiva).Updates(map[string]any{
				"estado_id":          estadoDevuelta.ID,
				"fecha_finalizacion": now,
			}).Error

			// Registrar comentario de rechazo en la tarea
			h.db.Create(&db.Comentario{
				DocumentoRadicadoID: radicado.ID,
				UsuarioID:           adminUser.ID,
				Descripcion:         "RECHAZO ADMIN: " + body.Mensaje,
				Fecha:               now,
			})
		}
	}
	// Actualizar estado a 'Rechazado' (rechazo final) y dejarlo inactivo
	updates := map[string]any{
		"estado_posesion":            "Rechazado",
		"usuario_actual_id":          0,
		"paso_actual_id":             gorm.Expr("NULL"),
		"tarea_pendiente_retorno_id": gorm.Expr("NULL"),
	}

	if err := h.db.Model(&db.DocumentoRadicado{}).Where("id = ?", radicado.ID).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Registrar trazabilidad
	h.db.Create(&db.Trazabilidad{
		DocumentoRadicadoID: radicado.ID,
		UsuarioID:           adminUser.ID,
		Accion:              "Rechazo",
		Descripcion:         body.Mensaje,
		Fecha:               time.Now(),
	})

	// Notificar al usuario que solicitó (si viene) o al dueño del documento
	if body.NotifyUserID != 0 {
		_, _ = h.notifSvc.CreateFromEvent(notificacion.CreateDTO{
			UsuarioID:           body.NotifyUserID,
			DocumentoRadicadoID: &radicado.ID,
			Mensaje:             "Su solicitud ha sido procesada: " + body.Mensaje,
			Estado:              "Pendiente",
			Tipo:                "Finalizado",
			FechaCreacion:       time.Now(),
		})
	}

	c.JSON(http.StatusOK, gin.H{"message": "Documento marcado como rechazado"})
}

// MarcarCompletado permite al administrador marcar el documento como completado (estado final).
func (h *Handler) MarcarCompletado(c *gin.Context) {
	adminUser := c.MustGet("user").(db.Usuario)

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id inválido"})
		return
	}

	var body struct {
		Mensaje      string `json:"mensaje"`
		NotifyUserID uint   `json:"notify_user_id"`
	}
	_ = c.ShouldBindJSON(&body)

	var radicado db.DocumentoRadicado
	if err := h.db.First(&radicado, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "documento no encontrado"})
		return
	}

	updates := map[string]any{
		"estado_posesion":            "Completado",
		"usuario_actual_id":          0,
		"paso_actual_id":             gorm.Expr("NULL"),
		"tarea_pendiente_retorno_id": gorm.Expr("NULL"),
	}

	if err := h.db.Model(&db.DocumentoRadicado{}).Where("id = ?", radicado.ID).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// ── AUTO-APRENDIZAJE DE NORMAS DE REPARTO ──
	_ = h.service.MemorizarNormasProveedorRuta(radicado.ID)

	// Registrar trazabilidad
	h.db.Create(&db.Trazabilidad{
		DocumentoRadicadoID: radicado.ID,
		UsuarioID:           adminUser.ID,
		Accion:              "Completado",
		Descripcion:         body.Mensaje,
		Fecha:               time.Now(),
	})

	// Notificar al usuario (opcional)
	if body.NotifyUserID != 0 && h.notifSvc != nil {
		_, _ = h.notifSvc.CreateFromEvent(notificacion.CreateDTO{
			UsuarioID:           body.NotifyUserID,
			DocumentoRadicadoID: &radicado.ID,
			Mensaje:             "Su radicado fue marcado como completado: " + body.Mensaje,
			Estado:              "Pendiente",
			Tipo:                "Finalizado",
			FechaCreacion:       time.Now(),
		})
	}

	// Notificar a administradores que el radicado fue completado
	if h.notifSvc != nil {
		var admins []db.Usuario
		if err := h.db.Joins("Rol").Where("Rol.nombre = ?", "Superadministrador").Find(&admins).Error; err == nil {
			docID := radicado.ID
			for _, a := range admins {
				_, _ = h.notifSvc.CreateFromEvent(notificacion.CreateDTO{
					UsuarioID:           a.ID,
					DocumentoRadicadoID: &docID,
					Mensaje:             "El radicado " + radicado.NumeroRadicado + " fue marcado como Completado.",
					Estado:              "Pendiente",
					Tipo:                "Finalizado",
					FechaCreacion:       time.Now(),
				})
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Documento marcado como completado"})
}

// ─────────────────────────────────────────────────────────────────
// Solicitudes de Cambio de Norma de Reparto
// ─────────────────────────────────────────────────────────────────

func (h *Handler) SolicitarCambioNormaReparto(c *gin.Context) {
	c.JSON(http.StatusBadRequest, gin.H{"error": "La solicitud de cambio para normas de reparto está deshabilitada."})
}

func (h *Handler) ListMisSolicitudesCambioNorma(c *gin.Context) {
	user := c.MustGet("user").(db.Usuario)
	var lista []db.SolicitudCambioNormaReparto
	if err := h.db.Preload("DocumentoRadicado").Preload("NormaReparto").Where("usuario_id = ?", user.ID).Order("fecha_creacion desc").Find(&lista).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, lista)
}

func (h *Handler) ListSolicitudesCambioNorma(c *gin.Context) {
	var lista []db.SolicitudCambioNormaReparto
	if err := h.db.Preload("DocumentoRadicado").Preload("NormaReparto").Preload("Usuario").Where("estado = ?", "Pendiente").Order("fecha_creacion desc").Find(&lista).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, lista)
}

func (h *Handler) DecidirSolicitudCambioNorma(c *gin.Context) {
	adminUser := c.MustGet("user").(db.Usuario)
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id inválido"})
		return
	}

	var body struct {
		Accept  bool   `json:"accept"`
		Mensaje string `json:"mensaje"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var sol db.SolicitudCambioNormaReparto
	if err := h.db.Preload("NormaReparto").First(&sol, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "solicitud no encontrada"})
		return
	}
	if sol.Estado != "Pendiente" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "solicitud ya procesada"})
		return
	}

	now := time.Now()
	estado := "Rechazada"
	if body.Accept {
		estado = "Aprobada"
	}
	sol.Estado = estado
	sol.ResueltoPorID = &adminUser.ID
	sol.Respuesta = body.Mensaje
	sol.FechaResolucion = &now

	if err := h.db.Save(&sol).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if body.Accept {
		var rnr db.RadicadoNormaReparto
		if err := h.db.First(&rnr, sol.RadicadoNormaRepartoID).Error; err == nil {
			rnr.Porcentaje = sol.NuevoPorcentaje
			h.db.Save(&rnr)
			if rnr.Porcentaje == 0 {
				h.db.Delete(&rnr)
			}
		}

		h.db.Create(&db.Trazabilidad{
			DocumentoRadicadoID: sol.DocumentoRadicadoID,
			UsuarioID:           adminUser.ID,
			Accion:              "Cambio Norma Aprobado",
			Descripcion:         fmt.Sprintf("Cambio de porcentaje aprobado para la norma %s (%.2f%% -> %.2f%%). %s", sol.NormaReparto.Nombre, sol.PorcentajeAnterior, sol.NuevoPorcentaje, body.Mensaje),
			Fecha:               time.Now(),
		})
	} else {
		h.db.Create(&db.Trazabilidad{
			DocumentoRadicadoID: sol.DocumentoRadicadoID,
			UsuarioID:           adminUser.ID,
			Accion:              "Cambio Norma Rechazado",
			Descripcion:         fmt.Sprintf("Solicitud de cambio de norma rechazada. %s", body.Mensaje),
			Fecha:               time.Now(),
		})
	}

	if h.notifSvc != nil {
		_, _ = h.notifSvc.CreateFromEvent(notificacion.CreateDTO{
			UsuarioID:           sol.UsuarioID,
			DocumentoRadicadoID: &sol.DocumentoRadicadoID,
			Mensaje:             fmt.Sprintf("Su solicitud de cambio de norma fue %s. %s", estado, body.Mensaje),
			Estado:              "Pendiente",
			Tipo:                "Sistema",
			FechaCreacion:       time.Now(),
		})
	}

	c.JSON(http.StatusOK, gin.H{"message": "Solicitud procesada", "estado": estado})
}

func (h *Handler) CrearSolicitudPermiso(c *gin.Context) {
	user := c.MustGet("user").(db.Usuario)

	var body struct {
		Tipo        string `json:"tipo" binding:"required"`
		ObjetoID    uint   `json:"objeto_id" binding:"required"`
		Accion      string `json:"accion" binding:"required"`
		Propietario uint   `json:"propietario_id" binding:"required"`
		Mensaje     string `json:"mensaje"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if body.Tipo == "norma_reparto" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "La solicitud de permiso para normas de reparto está deshabilitada."})
		return
	}
	if body.Tipo != "archivo" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tipo inválido"})
		return
	}
	if body.Accion != "editar" && body.Accion != "eliminar" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "acción inválida"})
		return
	}

	solicitud := db.SolicitudPermiso{
		Tipo:          body.Tipo,
		ObjetoID:      body.ObjetoID,
		Accion:        body.Accion,
		SolicitanteID: user.ID,
		PropietarioID: body.Propietario,
		Mensaje:       body.Mensaje,
		Estado:        "Pendiente",
		FechaCreacion: time.Now(),
	}
	if err := h.db.Create(&solicitud).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if h.notifSvc != nil {
		_, _ = h.notifSvc.CreateFromEvent(notificacion.CreateDTO{
			UsuarioID:     body.Propietario,
			Mensaje:       fmt.Sprintf("%s solicita permiso para %s un %s. %s", user.Nombre, body.Accion, body.Tipo, body.Mensaje),
			Estado:        "Pendiente",
			Tipo:          "Sistema",
			FechaCreacion: time.Now(),
		})
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Solicitud enviada", "solicitud_id": solicitud.ID})
}

func (h *Handler) ListMisSolicitudesPermiso(c *gin.Context) {
	user := c.MustGet("user").(db.Usuario)
	var solicitudes []db.SolicitudPermiso
	if err := h.db.Preload("Solicitante").Preload("Propietario").Where("solicitante_id = ?", user.ID).Order("fecha_creacion desc").Find(&solicitudes).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, solicitudes)
}

func (h *Handler) ListSolicitudesPermiso(c *gin.Context) {
	var solicitudes []db.SolicitudPermiso
	if err := h.db.Preload("Solicitante").Preload("Propietario").Where("estado = ?", "Pendiente").Order("fecha_creacion desc").Find(&solicitudes).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, solicitudes)
}

func (h *Handler) DecidirSolicitudPermiso(c *gin.Context) {
	user := c.MustGet("user").(db.Usuario)
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id inválido"})
		return
	}

	var sol db.SolicitudPermiso
	if err := h.db.First(&sol, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "solicitud no encontrada"})
		return
	}
	if sol.PropietarioID != user.ID && user.Rol != nil && user.Rol.Nombre != "Superadministrador" {
		c.JSON(http.StatusForbidden, gin.H{"error": "solo el propietario o el administrador puede decidir"})
		return
	}
	if sol.Estado != "Pendiente" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "solicitud ya procesada"})
		return
	}

	var body struct {
		Accept  bool   `json:"accept"`
		Mensaje string `json:"mensaje"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	now := time.Now()
	sol.Estado = "Rechazada"
	if body.Accept {
		sol.Estado = "Aprobada"
	}
	sol.ResueltoPorID = &user.ID
	sol.Respuesta = body.Mensaje
	sol.FechaResolucion = &now
	if err := h.db.Save(&sol).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if h.notifSvc != nil {
		_, _ = h.notifSvc.CreateFromEvent(notificacion.CreateDTO{
			UsuarioID:     sol.SolicitanteID,
			Mensaje:       fmt.Sprintf("Tu solicitud de permiso para %s fue %s. %s", sol.Tipo, sol.Estado, body.Mensaje),
			Estado:        "Pendiente",
			Tipo:          "Sistema",
			FechaCreacion: time.Now(),
		})
	}

	c.JSON(http.StatusOK, gin.H{"message": "Solicitd de permiso procesada", "estado": sol.Estado})
}

// ─────────────────────────────────────────────────────────────────
// Causación (Contabilidad)
// ─────────────────────────────────────────────────────────────────

func (h *Handler) Causar(c *gin.Context) {
	user := c.MustGet("user").(db.Usuario)
	
	// Validar permisos (Contabilidad o Superadministrador)
	if user.Rol == nil || (user.Rol.Nombre != "Contabilidad" && user.Rol.Nombre != "Superadministrador") {
		c.JSON(http.StatusForbidden, gin.H{"error": "No tienes permisos para realizar causaciones"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id inválido"})
		return
	}

	var body struct {
		Causado      bool   `json:"causado"`
		NumeroEgreso string `json:"numero_egreso"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var radicado db.DocumentoRadicado
	if err := h.db.First(&radicado, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "documento no encontrado"})
		return
	}

	// Regla de negocio: Si ya está causado, no se puede desmarcar a menos que sea Superadministrador
	if radicado.Causado && !body.Causado && user.Rol.Nombre != "Superadministrador" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Un radicado causado solo puede ser revertido por un Superadministrador"})
		return
	}

	// Regla de negocio: Si ya tiene número de egreso, no se puede modificar a menos que sea Superadministrador
	if radicado.NumeroEgreso != "" && body.NumeroEgreso != radicado.NumeroEgreso && user.Rol.Nombre != "Superadministrador" {
		c.JSON(http.StatusForbidden, gin.H{"error": "El número de egreso ya fue asignado y solo puede ser modificado por un Superadministrador"})
		return
	}

	updates := map[string]any{
		"causado":       body.Causado,
		"numero_egreso": body.NumeroEgreso,
	}

	if body.Causado && !radicado.Causado {
		now := time.Now()
		updates["fecha_causacion"] = &now
	} else if !body.Causado {
		updates["fecha_causacion"] = gorm.Expr("NULL")
	}

	if err := h.db.Model(&radicado).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Registrar trazabilidad
	accion := "Causación Realizada"
	desc := "El documento fue causado."
	if !body.Causado {
		accion = "Causación Revertida"
		desc = "El documento dejó de estar causado."
	}
	if body.NumeroEgreso != "" {
		desc += fmt.Sprintf(" Número de egreso: %s", body.NumeroEgreso)
	}

	h.db.Create(&db.Trazabilidad{
		DocumentoRadicadoID: radicado.ID,
		UsuarioID:           user.ID,
		Accion:              accion,
		Descripcion:         desc,
		Fecha:               time.Now(),
	})

	c.JSON(http.StatusOK, gin.H{"message": "Causación actualizada correctamente"})
}

// ─────────────────────────────────────────────────────────────────
// Pago (Tesorería)
// ─────────────────────────────────────────────────────────────────

func (h *Handler) Pagar(c *gin.Context) {
	user := c.MustGet("user").(db.Usuario)
	
	// Validar permisos (Tesorería o Superadministrador)
	if user.Rol == nil || (user.Rol.Nombre != "Tesorería" && user.Rol.Nombre != "Superadministrador") {
		c.JSON(http.StatusForbidden, gin.H{"error": "No tienes permisos para realizar pagos"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id inválido"})
		return
	}

	var body struct {
		Pagado bool `json:"pagado"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var radicado db.DocumentoRadicado
	if err := h.db.First(&radicado, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "documento no encontrado"})
		return
	}

	// Regla de negocio: Si ya está pagado, no se puede desmarcar a menos que sea Superadministrador
	if radicado.Pagado && !body.Pagado && user.Rol.Nombre != "Superadministrador" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Una factura pagada solo puede ser revertida por un Superadministrador"})
		return
	}

	updates := map[string]any{
		"pagado": body.Pagado,
	}

	var accionTrazabilidad string
	if body.Pagado && !radicado.Pagado {
		now := time.Now()
		updates["fecha_pago"] = &now
		accionTrazabilidad = "Pago Realizado"
	} else if !body.Pagado && radicado.Pagado {
		updates["fecha_pago"] = gorm.Expr("NULL")
		accionTrazabilidad = "Pago Revertido"
	}

	if err := h.db.Model(&radicado).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Trazabilidad si hubo cambio de estado
	if accionTrazabilidad != "" {
		desc := "La factura fue marcada como pagada."
		if !body.Pagado {
			desc = "Se revertió el pago de la factura."
		}
		h.db.Create(&db.Trazabilidad{
			DocumentoRadicadoID: radicado.ID,
			UsuarioID:           user.ID,
			Accion:              accionTrazabilidad,
			Descripcion:         desc,
			Fecha:               time.Now(),
		})
	}

	c.JSON(http.StatusOK, gin.H{"message": "Pago actualizado correctamente"})
}

// ─────────────────────────────────────────────────────────────────
// Comprobantes Subidos (Tesorería)
// ─────────────────────────────────────────────────────────────────

func (h *Handler) ToggleComprobantesSubidos(c *gin.Context) {
	user := c.MustGet("user").(db.Usuario)
	
	// Validar permisos (Tesorería o Superadministrador)
	if user.Rol == nil || (user.Rol.Nombre != "Tesorería" && user.Rol.Nombre != "Superadministrador") {
		c.JSON(http.StatusForbidden, gin.H{"error": "No tienes permisos para modificar este estado"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id inválido"})
		return
	}

	var body struct {
		ComprobantesSubidos bool `json:"comprobantes_subidos"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var radicado db.DocumentoRadicado
	if err := h.db.First(&radicado, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "documento no encontrado"})
		return
	}

	// Regla de negocio: Si ya está marcado como subido, solo el superadmin puede revertirlo
	if radicado.ComprobantesSubidos && !body.ComprobantesSubidos && user.Rol.Nombre != "Superadministrador" {
		c.JSON(http.StatusForbidden, gin.H{"error": "El bloqueo de comprobantes solo puede ser revertido por un Superadministrador"})
		return
	}

	if err := h.db.Model(&radicado).Update("comprobantes_subidos", body.ComprobantesSubidos).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Trazabilidad
	if radicado.ComprobantesSubidos != body.ComprobantesSubidos {
		accion := "Comprobantes Bloqueados"
		desc := "Se confirmó la subida de los comprobantes de pago y se bloqueó la carga de nuevos archivos."
		if !body.ComprobantesSubidos {
			accion = "Comprobantes Desbloqueados"
			desc = "Se habilitó nuevamente la carga de comprobantes de pago."
		}
		h.db.Create(&db.Trazabilidad{
			DocumentoRadicadoID: radicado.ID,
			UsuarioID:           user.ID,
			Accion:              accion,
			Descripcion:         desc,
			Fecha:               time.Now(),
		})
	}

	c.JSON(http.StatusOK, gin.H{"message": "Estado de comprobantes actualizado", "comprobantes_subidos": body.ComprobantesSubidos})
}
