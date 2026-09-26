package tarea

import (
	"time"

	"sigefae/internal/db"
)

type Response struct {
	ID                  uint            `json:"id"`
	DocumentoRadicadoID uint            `json:"documento_radicado_id"`
	UsuarioAsignadoID   *uint           `json:"usuario_asignado_id"`
	UsuarioAsignado     *db.Usuario     `json:"usuario_asignado,omitempty"`
	CargoAsignadoID     *uint           `json:"cargo_asignado_id,omitempty"`
	CargoAsignado       string          `json:"cargo_asignado"`
	EstadoID            uint            `json:"estado_id"`
	Estado              *db.EstadoTarea `json:"estado,omitempty"`
	Descripcion         string          `json:"descripcion"`
	FechaAsignacion     time.Time       `json:"fecha_asignacion"`
	FechaInicio         *time.Time      `json:"fecha_inicio"`
	FechaLimite         *time.Time      `json:"fecha_limite"`
	FechaFinalizacion   *time.Time      `json:"fecha_finalizacion"`
	CreatedAt           time.Time       `json:"created_at"`
}

func toResponse(tarea db.Tarea) Response {

	cargoAsignado := ""
	if tarea.CargoAsignadoCargo != nil {
		cargoAsignado = tarea.CargoAsignadoCargo.Nombre
	}
	return Response{
		ID:                  tarea.ID,
		DocumentoRadicadoID: tarea.DocumentoRadicadoID,
		UsuarioAsignadoID:   tarea.UsuarioAsignadoID,
		UsuarioAsignado:     tarea.UsuarioAsignado,
		CargoAsignadoID:     tarea.CargoAsignadoID,
		CargoAsignado:       cargoAsignado,
		EstadoID:            tarea.EstadoID,
		Estado:              tarea.Estado,
		Descripcion:         tarea.Descripcion,
		FechaAsignacion:     tarea.FechaAsignacion,
		FechaInicio:         tarea.FechaInicio,
		FechaLimite:         tarea.FechaLimite,
		FechaFinalizacion:   tarea.FechaFinalizacion,
		CreatedAt:           tarea.CreatedAt,
	}
}
