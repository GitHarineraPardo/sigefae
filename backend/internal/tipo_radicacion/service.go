package tipo_radicacion

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

	var existing db.TipoRadicacion

	err := s.db.
		Where("nombre = ?", req.Nombre).
		First(&existing).Error

	if err == nil {
		return nil, errors.New("el tipo de radicación ya existe")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	tipo := db.TipoRadicacion{
		Nombre: req.Nombre,
		Activo: true,
	}

	if err := s.db.Create(&tipo).Error; err != nil {
		return nil, err
	}

	response := toResponse(tipo)

	return &response, nil
}

func (s *Service) List() ([]Response, error) {

	var tipos []db.TipoRadicacion

	if err := s.db.
		Order("nombre ASC").
		Find(&tipos).Error; err != nil {

		return nil, err
	}

	response := make([]Response, 0, len(tipos))

	for _, tipo := range tipos {
		response = append(response, toResponse(tipo))
	}

	return response, nil
}

func (s *Service) UpdateStatus(id uint, activo bool) error {

	var tipo db.TipoRadicacion

	err := s.db.First(&tipo, id).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.New("tipo de radicación no encontrado")
	}

	if err != nil {
		return err
	}

	tipo.Activo = activo

	return s.db.Save(&tipo).Error
}

func (s *Service) Update(id uint, req UpdateRequest) (*Response, error) {

	var tipo db.TipoRadicacion

	err := s.db.First(&tipo, id).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("tipo de radicación no encontrado")
	}

	if err != nil {
		return nil, err
	}

	var existing db.TipoRadicacion

	err = s.db.
		Where("nombre = ? AND id <> ?", req.Nombre, id).
		First(&existing).Error

	if err == nil {
		return nil, errors.New("el tipo de radicación ya existe")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	if err := s.db.Model(&tipo).Updates(map[string]any{
		"nombre": req.Nombre,
	}).Error; err != nil {

		return nil, err
	}

	if err := s.db.First(&tipo, tipo.ID).Error; err != nil {
		return nil, err
	}

	response := toResponse(tipo)

	return &response, nil
}

func (s *Service) Delete(id uint) error {
	return s.db.Delete(&db.TipoRadicacion{}, id).Error
}

