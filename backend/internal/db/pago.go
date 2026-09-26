package db

import "gorm.io/gorm"

// ---------------------------------------------------------------------------
// Pago
// ---------------------------------------------------------------------------

type MetodoPago struct {
	ID         uint           `gorm:"primaryKey;column:id" json:"id"`
	TipoPagoID uint           `gorm:"column:tipo_pago_id;index:idx_metodo_pago_tipo_pago" json:"tipo_pago_id"`
	TipoPago   *TipoPago      `gorm:"foreignKey:TipoPagoID;references:ID" json:"tipo_pago,omitempty"`
	Nombre     string         `gorm:"column:nombre;type:varchar(255)" json:"nombre"`
	Activo     bool           `gorm:"column:activo;default:true" json:"activo"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

func (MetodoPago) TableName() string { return "metodo_pago" }

