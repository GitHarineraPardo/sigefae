package db

// ---------------------------------------------------------------------------
// Catálogos / tablas de referencia simples
// ---------------------------------------------------------------------------

type EstadoTarea struct {
	ID     uint   `gorm:"primaryKey;column:id" json:"id"`
	Nombre string `gorm:"column:nombre;type:varchar(255)" json:"nombre"`
	Activo bool   `gorm:"column:activo;default:true" json:"activo"`
}

func (EstadoTarea) TableName() string { return "estado_tarea" }

type TipoPago struct {
	ID     uint   `gorm:"primaryKey;column:id" json:"id"`
	Nombre string `gorm:"column:nombre;type:varchar(255)" json:"nombre"`
	Activo bool   `gorm:"column:activo;default:true" json:"activo"`
}

func (TipoPago) TableName() string { return "tipo_pago" }

type TipoRadicacion struct {
	ID     uint   `gorm:"primaryKey;column:id" json:"id"`
	Nombre string `gorm:"column:nombre;type:varchar(255)" json:"nombre"`
	Activo bool   `gorm:"column:activo;default:true" json:"activo"`
}

func (TipoRadicacion) TableName() string { return "tipo_radicacion" }

type EstadoCorreo struct {
	ID     uint   `gorm:"primaryKey;column:id" json:"id"`
	Nombre string `gorm:"column:nombre;type:varchar(255)" json:"nombre"`
	Activo bool   `gorm:"column:activo;default:true" json:"activo"`
}

func (EstadoCorreo) TableName() string { return "estado_correo" }

type ArchivoOrigen struct {
	ID     uint   `gorm:"primaryKey;column:id" json:"id"`
	Nombre string `gorm:"column:nombre;type:varchar(255)" json:"nombre"`
	Activo bool   `gorm:"column:activo;default:true" json:"activo"`
}

func (ArchivoOrigen) TableName() string { return "archivo_origen" }

type EstadoDocumentoRadicado struct {
	ID     uint   `gorm:"primaryKey;column:id" json:"id"`
	Nombre string `gorm:"column:nombre;type:varchar(255)" json:"nombre"`
	Activo bool   `gorm:"column:activo;default:true" json:"activo"`
}

func (EstadoDocumentoRadicado) TableName() string { return "estado_documento_radicado" }

type Rol struct {
	ID     uint   `gorm:"primaryKey;column:id" json:"id"`
	Nombre string `gorm:"column:nombre;type:varchar(255);uniqueIndex" json:"nombre"`
	Activo bool   `gorm:"column:activo;default:true" json:"activo"`
}

func (Rol) TableName() string { return "rol" }

type Moneda struct {
	ID     uint   `gorm:"primaryKey;column:id" json:"id"`
	Nombre string `gorm:"column:nombre;type:varchar(255)" json:"nombre"`
	Codigo string `gorm:"column:codigo;type:varchar(50)" json:"codigo"`
	Activo bool   `gorm:"column:activo;default:true" json:"activo"`
}

func (Moneda) TableName() string { return "moneda" }

type Area struct {
	ID     uint   `gorm:"primaryKey;column:id" json:"id"`
	Nombre string `gorm:"column:nombre;type:varchar(255);uniqueIndex" json:"nombre"`
	Activo bool   `gorm:"column:activo;type:bool" json:"activo"`
}

func (Area) TableName() string { return "area" }

type Origen struct {
	ID     uint   `gorm:"primaryKey;column:id" json:"id"`
	Nombre string `gorm:"column:nombre;type:varchar(255)" json:"nombre"`
}

func (Origen) TableName() string { return "origen" }

type TipoDocumento struct {
	ID     uint   `gorm:"primaryKey;column:id" json:"id"`
	Nombre string `gorm:"column:nombre;type:varchar(255)" json:"nombre"`
	Activo bool   `gorm:"column:activo;default:true" json:"activo"`
}

func (TipoDocumento) TableName() string { return "tipo_documento" }

type TipoPersona struct {
	ID     uint   `gorm:"primaryKey;column:id" json:"id"`
	Nombre string `gorm:"column:nombre;type:varchar(255)" json:"nombre"`
	Activo bool   `gorm:"column:activo" json:"activo"`
}

func (TipoPersona) TableName() string { return "tipo_persona" }

type CategoriaProveedor struct {
	ID          uint   `gorm:"primaryKey;column:id" json:"id"`
	Nombre      string `gorm:"column:nombre;type:varchar(255)" json:"nombre"`
	Descripcion string `gorm:"column:descripcion;type:varchar(500)" json:"descripcion"`
	Activo      bool   `gorm:"column:activo" json:"activo"`
}

func (CategoriaProveedor) TableName() string { return "categoria_proveedor" }

type ActividadEconomica struct {
	ID     uint   `gorm:"primaryKey;column:id" json:"id"`
	Nombre string `gorm:"column:nombre;type:varchar(255)" json:"nombre"`
	Codigo string `gorm:"column:codigo;type:varchar(20)" json:"codigo"`
	Activo bool   `gorm:"column:activo;default:true" json:"activo"`
}

func (ActividadEconomica) TableName() string { return "actividad_economica" }

type Pais struct {
	ID     uint   `gorm:"primaryKey;column:id" json:"id"`
	Nombre string `gorm:"column:nombre;type:varchar(255)" json:"nombre"`
	Codigo string `gorm:"column:codigo;type:varchar(20)" json:"codigo"`
	Activo bool   `gorm:"column:activo;default:true" json:"activo"`
}

func (Pais) TableName() string { return "pais" }

type TipoFactura struct {
	ID     uint   `gorm:"primaryKey;column:id" json:"id"`
	AreaID uint   `gorm:"column:area_id;index:idx_tipo_factura_area" json:"area_id"`
	Area   *Area  `gorm:"foreignKey:AreaID;references:ID" json:"area,omitempty"`
	Nombre string `gorm:"column:nombre;type:varchar(255)" json:"nombre"`
	Activo bool   `gorm:"column:activo;default:true" json:"activo"`
}

func (TipoFactura) TableName() string { return "tipo_factura" }

type Proyecto struct {
	ID     uint   `gorm:"primaryKey;column:id" json:"id"`
	Nombre string `gorm:"column:nombre;type:varchar(255)" json:"nombre"`
	Zona   string `gorm:"column:zona;type:varchar(50);default:'BUCARAMANGA'" json:"zona"`
	Activo bool   `gorm:"column:activo;default:true" json:"activo"`
}

func (Proyecto) TableName() string { return "proyecto" }
