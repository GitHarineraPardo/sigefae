package db

import "time"

// DocumentoComercial unifica Facturas (física/electrónica), Cuentas de Cobro, Notas de Crédito, etc.
type DocumentoComercial struct {
	ID                       uint                        `gorm:"primaryKey;column:id" json:"id"`
	Tipo                     string                      `gorm:"column:tipo;type:varchar(50);index:idx_doc_comercial_tipo" json:"tipo"` // "FACTURA_ELECTRONICA", "FACTURA_FISICA", "CUENTA_COBRO", "NOTA_CREDITO", "NOTA_DEBITO"
	NumeroDocumento          string                      `gorm:"column:numero_documento;type:varchar(100);uniqueIndex:uk_proveedor_documento_comercial" json:"numero_documento"`
	OrdenCompra              string                      `gorm:"column:orden_compra;type:varchar(100)" json:"orden_compra"`
	IDProveedor              uint                        `gorm:"column:id_proveedor;uniqueIndex:uk_proveedor_documento_comercial" json:"id_proveedor"`
	Proveedor                *Proveedor                  `gorm:"foreignKey:IDProveedor;references:ID" json:"proveedor,omitempty"`
	IDReceptor               uint                        `gorm:"column:id_receptor;index:idx_doc_comercial_receptor" json:"id_receptor"`
	Receptor                 *Receptor                   `gorm:"foreignKey:IDReceptor;references:ID" json:"receptor,omitempty"`
	IDArea                   uint                        `gorm:"column:id_area;index:idx_doc_comercial_area" json:"id_area"`
	Area                     *Area                       `gorm:"foreignKey:IDArea;references:ID" json:"area,omitempty"`
	TipoFacturaID            *uint                       `gorm:"column:tipo_factura_id;index:idx_doc_comercial_tipo_fac" json:"tipo_factura_id,omitempty"`
	TipoFactura              *TipoFactura                `gorm:"foreignKey:TipoFacturaID;references:ID" json:"tipo_factura,omitempty"`
	Asunto                   string                      `gorm:"column:asunto;type:varchar(255)" json:"asunto"`
	FechaDocumento           time.Time                   `gorm:"column:fecha_documento;index:idx_doc_comercial_fecha" json:"fecha_documento"`
	FechaVencimiento         *time.Time                  `gorm:"column:fecha_vencimiento;index:idx_doc_comercial_vencimiento" json:"fecha_vencimiento"`
	FormaPago                string                      `gorm:"column:forma_pago;type:varchar(50)" json:"forma_pago"`
	MonedaID                 uint                        `gorm:"column:moneda_id;index:idx_doc_comercial_moneda" json:"moneda_id"`
	Moneda                   *Moneda                     `gorm:"foreignKey:MonedaID;references:ID" json:"moneda,omitempty"`
	Subtotal                 float64                     `gorm:"column:subtotal" json:"subtotal"`
	Iva                      float64                     `gorm:"column:iva" json:"iva"`
	Total                    float64                     `gorm:"column:total;index:idx_doc_comercial_total" json:"total"`
	NumeroFolios             int                         `gorm:"column:numero_folios" json:"numero_folios"`
	OrientacionSelloRecibido string                      `gorm:"column:orientacion_sello_recibido;type:varchar(50)" json:"orientacion_sello_recibido"`
	Cufe                     *string                     `gorm:"column:cufe;type:varchar(255);uniqueIndex:uk_doc_comercial_cufe" json:"cufe,omitempty"` // Nulable: cuentas de cobro no tienen CUFE
	LDF                      *string                     `gorm:"column:ldf;type:varchar(255);uniqueIndex:uk_doc_comercial_ldf" json:"ldf,omitempty"`   // Identificador del proveedor tecnológico
	EventoDIAN               string                      `gorm:"column:evento_dian;type:varchar(50)" json:"evento_dian,omitempty"`                     // ACKNOWLEDGED, RECEIVEDGOODS, ACCEPTED, etc.
	FechaRecepcionDIAN       *time.Time                  `gorm:"column:fecha_recepcion_dian" json:"fecha_recepcion_dian,omitempty"`
	RutaXML                  string                      `gorm:"column:ruta_xml;type:varchar(500)" json:"ruta_xml,omitempty"`
	RutaPDF                  string                      `gorm:"column:ruta_pdf;type:varchar(500)" json:"ruta_pdf,omitempty"`
	Detalles                 []DetalleDocumentoComercial `gorm:"foreignKey:DocumentoComercialID" json:"detalles,omitempty"`
	Activo                   bool                        `gorm:"column:activo;default:true" json:"activo"`
	CreatedAt                time.Time                   `gorm:"column:created_at" json:"created_at"`
	UpdatedAt                time.Time                   `gorm:"column:updated_at" json:"updated_at"`
}

func (DocumentoComercial) TableName() string { return "documento_comercial" }

type DetalleDocumentoComercial struct {
	ID                   uint                `gorm:"primaryKey;column:id" json:"id"`
	DocumentoComercialID uint                `gorm:"column:documento_comercial_id;index:idx_detalle_doc" json:"documento_comercial_id"`
	DocumentoComercial   *DocumentoComercial `gorm:"foreignKey:DocumentoComercialID;references:ID" json:"documento_comercial,omitempty"`
	Descripcion          string              `gorm:"column:descripcion;type:varchar(255)" json:"descripcion"`
	ValorUnit            float64             `gorm:"column:valor_unitario" json:"valor_unitario"`
	IvaUnit              float64             `gorm:"column:iva_unitario" json:"iva_unitario"`
	Cantidad             float64             `gorm:"column:cantidad" json:"cantidad"`
	Total                float64             `gorm:"column:total" json:"total"`
	Activo               bool                `gorm:"column:activo;default:true" json:"activo"`
}

func (DetalleDocumentoComercial) TableName() string { return "detalle_documento_comercial" }
