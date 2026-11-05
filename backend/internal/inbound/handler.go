package inbound

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"sigefae/procesos_aplicacion/proveedor_inbound"
)

type Handler struct {
	ptService *proveedor_inbound.Service
}

func NewHandler(ptService *proveedor_inbound.Service) *Handler {
	return &Handler{
		ptService: ptService,
	}
}

type SyncRequest struct {
	Dias         int    `json:"dias"`
	StartingDate string `json:"starting_date"`
	EndingDate   string `json:"ending_date"`
}

func (h *Handler) Sync(c *gin.Context) {
	if h.ptService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "El servicio de integración con Proveedor Tecnológico DIAN no está configurado en las variables de entorno",
		})
		return
	}

	var req SyncRequest
	_ = c.ShouldBindJSON(&req)

	var stats proveedor_inbound.SyncStats
	var err error

	if req.StartingDate != "" && req.EndingDate != "" {
		var start, end time.Time
		start, err = time.Parse("2006-01-02", req.StartingDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Formato de starting_date inválido (debe ser YYYY-MM-DD)"})
			return
		}
		end, err = time.Parse("2006-01-02", req.EndingDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Formato de ending_date inválido (debe ser YYYY-MM-DD)"})
			return
		}
		// Ajustar hasta el fin del día (23:59:59)
		end = end.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
		stats, err = h.ptService.SyncRange(start, end)
	} else if req.Dias > 0 {
		stats, err = h.ptService.SyncRecent(req.Dias)
	} else {
		stats, err = h.ptService.SyncAuto()
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Sincronización completada exitosamente",
		"stats":   stats,
	})
}
