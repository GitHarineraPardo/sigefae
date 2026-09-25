package archivo

import (
	"time"

	"sigefae/internal/db"
)

type Response struct {
	ID                  uint      `json:"id"`
	DocumentoRadicadoID uint      `json:"documento_radicado_id"`
	Nombre              string    `json:"nombre"`
	Extension           string    `json:"extension"`
	Peso                int64     `json:"peso"`
	Ruta                string    `json:"ruta"`
	OrigenID            uint      `json:"origen_id"`
	CreadoPorID         uint      `json:"creado_por_id"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

func toResponse(archivo db.Archivo) Response {

	return Response{
		ID:                  archivo.ID,
		DocumentoRadicadoID: archivo.DocumentoRadicadoID,
		Nombre:              archivo.Nombre,
		Extension:           archivo.Extension,
		Peso:                archivo.Peso,
		Ruta:                archivo.Ruta,
		OrigenID:            archivo.OrigenID,
		CreadoPorID:         archivo.CreadoPorID,
		CreatedAt:           archivo.CreatedAt,
		UpdatedAt:           archivo.UpdatedAt,
	}
}
