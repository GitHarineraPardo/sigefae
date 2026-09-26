package proyecto

import (
	"sigefae/internal/db"

	"gorm.io/gorm"
)

type Service struct {
	db *gorm.DB
}

func New(database *gorm.DB) *Service {
	return &Service{db: database}
}

func (s *Service) Create(req Request) (Response, error) {
	model := db.Proyecto{
		Nombre: req.Nombre,
		Zona:   req.Zona,
		Activo: true,
	}
	if err := s.db.Create(&model).Error; err != nil {
		return Response{}, err
	}
	return s.mapToResponse(model), nil
}

func (s *Service) List() ([]Response, error) {
	var models []db.Proyecto
	if err := s.db.Order("nombre asc").Find(&models).Error; err != nil {
		return nil, err
	}
	var res []Response
	for _, m := range models {
		res = append(res, s.mapToResponse(m))
	}
	if len(res) == 0 {
		return []Response{}, nil
	}
	return res, nil
}

func (s *Service) UpdateStatus(id uint, activo bool) error {
	return s.db.Model(&db.Proyecto{}).Where("id = ?", id).Update("activo", activo).Error
}

func (s *Service) Update(id uint, req Request) (Response, error) {
	var model db.Proyecto
	if err := s.db.First(&model, id).Error; err != nil {
		return Response{}, err
	}
	model.Nombre = req.Nombre
	model.Zona = req.Zona
	if err := s.db.Save(&model).Error; err != nil {
		return Response{}, err
	}
	return s.mapToResponse(model), nil
}

func (s *Service) mapToResponse(model db.Proyecto) Response {
	return Response{
		ID:     model.ID,
		Nombre: model.Nombre,
		Zona:   model.Zona,
		Activo: model.Activo,
	}
}

func (s *Service) Delete(id uint) error {
	return s.db.Delete(&db.Proyecto{}, id).Error
}

