package paso_ruta

type CreateRequest struct {
	RutaID    uint   `json:"ruta_id" binding:"required"`
	Orden     int    `json:"orden" binding:"required"`
	Nombre    string `json:"nombre" binding:"required"`
	UsuarioID *uint  `json:"usuario_id"`
	CargoID   *uint  `json:"cargo_id"`
}

type UpdateStatusRequest struct {
	Activo bool `json:"activo"`
}

type UpdateRequest struct {
	RutaID    uint   `json:"ruta_id" binding:"required"`
	Orden     int    `json:"orden" binding:"required"`
	Nombre    string `json:"nombre" binding:"required"`
	UsuarioID *uint  `json:"usuario_id"`
	CargoID   *uint  `json:"cargo_id"`
}
