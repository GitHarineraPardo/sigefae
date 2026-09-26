package tarea

import "time"

type CreateDTO struct {
	DocumentoRadicadoID uint       `json:"documento_radicado_id" binding:"required"`
	UsuarioAsignadoID   *uint      `json:"usuario_asignado_id"`
	CargoAsignadoID     *uint      `json:"cargo_asignado_id"`
	EstadoID            uint       `json:"estado_id" binding:"required"`
	Descripcion         string     `json:"descripcion" binding:"required"`
	FechaLimite         *time.Time `json:"fecha_limite"`
}

type UpdateDTO struct {
	UsuarioAsignadoID *uint      `json:"usuario_asignado_id"`
	CargoAsignadoID   *uint      `json:"cargo_asignado_id"`
	EstadoID          uint       `json:"estado_id" binding:"required"`
	Descripcion       string     `json:"descripcion" binding:"required"`
	FechaInicio       *time.Time `json:"fecha_inicio"`
	FechaLimite       *time.Time `json:"fecha_limite"`
	FechaFinalizacion *time.Time `json:"fecha_finalizacion"`
}

type DevolverDTO struct {
	TareaDestinoID uint   `json:"tarea_destino_id" binding:"required"`
	Observacion    string `json:"observacion" binding:"required"`
	RetornoDirecto bool   `json:"retorno_directo"`
}
