package area

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

	var existing db.Area

	err := s.db.
		Where("nombre = ?", req.Nombre).
		First(&existing).Error

	if err == nil {
		return nil, errors.New("el área ya existe")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	area := db.Area{
		Nombre: req.Nombre,
		Activo: true,
	}

	if err := s.db.Create(&area).Error; err != nil {
		return nil, err
	}

	response := toResponse(area)

	return &response, nil
}

func (s *Service) List() ([]Response, error) {

	var areas []db.Area

	if err := s.db.
		Order("nombre ASC").
		Find(&areas).Error; err != nil {

		return nil, err
	}

	response := make([]Response, 0, len(areas))

	for _, area := range areas {
		response = append(response, toResponse(area))
	}

	return response, nil
}

func (s *Service) UpdateStatus(id uint, activo bool) error {

	var area db.Area

	err := s.db.
		First(&area, id).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.New("área no encontrada")
	}

	if err != nil {
		return err
	}

	area.Activo = activo

	return s.db.Save(&area).Error
}

func (s *Service) Update(id uint, req UpdateRequest) (*Response, error) {

	var area db.Area

	err := s.db.First(&area, id).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("área no encontrada")
	}

	if err != nil {
		return nil, err
	}

	// ==========================
	// Validar nombre repetido
	// ==========================

	var existing db.Area

	err = s.db.
		Where("nombre = ? AND id <> ?", req.Nombre, id).
		First(&existing).Error

	if err == nil {
		return nil, errors.New("el área ya existe")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// ==========================
	// Actualizar
	// ==========================

	area.Nombre = req.Nombre

	if err := s.db.Save(&area).Error; err != nil {
		return nil, err
	}

	response := toResponse(area)

	return &response, nil
}

func (s *Service) Delete(id uint) error {
	return s.db.Delete(&db.Area{}, id).Error
}

