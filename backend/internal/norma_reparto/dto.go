package norma_reparto

type CreateDTO struct {
	Codigo       string  `json:"codigo" binding:"required"`
	Nombre       string  `json:"nombre" binding:"required"`
	Sucursal     string  `json:"sucursal" binding:"required"`
	Departamento string  `json:"departamento" binding:"required"`
	Tipo         *string `json:"tipo"`
	TarifaIva    *string `json:"tarifa_iva"`
	Proyecto     string  `json:"proyecto" binding:"required"`
	Descripcion  *string `json:"descripcion" binding:"max=100"`
}

type UpdateDTO struct {
	Codigo       string  `json:"codigo"`
	Nombre       string  `json:"nombre"`
	Sucursal     string  `json:"sucursal"`
	Departamento string  `json:"departamento"`
	Tipo         *string `json:"tipo"`
	TarifaIva    *string `json:"tarifa_iva"`
	Proyecto     string  `json:"proyecto"`
	Descripcion  *string `json:"descripcion" binding:"max=100"`
}

type UpdateStatusDTO struct {
	Activo bool `json:"activo"`
}
