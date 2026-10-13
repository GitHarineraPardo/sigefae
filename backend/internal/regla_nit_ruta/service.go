package regla_nit_ruta

import (
	"errors"
	"strings"

	"sigefae/internal/db"

	"gorm.io/gorm"
)

type Service struct {
	db *gorm.DB
}

func New(database *gorm.DB) *Service {
	return &Service{db: database}
}

func NormalizarNIT(nit string) string {
	var normalizado strings.Builder
	for _, caracter := range nit {
		if caracter >= '0' && caracter <= '9' {
			normalizado.WriteRune(caracter)
		}
	}
	return normalizado.String()
}

func (s *Service) List() ([]db.ReglaNITRuta, error) {
	var items []db.ReglaNITRuta
	err := s.db.Preload("CargoAprobador").
		Where("activo = ?", true).Order("prioridad ASC, id ASC").Find(&items).Error
	return items, err
}

func (s *Service) Create(item *db.ReglaNITRuta) error {
	if err := validarRegla(item); err != nil {
		return err
	}
	return s.db.Create(item).Error
}

func (s *Service) Update(id uint, item *db.ReglaNITRuta) error {
	if err := validarRegla(item); err != nil {
		return err
	}
	return s.db.Model(&db.ReglaNITRuta{}).Where("id = ?", id).
		Select("nit", "cargo_aprobador_id", "posicion_insercion", "prioridad").
		Updates(item).Error
}

func validarRegla(item *db.ReglaNITRuta) error {
	item.NIT = NormalizarNIT(item.NIT)
	if item.NIT == "" {
		return errors.New("el NIT debe contener al menos un dígito")
	}
	if item.CargoAprobadorID == nil {
		return errors.New("debe seleccionar un cargo aprobador")
	}
	if item.Prioridad < 0 {
		return errors.New("la prioridad no puede ser negativa")
	}
	switch item.PosicionInsercion {
	case "PRIMERO", "ANTES_FINAL", "ULTIMO":
		return nil
	default:
		return errors.New("la posición de inserción no es válida")
	}
}

func (s *Service) Delete(id uint) error {
	return s.db.Model(&db.ReglaNITRuta{}).Where("id = ?", id).Update("activo", false).Error
}

func (s *Service) UpdateStatus(id uint, activo bool) error {
	return s.db.Model(&db.ReglaNITRuta{}).Where("id = ?", id).Update("activo", activo).Error
}
