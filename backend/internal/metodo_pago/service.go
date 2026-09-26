package metodo_pago

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

func (s *Service) Create(req CreateRequest) (*Response, error) {

	// ==========================
	// Validar TipoPago
	// ==========================

	var tipo db.TipoPago

	err := s.db.
		First(&tipo, req.TipoPagoID).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("el tipo de pago no existe")
	}

	if err != nil {
		return nil, err
	}

	// ==========================
	// Validar nombre repetido
	// ==========================

	var existing db.MetodoPago

	err = s.db.
		Where("tipo_pago_id = ? AND nombre = ?", req.TipoPagoID, req.Nombre).
		First(&existing).Error

	if err == nil {
		return nil, errors.New("el método de pago ya existe")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// ==========================
	// Crear
	// ==========================

	metodo := db.MetodoPago{
		TipoPagoID: req.TipoPagoID,
		Nombre:     req.Nombre,
		Activo:     true,
	}

	if err := s.db.Create(&metodo).Error; err != nil {
		return nil, err
	}

	if err := s.db.
		Preload("TipoPago").
		First(&metodo, metodo.ID).Error; err != nil {
		return nil, err
	}

	response := toResponse(metodo)

	return &response, nil
}

func (s *Service) List() ([]Response, error) {

	var metodos []db.MetodoPago

	if err := s.db.
		Preload("TipoPago").
		Order("nombre ASC").
		Find(&metodos).Error; err != nil {

		return nil, err
	}

	response := make([]Response, 0, len(metodos))

	for _, metodo := range metodos {
		response = append(response, toResponse(metodo))
	}

	return response, nil
}

func (s *Service) UpdateStatus(id uint, activo bool) error {

	var metodo db.MetodoPago

	err := s.db.
		First(&metodo, id).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.New("método de pago no encontrado")
	}

	if err != nil {
		return err
	}

	metodo.Activo = activo

	return s.db.Save(&metodo).Error
}

func (s *Service) Update(id uint, req UpdateRequest) (*Response, error) {

	var metodo db.MetodoPago

	// ==========================
	// Buscar método
	// ==========================

	err := s.db.
		First(&metodo, id).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("método de pago no encontrado")
	}

	if err != nil {
		return nil, err
	}

	// ==========================
	// Validar TipoPago
	// ==========================

	var tipo db.TipoPago

	err = s.db.
		First(&tipo, req.TipoPagoID).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("el tipo de pago no existe")
	}

	if err != nil {
		return nil, err
	}

	// ==========================
	// Validar nombre repetido
	// ==========================

	var existing db.MetodoPago

	err = s.db.
		Where(
			"nombre = ? AND tipo_pago_id = ? AND id <> ?",
			req.Nombre,
			req.TipoPagoID,
			id,
		).
		First(&existing).Error

	if err == nil {
		return nil, errors.New("el método de pago ya existe")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// ==========================
	// Actualizar
	// ==========================

	if err := s.db.
		Model(&db.MetodoPago{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"nombre":       req.Nombre,
			"tipo_pago_id": req.TipoPagoID,
		}).Error; err != nil {

		return nil, err
	}

	// ==========================
	// Recargar con relaciones
	// ==========================

	if err := s.db.
		Preload("TipoPago").
		First(&metodo, id).Error; err != nil {

		return nil, err
	}

	response := toResponse(metodo)

	return &response, nil
}

func (s *Service) Delete(id uint) error {
	return s.db.Delete(&db.MetodoPago{}, id).Error
}

