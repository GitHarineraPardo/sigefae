package documento_comercial

import (
	"time"

	"sigefae/internal/db"
)

type ProveedorResponse struct {
	ID              uint   `json:"id"`
	RazonSocial     string `json:"razon_social"`
	NumeroDocumento string `json:"numero_documento"`
}

type ReceptorResponse struct {
	ID              uint   `json:"id"`
	Nombre          string `json:"nombre"`
	NumeroDocumento string `json:"numero_documento"`
}

type AreaResponse struct {
	ID     uint   `json:"id"`
	Nombre string `json:"nombre"`
}

type TipoFacturaResponse struct {
	ID     uint   `json:"id"`
	Nombre string `json:"nombre"`
}

type MonedaResponse struct {
	ID     uint   `json:"id"`
	Nombre string `json:"nombre"`
}


type DetalleResponse struct {
	ID          uint    `json:"id"`
	Descripcion string  `json:"descripcion"`
	Cantidad    float64 `json:"cantidad"`
	ValorUnit   float64 `json:"valor_unitario"`
	IvaUnit     float64 `json:"iva_unitario"`
	Total       float64 `json:"total"`
}

type Response struct {
	ID uint `json:"id"`

	Tipo            string `json:"tipo"`
	NumeroDocumento string `json:"numero_documento"`
	OrdenCompra     string `json:"orden_compra"`

	IDProveedor uint               `json:"id_proveedor"`
	Proveedor   *ProveedorResponse `json:"proveedor,omitempty"`

	IDReceptor uint              `json:"id_receptor"`
	Receptor   *ReceptorResponse `json:"receptor,omitempty"`

	IDArea uint          `json:"id_area"`
	Area   *AreaResponse `json:"area,omitempty"`

	TipoFacturaID *uint                `json:"tipo_factura_id,omitempty"`
	TipoFactura   *TipoFacturaResponse `json:"tipo_factura,omitempty"`

	Asunto string `json:"asunto"`

	FechaDocumento   time.Time  `json:"fecha_documento"`
	FechaVencimiento *time.Time `json:"fecha_vencimiento,omitempty"`
	FormaPago        string     `json:"forma_pago"`


	MonedaID uint            `json:"moneda_id"`
	Moneda   *MonedaResponse `json:"moneda,omitempty"`

	Subtotal float64 `json:"subtotal"`
	Iva      float64 `json:"iva"`
	Total    float64 `json:"total"`

	NumeroFolios             int    `json:"numero_folios"`
	OrientacionSelloRecibido string `json:"orientacion_sello_recibido"`

	Cufe *string `json:"cufe,omitempty"`

	// Campos DIAN Inbound (PT API)
	LDF                *string    `json:"ldf,omitempty"`
	EventoDIAN         string     `json:"evento_dian,omitempty"`
	FechaRecepcionDIAN *time.Time `json:"fecha_recepcion_dian,omitempty"`
	RutaXML            string     `json:"ruta_xml,omitempty"`
	RutaPDF            string     `json:"ruta_pdf,omitempty"`

	Detalles []DetalleResponse `json:"detalles,omitempty"`

	Activo bool `json:"activo"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func toResponse(documento db.DocumentoComercial) Response {

	response := Response{
		ID:                       documento.ID,
		Tipo:                     documento.Tipo,
		NumeroDocumento:          documento.NumeroDocumento,
		OrdenCompra:              documento.OrdenCompra,
		IDProveedor:              documento.IDProveedor,
		IDReceptor:               documento.IDReceptor,
		IDArea:                   documento.IDArea,
		TipoFacturaID:            documento.TipoFacturaID,
		Asunto:                   documento.Asunto,
		FechaDocumento:           documento.FechaDocumento,
		FechaVencimiento:         documento.FechaVencimiento,
		FormaPago:                documento.FormaPago,
		MonedaID:                 documento.MonedaID,
		Subtotal:                 documento.Subtotal,
		Iva:                      documento.Iva,
		Total:                    documento.Total,
		NumeroFolios:             documento.NumeroFolios,
		OrientacionSelloRecibido: documento.OrientacionSelloRecibido,
		Cufe:                     documento.Cufe,
		LDF:                      documento.LDF,
		EventoDIAN:               documento.EventoDIAN,
		FechaRecepcionDIAN:       documento.FechaRecepcionDIAN,
		RutaXML:                  documento.RutaXML,
		RutaPDF:                  documento.RutaPDF,
		Activo:                   documento.Activo,
		CreatedAt:                documento.CreatedAt,
		UpdatedAt:                documento.UpdatedAt,
	}

	if documento.Proveedor != nil {
		response.Proveedor = &ProveedorResponse{
			ID:              documento.Proveedor.ID,
			RazonSocial:     documento.Proveedor.RazonSocial,
			NumeroDocumento: documento.Proveedor.NumeroDocumento,
		}
	}

	if documento.Receptor != nil {
		response.Receptor = &ReceptorResponse{
			ID:              documento.Receptor.ID,
			Nombre:          documento.Receptor.Nombre,
			NumeroDocumento: documento.Receptor.NumeroDocumento,
		}
	}

	if documento.Area != nil {
		response.Area = &AreaResponse{
			ID:     documento.Area.ID,
			Nombre: documento.Area.Nombre,
		}
	}

	if documento.TipoFactura != nil {
		response.TipoFactura = &TipoFacturaResponse{
			ID:     documento.TipoFactura.ID,
			Nombre: documento.TipoFactura.Nombre,
		}
	}

	if documento.Moneda != nil {
		response.Moneda = &MonedaResponse{
			ID:     documento.Moneda.ID,
			Nombre: documento.Moneda.Nombre,
		}
	}

	if len(documento.Detalles) > 0 {
		for _, d := range documento.Detalles {
			response.Detalles = append(response.Detalles, DetalleResponse{
				ID:          d.ID,
				Descripcion: d.Descripcion,
				Cantidad:    d.Cantidad,
				ValorUnit:   d.ValorUnit,
				IvaUnit:     d.IvaUnit,
				Total:       d.Total,
			})
		}
	}

	return response
}
