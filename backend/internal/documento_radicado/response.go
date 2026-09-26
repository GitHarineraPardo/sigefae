package documento_radicado

import (
	"time"

	"sigefae/internal/db"
)

type DocumentoComercialResponse struct {
	ID              uint   `json:"id"`
	Tipo            string `json:"tipo"`
	NumeroDocumento string `json:"numero_documento"`
}

type TipoRadicacionResponse struct {
	ID     uint   `json:"id"`
	Nombre string `json:"nombre"`
}

type RutaResponse struct {
	ID     uint   `json:"id"`
	Nombre string `json:"nombre"`
}

type UsuarioResponse struct {
	ID     uint   `json:"id"`
	Nombre string `json:"nombre"`
}

type PasoRutaResponse struct {
	ID     uint   `json:"id"`
	Nombre string `json:"nombre"`
}

type EstadoDocumentoRadicadoResponse struct {
	ID     uint   `json:"id"`
	Nombre string `json:"nombre"`
}

type CodigoQrResponse struct {
	ID  uint   `json:"id"`
	Url string `json:"url"`
}

type MetodoPagoResponse struct {
	ID     uint   `json:"id"`
	Nombre string `json:"nombre"`
}

type Response struct {
	ID                     uint                             `json:"id"`
	DocumentoComercialID   uint                             `json:"documento_comercial_id"`
	DocumentoComercial     *DocumentoComercialResponse      `json:"documento_comercial,omitempty"`
	TipoRadicacionID       uint                             `json:"tipo_radicacion_id"`
	TipoRadicacion         *TipoRadicacionResponse          `json:"tipo_radicacion,omitempty"`
	RutaID                 uint                             `json:"ruta_id"`
	Ruta                   *RutaResponse                    `json:"ruta,omitempty"`
	NumeroRadicado         string                           `json:"numero_radicado"`
	FechaRadicacion        time.Time                        `json:"fecha_radicacion"`
	UsuarioActualID        *uint                            `json:"usuario_actual_id"`
	CargoActualID          *uint                            `json:"cargo_actual_id,omitempty"`
	CargoActual            string                           `json:"cargo_actual"`
	UsuarioActual          *UsuarioResponse                 `json:"usuario_actual,omitempty"`
	EstadoPosesion         string                           `json:"estado_posesion"`
	PasoActualID           *uint                            `json:"paso_actual_id,omitempty"`
	PasoActual             *PasoRutaResponse                `json:"paso_actual,omitempty"`
	PasoPendienteRetornoID *uint                            `json:"paso_pendiente_retorno_id,omitempty"`
	PasoPendienteRetorno   *PasoRutaResponse                `json:"paso_pendiente_retorno,omitempty"`
	EstadoID               uint                             `json:"estado_id"`
	Estado                 *EstadoDocumentoRadicadoResponse `json:"estado,omitempty"`
	UltimaActividad        *time.Time                       `json:"ultima_actividad,omitempty"`
	NormasReparto          []NormaRepartoResponse           `json:"normas_reparto,omitempty"`
	QrID                   uint                             `json:"qr_id"`
	Qr                     *CodigoQrResponse                `json:"qr,omitempty"`
	MetodoPagoID           uint                             `json:"metodo_pago_id"`
	MetodoPago             *MetodoPagoResponse              `json:"metodo_pago,omitempty"`
	Archivos               []ArchivoResponse                `json:"archivos,omitempty"`
	UpdatedAt              time.Time                        `json:"updated_at"`
}

func toResponse(documento db.DocumentoRadicado) Response {

	cargoActual := ""
	if documento.CargoActual != nil {
		cargoActual = documento.CargoActual.Nombre
	}
	response := Response{
		ID:                     documento.ID,
		DocumentoComercialID:   documento.DocumentoComercialID,
		TipoRadicacionID:       documento.TipoRadicacionID,
		RutaID:                 documento.RutaID,
		NumeroRadicado:         documento.NumeroRadicado,
		FechaRadicacion:        documento.FechaRadicacion,
		UsuarioActualID:        documento.UsuarioActualID,
		CargoActualID:          documento.CargoActualID,
		CargoActual:            cargoActual,
		EstadoPosesion:         documento.EstadoPosesion,
		PasoActualID:           documento.PasoActualID,
		PasoPendienteRetornoID: documento.PasoPendienteRetornoID,
		EstadoID:               documento.EstadoID,
		UltimaActividad:        documento.UltimaActividad,
		QrID:                   documento.QrID,
		MetodoPagoID:           documento.MetodoPagoID,
		UpdatedAt:              documento.UpdatedAt,
	}

	if documento.DocumentoComercial != nil {
		response.DocumentoComercial = &DocumentoComercialResponse{
			ID:              documento.DocumentoComercial.ID,
			Tipo:            documento.DocumentoComercial.Tipo,
			NumeroDocumento: documento.DocumentoComercial.NumeroDocumento,
		}
	}

	if documento.TipoRadicacion != nil {
		response.TipoRadicacion = &TipoRadicacionResponse{
			ID:     documento.TipoRadicacion.ID,
			Nombre: documento.TipoRadicacion.Nombre,
		}
	}

	if documento.Ruta != nil {
		response.Ruta = &RutaResponse{
			ID:     documento.Ruta.ID,
			Nombre: documento.Ruta.Nombre,
		}
	}

	if documento.UsuarioActual != nil {
		response.UsuarioActual = &UsuarioResponse{
			ID:     documento.UsuarioActual.ID,
			Nombre: documento.UsuarioActual.Nombre,
		}
	}

	if documento.PasoActual != nil {
		response.PasoActual = &PasoRutaResponse{
			ID:     documento.PasoActual.ID,
			Nombre: documento.PasoActual.Nombre,
		}
	}

	if documento.PasoPendienteRetorno != nil {
		response.PasoPendienteRetorno = &PasoRutaResponse{
			ID:     documento.PasoPendienteRetorno.ID,
			Nombre: documento.PasoPendienteRetorno.Nombre,
		}
	}

	if documento.Estado != nil {
		response.Estado = &EstadoDocumentoRadicadoResponse{
			ID:     documento.Estado.ID,
			Nombre: documento.Estado.Nombre,
		}
	}

	if documento.Qr != nil {
		response.Qr = &CodigoQrResponse{
			ID:  documento.Qr.ID,
			Url: documento.Qr.Url,
		}
	}

	if documento.MetodoPago != nil {
		response.MetodoPago = &MetodoPagoResponse{
			ID:     documento.MetodoPago.ID,
			Nombre: documento.MetodoPago.Nombre,
		}
	}
	if len(documento.NormasReparto) > 0 {
		for _, nr := range documento.NormasReparto {
			item := NormaRepartoResponse{
				ID:             nr.ID,
				NormaRepartoID: nr.NormaRepartoID,
				Porcentaje:     nr.Porcentaje,
			}
			if nr.NormaReparto != nil {
				item.Codigo = nr.NormaReparto.Codigo
				item.Nombre = nr.NormaReparto.Nombre
				item.Sucursal = nr.NormaReparto.Sucursal
				item.Departamento = nr.NormaReparto.Departamento
				item.Proyecto = nr.NormaReparto.Proyecto
				item.Descripcion = nr.NormaReparto.Descripcion
			}
			response.NormasReparto = append(response.NormasReparto, item)
		}
	}

	if len(documento.Archivos) > 0 {
		for _, arch := range documento.Archivos {
			item := ArchivoResponse{
				ID:        arch.ID,
				Nombre:    arch.Nombre,
				Extension: arch.Extension,
				Peso:      arch.Peso,
				Ruta:      arch.Ruta,
				OrigenID:  arch.OrigenID,
				CreatedAt: arch.CreatedAt,
			}
			response.Archivos = append(response.Archivos, item)
		}
	}

	return response
}

type NormaRepartoResponse struct {
	ID             uint    `json:"id"`
	NormaRepartoID uint    `json:"norma_reparto_id"`
	Codigo         string  `json:"codigo"`
	Nombre         string  `json:"nombre"`
	Sucursal       string  `json:"sucursal"`
	Departamento   string  `json:"departamento"`
	Proyecto       string  `json:"proyecto,omitempty"`
	Descripcion    *string `json:"descripcion,omitempty"`
	Porcentaje     float64 `json:"porcentaje"`
}

type ArchivoResponse struct {
	ID        uint      `json:"id"`
	Nombre    string    `json:"nombre"`
	Extension string    `json:"extension"`
	Peso      int64     `json:"peso"`
	Ruta      string    `json:"ruta"`
	OrigenID  uint      `json:"origen_id"`
	CreatedAt time.Time `json:"created_at"`
}
