package evento_dian

type EmitirEventoDTO struct {
	// Tipo puede ser "030", "032", "034", "031" o "ACUSE", "RECIBO", "ACEPTACION", "RECLAMO"
	Tipo         string `json:"tipo" binding:"required"`
	Motivo       string `json:"motivo"`
	Observacion  string `json:"observacion"`
	Comentario   string `json:"comentario"`
	CodigoMotivo string `json:"codigo_motivo"` // Requerido para Reclamo (031), ej: "01", "02", "03", "04"
	Medio        string `json:"medio"`         // Opcional, buzón receptor
}
