package gasto_proveedor

import (
	"sigefae/internal/db"
	"time"
)

type ProveedorGastoDTO struct {
	ID               uint    `json:"id"`
	RazonSocial      string  `json:"razon_social"`
	NIT              string  `json:"nit"`
	TotalGastado     float64 `json:"total_gastado"`
	TotalSubtotal    float64 `json:"total_subtotal"`
	TotalIva         float64 `json:"total_iva"`
	CantidadFacturas int     `json:"cantidad_facturas"`
	PorcentajeGasto  float64 `json:"porcentaje_gasto"`
}

type ItemGastoDTO struct {
	Descripcion      string  `json:"descripcion"`
	CantidadTotal    float64 `json:"cantidad_total"`
	TotalGastado     float64 `json:"total_gastado"`
	ProveedorID      uint    `json:"proveedor_id"`
	ProveedorNombre  string  `json:"proveedor_nombre"`
	CantidadFacturas int     `json:"cantidad_facturas"`
	PorcentajeGasto  float64 `json:"porcentaje_gasto"`
}

type NormaGastoDTO struct {
	Codigo           string  `json:"codigo"`
	Nombre           string  `json:"nombre"`
	Sucursal         string  `json:"sucursal"`
	Departamento     string  `json:"departamento"`
	TotalGastado     float64 `json:"total_gastado"`
	CantidadFacturas int     `json:"cantidad_facturas"`
	PorcentajeGasto  float64 `json:"porcentaje_gasto"`
}

type RadicadoGastoDTO struct {
	ID                   uint                    `json:"id"`
	NumeroRadicado       string                  `json:"numero_radicado"`
	FechaRadicacion      time.Time               `json:"fecha_radicacion"`
	EstadoPosesion       string                  `json:"estado_posesion"`
	DocumentoComercialID uint                    `json:"documento_comercial_id"`
	NumeroDocumento      string                  `json:"numero_documento"`
	FechaDocumento       time.Time               `json:"fecha_documento"`
	Total                float64                 `json:"total"`
	Subtotal             float64                 `json:"subtotal"`
	Iva                  float64                 `json:"iva"`
	ProveedorID          uint                    `json:"proveedor_id"`
	ProveedorRazonSocial string                  `json:"proveedor_razon_social"`
	ProveedorNIT         string                  `json:"proveedor_nit"`
	Detalles             []db.DetalleDocumentoComercial `json:"detalles,omitempty"`
	NormasReparto        []db.RadicadoNormaReparto      `json:"normas_reparto,omitempty"`
}

type GastosResponse struct {
	CargoID             *uint               `json:"cargo_id,omitempty"`
	CargoNombre         string              `json:"cargo_nombre"`
	TotalGeneral        float64             `json:"total_general"`
	SubtotalGeneral     float64             `json:"subtotal_general"`
	IvaGeneral          float64             `json:"iva_general"`
	CantidadFacturas    int                 `json:"cantidad_facturas"`
	CantidadProveedores int                 `json:"cantidad_proveedores"`
	PromedioFactura     float64             `json:"promedio_factura"`
	Proveedores         []ProveedorGastoDTO `json:"proveedores"`
	Items               []ItemGastoDTO      `json:"items"`
	Normas              []NormaGastoDTO     `json:"normas"`
	Radicados           []RadicadoGastoDTO  `json:"radicados"`
}
