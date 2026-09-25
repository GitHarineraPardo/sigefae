package paso_ruta

import (
	"time"

	"sigefae/internal/db"
)

type Response struct {
	ID uint `json:"id"`

	RutaID uint   `json:"ruta_id"`
	Ruta   string `json:"ruta"`
	Zona   string `json:"zona"`
	Area   string `json:"area"`

	Orden  int    `json:"orden"`
	Nombre string `json:"nombre"`

	UsuarioID uint   `json:"usuario_id"`
	Usuario   string `json:"usuario"`

	Activo bool `json:"activo"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func toResponse(p db.PasoRuta) Response {

	ruta := ""
	usuario := ""
	zona := ""
	area := ""

	if p.Ruta != nil {
		ruta = p.Ruta.Nombre
		zona = p.Ruta.Zona
		if p.Ruta.Area != nil {
			area = p.Ruta.Area.Nombre
		}
	}

	if p.Usuario != nil {
		usuario = p.Usuario.Nombre
	}

	return Response{
		ID: p.ID,

		RutaID: p.RutaID,
		Ruta:   ruta,
		Zona:   zona,
		Area:   area,

		Orden:  p.Orden,
		Nombre: p.Nombre,

		UsuarioID: p.UsuarioID,
		Usuario:   usuario,

		Activo: p.Activo,

		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}
}
