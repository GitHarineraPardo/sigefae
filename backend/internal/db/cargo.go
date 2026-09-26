package db

import (
	"time"

	"gorm.io/gorm"
)

type Cargo struct {
	ID        uint           `gorm:"primaryKey;column:id" json:"id"`
	Nombre    string         `gorm:"column:nombre;type:varchar(255);uniqueIndex" json:"nombre"`
	Activo    bool           `gorm:"column:activo;default:true" json:"activo"`
	CreatedAt time.Time      `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Cargo) TableName() string { return "cargo" }
