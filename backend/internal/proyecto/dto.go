package proyecto

type Request struct {
	Nombre string `json:"nombre" binding:"required"`
	Zona   string `json:"zona" binding:"required"`
}
