package db

import (
	"time"

	"gorm.io/gorm"
)

type ReglaNITRuta struct {
	ID                uint           `gorm:"primaryKey;column:id" json:"id"`
	NIT               string         `gorm:"column:nit;type:varchar(50);index:idx_regla_nit" json:"nit"`
	CargoAprobadorID  *uint          `gorm:"column:cargo_aprobador_id;index:idx_regla_nit_cargo" json:"cargo_aprobador_id,omitempty"`
	CargoAprobador    *Cargo         `gorm:"foreignKey:CargoAprobadorID;references:ID" json:"cargo_aprobador,omitempty"`
	PosicionInsercion string         `gorm:"column:posicion_insercion;type:varchar(20);default:'ULTIMO'" json:"posicion_insercion"`
	Prioridad         int            `gorm:"column:prioridad;default:0" json:"prioridad"`
	Activo            bool           `gorm:"column:activo;default:true" json:"activo"`
	CreatedAt         time.Time      `gorm:"column:created_at" json:"created_at"`
	UpdatedAt         time.Time      `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`
}

func (ReglaNITRuta) TableName() string { return "regla_nit_ruta" }
