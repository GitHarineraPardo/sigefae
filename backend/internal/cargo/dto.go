package cargo

import "time"

type CargoDTO struct {
	ID        uint      `json:"id"`
	Nombre    string    `json:"nombre"`
	Activo    bool      `json:"activo"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateDTO struct {
	Nombre string `json:"nombre" binding:"required"`
}

type UpdateDTO struct {
	Nombre string `json:"nombre"`
	Activo *bool  `json:"activo"`
}
