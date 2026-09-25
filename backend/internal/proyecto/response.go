package proyecto

type Response struct {
	ID     uint   `json:"id"`
	Nombre string `json:"nombre"`
	Zona   string `json:"zona"`
	Activo bool   `json:"activo"`
}
