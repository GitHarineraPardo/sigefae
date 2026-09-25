package documento_radicado

type CreateDTO struct {
	DocumentoComercialID uint                   `json:"documento_comercial_id" binding:"required"`
	TipoRadicacionID     uint                   `json:"tipo_radicacion_id" binding:"required"`
	RutaID               uint                   `json:"ruta_id" binding:"required"`
	MetodoPagoID         uint                   `json:"metodo_pago_id" binding:"required"`
	NormasReparto        []NormaRepartoInputDTO `json:"normas_reparto"`
	ActivosFijos         []ActivoFijoInputDTO   `json:"activos_fijos"`
	NumeroRadicado       string                 `json:"numero_radicado"` // opcional, si vacío se autogenera
	EsMalambo            bool                   `json:"es_malambo"`
}
type UpdateDTO struct {
	TipoRadicacionID uint   `json:"tipo_radicacion_id"`
	RutaID           uint   `json:"ruta_id"`
	MetodoPagoID     uint   `json:"metodo_pago_id"`
	NumeroRadicado   string `json:"numero_radicado"`
	UsuarioActualID  uint   `json:"usuario_actual_id"`
	EstadoPosesion   string `json:"estado_posesion"`
	PasoActualID     uint   `json:"paso_actual_id"`
	EstadoID         uint   `json:"estado_id"`
}

type NormaRepartoInputDTO struct {
	NormaRepartoID uint    `json:"norma_reparto_id" binding:"required"`
	Porcentaje     float64 `json:"porcentaje" binding:"required"`
	Proyecto       string  `json:"proyecto"`
	Descripcion    string  `json:"descripcion"`
}
