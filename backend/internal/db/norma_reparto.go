package db

import (
	"time"

	"gorm.io/gorm"
)

type NormaReparto struct {
	ID           uint           `gorm:"primaryKey;column:id" json:"id"`
	Codigo       string         `gorm:"column:codigo;type:varchar(20);not null;uniqueIndex:uk_norma_reparto_codigo" json:"codigo"`
	Nombre       string         `gorm:"column:nombre;type:varchar(100);not null" json:"nombre"`
	Sucursal     string         `gorm:"column:sucursal;type:varchar(50);not null" json:"sucursal"`
	Departamento string         `gorm:"column:departamento;type:varchar(50);not null" json:"departamento"`
	Tipo         *string        `gorm:"column:tipo;type:varchar(20)" json:"tipo,omitempty"`
	TarifaIva    *string        `gorm:"column:tarifa_iva;type:varchar(10)" json:"tarifa_iva,omitempty"`
	Proyecto     string         `gorm:"column:proyecto;type:varchar(100);not null;default:''" json:"proyecto"`
	Descripcion  *string        `gorm:"column:descripcion;type:text" json:"descripcion,omitempty"`
	Activo       bool           `gorm:"column:activo;default:true" json:"activo"`
	CreatedAt    time.Time      `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

func (NormaReparto) TableName() string { return "norma_reparto" }

type RadicadoNormaReparto struct {
	ID                  uint           `gorm:"primaryKey;column:id" json:"id"`
	DocumentoRadicadoID uint           `gorm:"column:documento_radicado_id;not null;index:idx_rnr_radicado" json:"documento_radicado_id"`
	NormaRepartoID      uint           `gorm:"column:norma_reparto_id;not null;index:idx_rnr_norma" json:"norma_reparto_id"`
	Porcentaje          float64        `gorm:"column:porcentaje;type:decimal(5,2);not null" json:"porcentaje"`
	Proyecto            string         `gorm:"column:proyecto;type:varchar(100);default:''" json:"proyecto"`
	Descripcion         string         `gorm:"column:descripcion;type:text;default:''" json:"descripcion"`
	CreadoPorID         uint           `gorm:"column:creado_por_id;index:idx_rnr_creado_por" json:"creado_por_id"`
	CreadoPor           *Usuario       `gorm:"foreignKey:CreadoPorID;references:ID" json:"creado_por,omitempty"`
	NormaReparto        *NormaReparto  `gorm:"foreignKey:NormaRepartoID;references:ID" json:"norma_reparto,omitempty"`
	DeletedAt           gorm.DeletedAt `gorm:"index" json:"-"`
}

func (RadicadoNormaReparto) TableName() string { return "radicado_norma_reparto" }

