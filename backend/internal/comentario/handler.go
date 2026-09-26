package comentario

import (
	"net/http"
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

	// ── NOTIFICAR al responsable del radicado ──
	if h.notifSvc != nil && dto.DocumentoRadicadoID != 0 && dto.UsuarioID != 0 {
		var rad db.DocumentoRadicado
		if err := h.db.First(&rad, dto.DocumentoRadicadoID).Error; err == nil {
			if rad.UsuarioActualID != nil && *rad.UsuarioActualID != dto.UsuarioID {
				docID := dto.DocumentoRadicadoID

				mensajeNotif := "Nuevo comentario en " + rad.NumeroRadicado
				if len(dto.Descripcion) > 0 {
					trunc := dto.Descripcion
					if len(trunc) > 40 {
						trunc = trunc[:40] + "..."
					}
					mensajeNotif = "Nuevo comentario en #" + rad.NumeroRadicado + ": \"" + trunc + "\""
				}

				h.notifSvc.CreateFromEvent(notificacion.CreateDTO{
					UsuarioID:           *rad.UsuarioActualID,
					DocumentoRadicadoID: &docID,
					Mensaje:             mensajeNotif,
					Estado:              "Pendiente",
					Tipo:                "Sistema",
					FechaCreacion:       time.Now(),
				})
			}
		}
	}

	// ── REGISTRAR Trazabilidad ──
	h.db.Create(&db.Trazabilidad{
		DocumentoRadicadoID: dto.DocumentoRadicadoID,
		UsuarioID:           dto.UsuarioID,
		Accion:              "Comentario Agregado",
		Descripcion:         "El usuario agregó un nuevo comentario: \"" + dto.Descripcion + "\"",
		Fecha:               time.Now(),
	})

	c.JSON(http.StatusCreated, response)
}

func (h *Handler) List(c *gin.Context) {
	documentoRadicadoID := c.Query("documento_radicado_id")
	response, err := h.service.List(documentoRadicadoID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, response)
}
