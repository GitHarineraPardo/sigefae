package documento_comercial

import (
	"sigefae/internal/db"
	"time"
)

type CreateDTO struct {
	Tipo                     string     `json:"tipo" binding:"required"`
	NumeroDocumento          string     `json:"numero_documento" binding:"required"`
	OrdenCompra              string     `json:"orden_compra"`
	IDProveedor              uint       `json:"id_proveedor" binding:"required"`
	IDReceptor               uint       `json:"id_receptor" binding:"required"`
	IDArea                   uint       `json:"id_area" binding:"required"`
	TipoFacturaID            *uint      `json:"tipo_factura_id"`
	Asunto                   string     `json:"asunto"`
	FechaDocumento           time.Time  `json:"fecha_documento" binding:"required"`
	FechaVencimiento         *time.Time `json:"fecha_vencimiento"`
	MonedaID                 uint       `json:"moneda_id" binding:"required"`
	Subtotal                 float64    `json:"subtotal"`
	Iva                      float64    `json:"iva"`
	Total                    float64    `json:"total"`
	NumeroFolios             int        `json:"numero_folios"`
	OrientacionSelloRecibido string     `json:"orientacion_sello_recibido"`
	Cufe                     *string    `json:"cufe"`
	Activo                   bool       `json:"activo"`
	// ← NUEVO
	Detalles []db.DetalleDocumentoComercial `json:"detalles"`
}

// UpdateDocumentoComercialDTO — solo lo que un usuario debe completar/corregir
type UpdateDocumentoComercialDTO struct {
	OrdenCompra              string     `json:"orden_compra"`
	IDArea                   uint       `json:"id_area" binding:"required"`
	Asunto                   string     `json:"asunto"`
	FechaVencimiento         *time.Time `json:"fecha_vencimiento"`
	OrientacionSelloRecibido string     `json:"orientacion_sello_recibido" binding:"omitempty,oneof=HORIZONTAL VERTICAL"`
	NumeroFolios             int        `json:"numero_folios"`
}

type UpdateStatusDTO struct {
	Activo bool `json:"activo"`
}
