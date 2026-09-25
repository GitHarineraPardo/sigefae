package db

import "time"

// ActivoFijo representa una línea de distribución de activo fijo para un documento radicado.
// Actúa en conjunto con las normas de reparto: entre ambos deben sumar 100%.
type RadicadoActivoFijo struct {
	ID                  uint      `gorm:"primaryKey;column:id" json:"id"`
	DocumentoRadicadoID uint      `gorm:"column:documento_radicado_id;not null;index:idx_raf_radicado" json:"documento_radicado_id"`
	Nombre              string    `gorm:"column:nombre;type:varchar(150);not null" json:"nombre"`
	Sucursal            string    `gorm:"column:sucursal;type:varchar(50);not null" json:"sucursal"`
	Proyecto            string    `gorm:"column:proyecto;type:varchar(100);default:''" json:"proyecto"`
	Porcentaje          float64   `gorm:"column:porcentaje;type:decimal(5,2);not null" json:"porcentaje"`
	Descripcion         string    `gorm:"column:descripcion;type:text;default:''" json:"descripcion"`
	CreadoPorID         uint      `gorm:"column:creado_por_id;index:idx_raf_creado_por" json:"creado_por_id"`
	CreadoPor           *Usuario  `gorm:"foreignKey:CreadoPorID;references:ID" json:"creado_por,omitempty"`
	CreatedAt           time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

func (RadicadoActivoFijo) TableName() string { return "radicado_activo_fijo" }
