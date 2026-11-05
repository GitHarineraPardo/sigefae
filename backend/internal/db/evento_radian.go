package db

import "time"

type EventoRadian struct {
	ID                   uint                `gorm:"primaryKey;column:id" json:"id"`
	DocumentoComercialID uint                `gorm:"column:documento_comercial_id;index:idx_evento_doc" json:"documento_comercial_id"`
	DocumentoComercial   *DocumentoComercial `gorm:"foreignKey:DocumentoComercialID;references:ID" json:"documento_comercial,omitempty"`
	TipoEvento           string              `gorm:"column:tipo_evento;type:varchar(50);index:idx_evento_tipo" json:"tipo_evento"` // ACKNOWLEDGED, RECEIVEDGOODS, ACCEPTED, REJECTED
	CodigoDIAN           string              `gorm:"column:codigo_dian;type:varchar(10)" json:"codigo_dian"`                       // ej: 030, 032, 033, 034
	FechaEnvio           time.Time           `gorm:"column:fecha_envio" json:"fecha_envio"`
	Exitoso              bool                `gorm:"column:exitoso;default:true" json:"exitoso"`
	MensajeRespuesta     string              `gorm:"column:mensaje_respuesta;type:text" json:"mensaje_respuesta"`
	UsuarioID            *uint               `gorm:"column:usuario_id" json:"usuario_id,omitempty"`
	Usuario              *Usuario            `gorm:"foreignKey:UsuarioID;references:ID" json:"usuario,omitempty"`
	Observacion          string              `gorm:"column:observacion;type:varchar(500)" json:"observacion,omitempty"`
	CreatedAt            time.Time           `gorm:"column:created_at" json:"created_at"`
}

func (EventoRadian) TableName() string { return "evento_radian" }
