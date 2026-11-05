package proveedor_inbound

// LoginRequest es el cuerpo para autenticarse en la API del proveedor
type LoginRequest struct {
	U string `json:"u"`
	P string `json:"p"`
}

// LoginResponse es la respuesta de autenticación
type LoginResponse struct {
	DisplayName string `json:"displayName"`
	AccessToken string `json:"accessToken"`
	TenantID    string `json:"tenantId"`
}

// ListDocumentsRequest representa los parámetros de filtro para listar documentos recibidos
type ListDocumentsRequest struct {
	PageNumber           int    `json:"PageNumber"`
	PageSize             int    `json:"PageSize"`
	DocumentTypeCode     string `json:"DocumentTypeCode"`     // ej: "FACTURA-UBL"
	ReceiverStartingDate string `json:"ReceiverStartingDate"` // formato ISO: "2026-09-01T00:00:00"
	ReceiverEndingDate   string `json:"ReceiverEndingDate"`   // formato ISO: "2026-11-30T23:59:00.00"
	LifecycleStatus      string `json:"Lifecycle_Status"`
	FormaPago            string `json:"FormaPago"`
	LDF                  string `json:"LDF"`
	DocumentNumberFull   string `json:"DocumentNumberFull"`
}

// ItemDocumento representa cada documento recibido devuelto en la lista
type ItemDocumento struct {
	FechaFactura             string  `json:"FechaFactura"`
	NumeroFactura            string  `json:"NumeroFactura"`
	NIT                      string  `json:"NIT"`
	RazonSocial              string  `json:"RazonSocial"`
	FormaPago                string  `json:"FormaPago"`
	UltimaFechaActualizacion *string `json:"UltimaFechaActualizacion"`
	Evento                   string  `json:"Evento"` // ACKNOWLEDGED, RECEIVEDGOODS, ACCEPTED, REJECTED
	FechaRecepcion           string  `json:"FechaRecepcion"`
	IVA                      string  `json:"IVA"`
	Monto                    string  `json:"Monto"`
	LDF                      string  `json:"LDF"`
	TipoDocumento            string  `json:"TipoDocumento"`
	CUFE                     string  `json:"CUFE"`
}

// ListDocumentsResponse representa la respuesta paginada al listar documentos
type ListDocumentsResponse struct {
	Items           []ItemDocumento `json:"items"`
	PageNumber      int             `json:"pageNumber"`
	PageSize        int             `json:"pageSize"`
	TotalItemCount  int             `json:"totalItemCount"`
	TotalPages      int             `json:"totalPages"`
	AdditionalItems any             `json:"additionalItems"`
}
