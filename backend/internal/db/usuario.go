package db

import (
	"time"

	"gorm.io/gorm"
)

// ---------------------------------------------------------------------------
// Usuarios / roles / rutas de flujo
// ---------------------------------------------------------------------------

type Usuario struct {
	ID             uint           `gorm:"primaryKey;column:id" json:"id"`
	Nombre         string         `gorm:"column:nombre;type:varchar(255)" json:"nombre"`
	Email          string         `gorm:"column:email;type:varchar(255);uniqueIndex" json:"email"`
	HashContrasena string         `gorm:"column:contrasena;type:varchar(255)" json:"-"`
	CargoID        *uint          `gorm:"column:cargo_id;index:idx_usuario_cargo_id" json:"cargo_id,omitempty"`
	Cargo          *Cargo         `gorm:"foreignKey:CargoID;references:ID" json:"cargo,omitempty"`
	Activo         bool           `gorm:"column:activo;default:true" json:"activo"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`

	RolID uint `gorm:"column:rol_id;index:idx_usuario_rol_id" json:"rol_id"`
	Rol   *Rol `gorm:"foreignKey:RolID;references:ID" json:"rol,omitempty"`

	Comentarios    []Comentario   `gorm:"foreignKey:UsuarioID" json:"-"`
	Trazabilidades []Trazabilidad `gorm:"foreignKey:UsuarioID" json:"-"`
	Asignaciones   []PasoRuta     `gorm:"foreignKey:UsuarioID" json:"-"`
}

func (Usuario) TableName() string { return "usuario" }

type Ruta struct {
	ID        uint           `gorm:"primaryKey;column:id" json:"id"`
	Nombre    string         `gorm:"column:nombre;type:varchar(255);uniqueIndex:uk_ruta_nombre_area_zona" json:"nombre"`
	Zona      string         `gorm:"column:zona;type:varchar(100);default:'BUCARAMANGA';uniqueIndex:uk_ruta_nombre_area_zona" json:"zona"`
	Version   float64        `gorm:"column:version" json:"version"`
	AreaID    uint           `gorm:"column:area_id;index:idx_ruta_area;uniqueIndex:uk_ruta_nombre_area_zona" json:"area_id"`
	Area      *Area          `gorm:"foreignKey:AreaID;references:ID" json:"area,omitempty"`
	Pasos     []PasoRuta     `gorm:"foreignKey:RutaID" json:"pasos,omitempty"`
	Activo    bool           `gorm:"column:activo;default:true" json:"activo"`
	CreatedAt time.Time      `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Ruta) TableName() string { return "ruta" }

// PasoRuta representa cada paso secuencial de una ruta versionada.
type PasoRuta struct {
	ID        uint           `gorm:"primaryKey;column:id" json:"id"`
	RutaID    uint           `gorm:"column:ruta_id;index:idx_paso_ruta_id" json:"ruta_id"`
	Ruta      *Ruta          `gorm:"foreignKey:RutaID;references:ID" json:"ruta,omitempty"`
	Orden     int            `gorm:"column:orden;index:idx_paso_orden" json:"orden"`
	Nombre    string         `gorm:"column:nombre;type:varchar(255)" json:"nombre"`
	UsuarioID *uint          `gorm:"column:usuario_id;index:idx_paso_usuario_id" json:"usuario_id,omitempty"`
	CargoID   *uint          `gorm:"column:cargo_id;index:idx_paso_cargo_id" json:"cargo_id,omitempty"`
	Cargo     *Cargo         `gorm:"foreignKey:CargoID;references:ID" json:"cargo,omitempty"`
	Usuario   *Usuario       `gorm:"foreignKey:UsuarioID;references:ID" json:"usuario,omitempty"`
	Activo    bool           `gorm:"column:activo;default:true" json:"activo"`
	CreatedAt time.Time      `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (PasoRuta) TableName() string { return "paso_ruta" }

