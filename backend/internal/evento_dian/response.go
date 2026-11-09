package evento_dian

import "time"

type Response struct {
	ID                   uint      `json:"id"`
	DocumentoComercialID uint      `json:"documento_comercial_id"`
	TipoEvento           string    `json:"tipo_evento"`
	CodigoDIAN           string    `json:"codigo_dian"`
	FechaEnvio           time.Time `json:"fecha_envio"`
	Exitoso              bool      `json:"exitoso"`
	MensajeRespuesta     string    `json:"mensaje_respuesta"`
	UsuarioID            *uint     `json:"usuario_id,omitempty"`
	UsuarioNombre        string    `json:"usuario_nombre,omitempty"`
	Observacion          string    `json:"observacion"`
	CreatedAt            time.Time `json:"created_at"`
}
