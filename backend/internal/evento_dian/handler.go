package evento_dian

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"sigefae/internal/db"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// List devuelve la lista histórica de eventos DIAN para un radicado
func (h *Handler) List(c *gin.Context) {
	idParam := c.Param("id")
	radicadoID, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de radicado inválido"})
		return
	}

	eventos, err := h.service.ListByRadicadoID(uint(radicadoID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, eventos)
}

// Emitir valida y envía un evento electrónico DIAN (Acuse, Recibo, Aceptación o Reclamo)
func (h *Handler) Emitir(c *gin.Context) {
	idParam := c.Param("id")
	radicadoID, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de radicado inválido"})
		return
	}

	userVal, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Usuario no autenticado"})
		return
	}
	user := userVal.(db.Usuario)

	var dto EmitirEventoDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Parámetros inválidos: " + err.Error()})
		return
	}

	evento, err := h.service.EmitirEvento(uint(radicadoID), user.ID, dto)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Evento emitido exitosamente ante la DIAN",
		"evento":  evento,
	})
}
