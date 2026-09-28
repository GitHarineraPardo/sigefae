package user

type CreateRequest struct {
	Nombre     string `json:"nombre" binding:"required"`
	Email      string `json:"email" binding:"required,email"`
	IDExterno  string `json:"id_externo"`
	Contrasena string `json:"contrasena" binding:"required,min=6"`
	CargoID    uint   `json:"cargo_id" binding:"required"`
	RolID      uint   `json:"rol_id" binding:"required"`
}

type Response struct {
	ID        uint    `json:"id"`
	Nombre    string  `json:"nombre"`
	Email     string  `json:"email"`
	IDExterno *string `json:"id_externo,omitempty"`
	CargoID   *uint   `json:"cargo_id,omitempty"`
	Cargo     string  `json:"cargo"`
	Rol       string  `json:"rol"`
	Activo    bool    `json:"activo"`
}

type UpdateRequest struct {
	Nombre    string  `json:"nombre" binding:"required"`
	Email     string  `json:"email" binding:"required,email"`
	IDExterno *string `json:"id_externo"`
	CargoID   uint    `json:"cargo_id" binding:"required"`
	RolID     uint    `json:"rol_id" binding:"required"`
}

type UpdateStatusRequest struct {
	Activo bool `json:"activo"`
}

type UpdatePasswordRequest struct {
	Contrasena string `json:"contrasena" binding:"required,min=8"`
}
