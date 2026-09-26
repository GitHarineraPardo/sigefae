package db

import (
	"time"

	"gorm.io/gorm"
)

// ---------------------------------------------------------------------------
// Documento radicado (flujo de radicación / trazabilidad)
// ---------------------------------------------------------------------------

type DocumentoRadicado struct {
	ID                      uint                     `gorm:"primaryKey;column:id" json:"id"`
	DocumentoComercialID    uint                     `gorm:"column:documento_comercial_id;uniqueIndex:uk_radicado_documento" json:"documento_comercial_id"`
	DocumentoComercial      *DocumentoComercial      `gorm:"foreignKey:DocumentoComercialID;references:ID" json:"documento_comercial,omitempty"`
	TipoRadicacionID        uint                     `gorm:"column:tipo_radicacion_id;index:idx_radicado_tipo" json:"tipo_radicacion_id"`
	TipoRadicacion          *TipoRadicacion          `gorm:"foreignKey:TipoRadicacionID;references:ID" json:"tipo_radicacion,omitempty"`
	RutaID                  uint                     `gorm:"column:ruta_id;index:idx_radicado_ruta" json:"ruta_id"`
	Ruta                    *Ruta                    `gorm:"foreignKey:RutaID;references:ID" json:"ruta,omitempty"`
	NumeroRadicado          string                   `gorm:"column:numero_radicado;type:varchar(100);uniqueIndex:uk_numero_radicado" json:"numero_radicado"`
	FechaRadicacion         time.Time                `gorm:"column:fecha_radicacion;index:idx_radicado_fecha" json:"fecha_radicacion"`
	UsuarioActualID         *uint                    `gorm:"column:usuario_actual_id;index:idx_radicado_usuario" json:"usuario_actual_id,omitempty"`
	CargoActualID           *uint                    `gorm:"column:cargo_actual_id;index:idx_radicado_cargo_actual" json:"cargo_actual_id,omitempty"`
	CargoActual             *Cargo                   `gorm:"foreignKey:CargoActualID;references:ID" json:"cargo_actual,omitempty"`
	UsuarioActual           *Usuario                 `gorm:"foreignKey:UsuarioActualID;references:ID" json:"usuario_actual,omitempty"`
	EstadoPosesion          string                   `gorm:"column:estado_posesion;type:varchar(50)" json:"estado_posesion"` // "Libre", "Tomado", "EnProceso"
	PasoActualID            *uint                    `gorm:"column:paso_actual_id;index:idx_radicado_paso_actual" json:"paso_actual_id,omitempty"`
	PasoActual              *PasoRuta                `gorm:"foreignKey:PasoActualID;references:ID" json:"paso_actual,omitempty"`
	PasoPendienteRetornoID  *uint                    `gorm:"column:paso_pendiente_retorno_id;index:idx_radicado_paso_retorno" json:"paso_pendiente_retorno_id,omitempty"`
	PasoPendienteRetorno    *PasoRuta                `gorm:"foreignKey:PasoPendienteRetornoID;references:ID" json:"paso_pendiente_retorno,omitempty"`
	TareaPendienteRetornoID *uint                    `gorm:"column:tarea_pendiente_retorno_id;index:idx_radicado_tarea_retorno" json:"tarea_pendiente_retorno_id,omitempty"`
	TareaPendienteRetorno   *Tarea                   `gorm:"foreignKey:TareaPendienteRetornoID;references:ID" json:"tarea_pendiente_retorno,omitempty"`
	EstadoID                uint                     `gorm:"column:estado_id;index:idx_radicado_estado" json:"estado_id"`
	Estado                  *EstadoDocumentoRadicado `gorm:"foreignKey:EstadoID;references:ID" json:"estado,omitempty"`
	UltimaActividad         *time.Time               `gorm:"column:ultima_actividad" json:"ultima_actividad"` // Puntero para evitar problemas de fecha cero
	QrID                    uint                     `gorm:"column:qr_id;index:idx_radicado_qr" json:"qr_id"`
	Qr                      *CodigoQr                `gorm:"foreignKey:QrID;references:ID" json:"qr,omitempty"`
	Causado                 bool                     `gorm:"column:causado;default:false" json:"causado"`
	FechaCausacion          *time.Time               `gorm:"column:fecha_causacion" json:"fecha_causacion"`
	NumeroEgreso            string                   `gorm:"column:numero_egreso;type:varchar(100)" json:"numero_egreso"`
	Pagado                  bool                     `gorm:"column:pagado;default:false" json:"pagado"`
	FechaPago               *time.Time               `gorm:"column:fecha_pago" json:"fecha_pago"`
	ComprobantesSubidos     bool                     `gorm:"column:comprobantes_subidos;default:false" json:"comprobantes_subidos"`
	NormasReparto           []RadicadoNormaReparto   `gorm:"foreignKey:DocumentoRadicadoID" json:"normas_reparto,omitempty"`
	MetodoPagoID            uint                     `gorm:"column:metodo_pago_id;index:idx_radicado_metodo_pago" json:"metodo_pago_id"`
	MetodoPago              *MetodoPago              `gorm:"foreignKey:MetodoPagoID;references:ID" json:"metodo_pago,omitempty"`
	Archivos                []Archivo                `gorm:"foreignKey:DocumentoRadicadoID" json:"archivos,omitempty"` // Los archivos pertenecen al Expediente!
	Comentarios             []Comentario             `gorm:"foreignKey:DocumentoRadicadoID" json:"comentarios,omitempty"`
	Tareas                  []Tarea                  `gorm:"foreignKey:DocumentoRadicadoID" json:"tareas,omitempty"`
	Trazabilidades          []Trazabilidad           `gorm:"foreignKey:DocumentoRadicadoID" json:"trazabilidades,omitempty"`
	Aprobaciones            []RegistroAprobacion     `gorm:"foreignKey:DocumentoRadicadoID" json:"aprobaciones,omitempty"`
	UpdatedAt               time.Time                `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt               gorm.DeletedAt           `gorm:"index" json:"-"`
}

func (DocumentoRadicado) TableName() string { return "documento_radicado" }

type CodigoQr struct {
	ID        uint           `gorm:"primaryKey;column:id" json:"id"`
	Url       string         `gorm:"column:url;type:varchar(500)" json:"url"`
	Activo    bool           `gorm:"column:activo;default:true" json:"activo"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (CodigoQr) TableName() string { return "codigo_qr" }

type Comentario struct {
	ID                  uint               `gorm:"primaryKey;column:id" json:"id"`
	DocumentoRadicadoID uint               `gorm:"column:documento_radicado_id;index:idx_comentario_documento" json:"documento_radicado_id"`
	DocumentoRadicado   *DocumentoRadicado `gorm:"foreignKey:DocumentoRadicadoID;references:ID" json:"documento_radicado,omitempty"`
	Descripcion         string             `gorm:"column:descripcion;type:text" json:"descripcion"`
	UsuarioID           uint               `gorm:"column:usuario_id;index:idx_comentario_usuario" json:"usuario_id"`
	Usuario             *Usuario           `gorm:"foreignKey:UsuarioID;references:ID" json:"usuario,omitempty"`
	Fecha               time.Time          `gorm:"column:fecha;index:idx_comentario_fecha" json:"fecha"`
	DeletedAt           gorm.DeletedAt     `gorm:"index" json:"-"`
}

func (Comentario) TableName() string { return "comentario" }

type Trazabilidad struct {
	ID                  uint               `gorm:"primaryKey;column:id" json:"id"`
	DocumentoRadicadoID uint               `gorm:"column:documento_radicado_id;index:idx_trazabilidad_documento" json:"documento_radicado_id"`
	DocumentoRadicado   *DocumentoRadicado `gorm:"foreignKey:DocumentoRadicadoID;references:ID" json:"documento_radicado,omitempty"`
	UsuarioID           uint               `gorm:"column:usuario_id;index:idx_trazabilidad_usuario" json:"usuario_id"`
	Usuario             *Usuario           `gorm:"foreignKey:UsuarioID;references:ID" json:"usuario,omitempty"`
	Accion              string             `gorm:"column:accion;type:varchar(255)" json:"accion"`
	Descripcion         string             `gorm:"column:descripcion;type:text" json:"descripcion"`
	Fecha               time.Time          `gorm:"column:fecha;index:idx_trazabilidad_fecha" json:"fecha"`
	DeletedAt           gorm.DeletedAt     `gorm:"index" json:"-"`
}

func (Trazabilidad) TableName() string { return "trazabilidad" }

