package user

import "sigefae/internal/db"

func toResponse(user db.Usuario) Response {

	response := Response{
		ID:        user.ID,
		Nombre:    user.Nombre,
		Email:     user.Email,
		IDExterno: user.IDExterno,
		CargoID:   user.CargoID,
		Activo:    user.Activo,
	}
	if user.Cargo != nil {
		response.Cargo = user.Cargo.Nombre
	}

	if user.Rol != nil {
		response.Rol = user.Rol.Nombre
	}

	return response
}
