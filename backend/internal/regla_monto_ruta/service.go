package regla_monto_ruta

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

func (s *Service) List() ([]db.ReglaMontoRuta, error) {
	var items []db.ReglaMontoRuta
	err := s.db.Preload("Area").Preload("Ruta").Preload("UsuarioAprobador").Preload("CargoAprobador").Preload("RolAprobador").
		Where("activo = ?", true).Order("monto_minimo_smmlv ASC").Find(&items).Error
	return items, err
}

func (s *Service) Create(item *db.ReglaMontoRuta) error {
	return s.db.Create(item).Error
}

func (s *Service) Update(id uint, item *db.ReglaMontoRuta) error {
	return s.db.Model(&db.ReglaMontoRuta{}).Where("id = ?", id).Updates(item).Error
}

func (s *Service) Delete(id uint) error {
	return s.db.Model(&db.ReglaMontoRuta{}).Where("id = ?", id).Update("activo", false).Error
}

func (s *Service) UpdateStatus(id uint, activo bool) error {
	return s.db.Model(&db.ReglaMontoRuta{}).Where("id = ?", id).Update("activo", activo).Error
}
