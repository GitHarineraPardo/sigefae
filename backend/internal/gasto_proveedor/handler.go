package gasto_proveedor

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"sigefae/internal/db"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) GetGastos(c *gin.Context) {
	userVal, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Usuario no autenticado"})
		return
	}
	user := userVal.(db.Usuario)

	fechaInicio := c.Query("fecha_inicio")
	fechaFin := c.Query("fecha_fin")

	resp, err := h.service.GetGastos(user, fechaInicio, fechaFin)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}
