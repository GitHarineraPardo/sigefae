package db

import "gorm.io/gorm"

// ---------------------------------------------------------------------------
// Receptor
// ---------------------------------------------------------------------------

type Receptor struct {
	ID              uint           `gorm:"primaryKey;column:id" json:"id"`
	Nombre          string         `gorm:"column:nombre;type:varchar(255)" json:"nombre"`
	TipoDocumentoID uint           `gorm:"column:tipo_documento_id;index:idx_receptor_tipo_documento_id" json:"tipo_documento_id"`
	TipoDocumento   *TipoDocumento `gorm:"foreignKey:TipoDocumentoID;references:ID" json:"tipo_documento,omitempty"`
	NumeroDocumento string         `gorm:"column:numero_documento;type:varchar(50);uniqueIndex:uk_receptorRFrespo_documento" json:"numero_documento"`
	Activo          bool           `gorm:"column:activo;default:true" json:"activo"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Receptor) TableName() string { return "receptor" }

