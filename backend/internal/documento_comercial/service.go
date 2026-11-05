package documento_comercial

import (
	"errors"

	"gorm.io/gorm"

	"sigefae/internal/db"
)

type Service struct {
	db *gorm.DB
}

func New(database *gorm.DB) *Service {

	return &Service{
		db: database,
	}
}
func (s *Service) Create(dto CreateDTO) (*Response, error) {

	// ==========================
	// Validar proveedor
	// ==========================

	var proveedor db.Proveedor

	err := s.db.First(&proveedor, dto.IDProveedor).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("el proveedor no existe")
	}

	if err != nil {
		return nil, err
	}

	// ==========================
	// Validar receptor
	// ==========================

	var receptor db.Receptor

	err = s.db.First(&receptor, dto.IDReceptor).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("el receptor no existe")
	}

	if err != nil {
		return nil, err
	}

	// ==========================
	// Validar área
	// ==========================

	var area db.Area

	err = s.db.First(&area, dto.IDArea).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("el área no existe")
	}

	if err != nil {
		return nil, err
	}

	// ==========================
	// Validar moneda
	// ==========================

	var moneda db.Moneda

	err = s.db.First(&moneda, dto.MonedaID).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("la moneda no existe")
	}

	if err != nil {
		return nil, err
	}

	// ==========================
	// Validar tipo factura (opcional)
	// ==========================

	if dto.TipoFacturaID != nil {

		var tipoFactura db.TipoFactura

		err = s.db.First(&tipoFactura, *dto.TipoFacturaID).Error

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("el tipo de factura no existe")
		}

		if err != nil {
			return nil, err
		}
	}

	// ==========================
	// Validar proveedor + documento (duplicado)
	// ==========================

	var existing db.DocumentoComercial

	err = s.db.
		Where(
			"id_proveedor = ? AND numero_documento = ?",
			dto.IDProveedor,
			dto.NumeroDocumento,
		).
		First(&existing).Error

	if err == nil {
		return nil, errors.New("el documento comercial ya existe para este proveedor")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// ==========================
	// Validar CUFE
	// ==========================

	if dto.Cufe != nil {

		var existingCUFE db.DocumentoComercial

		err = s.db.
			Where("cufe = ?", *dto.Cufe).
			First(&existingCUFE).Error

		if err == nil {
			return nil, errors.New("el CUFE ya existe")
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
	}

	// ==========================
	// Crear documento + detalles en transacción
	// ==========================

	tx := s.db.Begin()

	documento := db.DocumentoComercial{
		Tipo:                     dto.Tipo,
		NumeroDocumento:          dto.NumeroDocumento,
		OrdenCompra:              dto.OrdenCompra,
		IDProveedor:              dto.IDProveedor,
		IDReceptor:               dto.IDReceptor,
		IDArea:                   dto.IDArea,
		TipoFacturaID:            dto.TipoFacturaID,
		Asunto:                   dto.Asunto,
		FechaDocumento:           dto.FechaDocumento,
		FechaVencimiento:         dto.FechaVencimiento,
		MonedaID:                 dto.MonedaID,
		Subtotal:                 dto.Subtotal,
		Iva:                      dto.Iva,
		Total:                    dto.Total,
		NumeroFolios:             dto.NumeroFolios,
		OrientacionSelloRecibido: dto.OrientacionSelloRecibido,
		Cufe:                     dto.Cufe,
		Activo:                   dto.Activo,
	}

	if err := tx.Create(&documento).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// ── Crear detalles ──
	for i := range dto.Detalles {
		det := dto.Detalles[i]
		det.DocumentoComercialID = documento.ID
		det.Activo = true
		if err := tx.Create(&det).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	tx.Commit()

	// Recargar con relaciones
	if err := s.db.
		Preload("Proveedor").
		Preload("Receptor").
		Preload("Area").
		Preload("TipoFactura").
		Preload("Moneda").
		Preload("Detalles").
		First(&documento, documento.ID).Error; err != nil {

		return nil, err
	}

	response := toResponse(documento)

	return &response, nil
}
func (s *Service) List() ([]Response, error) {

	var documentos []db.DocumentoComercial

	if err := s.db.
		Preload("Proveedor").
		Preload("Receptor").
		Preload("Area").
		Preload("TipoFactura").
		Preload("Moneda").
		Order("created_at DESC").
		Find(&documentos).Error; err != nil {

		return nil, err
	}

	response := make([]Response, 0, len(documentos))

	for _, documento := range documentos {
		response = append(response, toResponse(documento))
	}

	return response, nil
}
func (s *Service) Update(id uint, dto UpdateDocumentoComercialDTO) (*db.DocumentoComercial, error) {
	var doc db.DocumentoComercial
	if err := s.db.First(&doc, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("el documento comercial no existe")
		}
		return nil, err
	}

	// ── Validar Área ──
	var area db.Area
	if err := s.db.First(&area, dto.IDArea).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("el área no existe")
		}
		return nil, err
	}

	// ── Actualizar SOLO campos administrativos ──
	doc.OrdenCompra = dto.OrdenCompra
	doc.IDArea = dto.IDArea
	doc.Asunto = dto.Asunto
	doc.FechaVencimiento = dto.FechaVencimiento
	doc.OrientacionSelloRecibido = dto.OrientacionSelloRecibido
	doc.NumeroFolios = dto.NumeroFolios

	if err := s.db.Save(&doc).Error; err != nil {
		return nil, err
	}

	// Recargar relaciones
	if err := s.db.
		Preload("Proveedor").
		Preload("Receptor").
		Preload("Area").
		Preload("Moneda").
		Preload("Detalles").
		First(&doc, doc.ID).Error; err != nil {
		return nil, err
	}

	return &doc, nil
}
func (s *Service) Delete(id uint) error {

	var documento db.DocumentoComercial

	if err := s.db.First(&documento, id).Error; err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("documento comercial no encontrado")
		}

		return err
	}

	documento.Activo = !documento.Activo

	return s.db.Save(&documento).Error
}

func (s *Service) GetByID(id uint) (*Response, error) {

	var documento db.DocumentoComercial

	if err := s.db.
		Preload("Proveedor").
		Preload("Receptor").
		Preload("Area").
		Preload("TipoFactura").
		Preload("Moneda").
		Preload("Detalles").
		First(&documento, id).Error; err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("el documento comercial no existe")
		}

		return nil, err
	}

	response := toResponse(documento)

	return &response, nil
}

// ListPendientes devuelve documentos comerciales que están "Procesados" (extraídos del XML)
// pero que aún NO han sido radicados. Estos son los que la persona debe revisar.
func (s *Service) ListPendientes() ([]Response, error) {

	var documentos []db.DocumentoComercial

	// Documentos que:
	// NO tienen un documento_radicado asociado
	subQuery := s.db.Table("documento_radicado").Select("documento_comercial_id")

	if err := s.db.
		Preload("Proveedor").
		Preload("Receptor").
		Preload("Area").
		Preload("TipoFactura").
		Preload("Moneda").
		Preload("Detalles").
		Where("id NOT IN (?)", subQuery).
		Where("activo = ?", true).
		Order("created_at DESC").
		Find(&documentos).Error; err != nil {

		return nil, err
	}

	response := make([]Response, 0, len(documentos))

	for _, documento := range documentos {
		response = append(response, toResponse(documento))
	}

	return response, nil
}
