package cargo

import (
	"errors"
	"sigefae/internal/db"
	"strings"

	"gorm.io/gorm"
)

type Service struct {
	db *gorm.DB
}

func New(database *gorm.DB) *Service {
	return &Service{db: database}
}

func (s *Service) ListActivos() ([]CargoDTO, error) {
	var cargos []db.Cargo
	if err := s.db.Where("activo = ?", true).Order("nombre asc").Find(&cargos).Error; err != nil {
		return nil, err
	}
	return s.toDTOList(cargos), nil
}

func (s *Service) List() ([]CargoDTO, error) {
	var cargos []db.Cargo
	if err := s.db.Order("nombre asc").Find(&cargos).Error; err != nil {
		return nil, err
	}
	return s.toDTOList(cargos), nil
}

func (s *Service) Create(dto CreateDTO) (*CargoDTO, error) {
	nombreTrim := strings.TrimSpace(dto.Nombre)
	if nombreTrim == "" {
		return nil, errors.New("el nombre del cargo no puede estar vacío")
	}

	var cargo db.Cargo
	err := s.db.Unscoped().Where("nombre = ?", nombreTrim).First(&cargo).Error
	if err == nil {
		if !cargo.DeletedAt.Valid {
			return nil, errors.New("ya existe un cargo con este nombre")
		}
		cargo.DeletedAt = gorm.DeletedAt{}
		cargo.Activo = true
		if err := s.db.Unscoped().Save(&cargo).Error; err != nil {
			return nil, err
		}
		res := s.toDTO(cargo)
		return &res, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	cargo = db.Cargo{
		Nombre: nombreTrim,
		Activo: true,
	}

	if err := s.db.Create(&cargo).Error; err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") || strings.Contains(err.Error(), "unique constraint") {
			return nil, errors.New("ya existe un cargo con este nombre")
		}
		return nil, err
	}

	res := s.toDTO(cargo)
	return &res, nil
}

func (s *Service) Update(id uint, dto UpdateDTO) (*CargoDTO, error) {
	var cargo db.Cargo
	if err := s.db.First(&cargo, id).Error; err != nil {
		return nil, errors.New("cargo no encontrado")
	}

	if dto.Nombre != "" {
		cargo.Nombre = strings.TrimSpace(dto.Nombre)
		if cargo.Nombre == "" {
			return nil, errors.New("el nombre del cargo no puede estar vacío")
		}
	}
	if dto.Activo != nil {
		cargo.Activo = *dto.Activo
	}

	if err := s.db.Save(&cargo).Error; err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") || strings.Contains(err.Error(), "unique constraint") {
			return nil, errors.New("ya existe un cargo con este nombre")
		}
		return nil, err
	}

	res := s.toDTO(cargo)
	return &res, nil
}

func (s *Service) UpdateStatus(id uint, activo bool) error {
	return s.db.Model(&db.Cargo{}).Where("id = ?", id).Update("activo", activo).Error
}

func (s *Service) Delete(id uint) error {
	references := []struct {
		model any
		field string
	}{
		{model: &db.Usuario{}, field: "cargo_id"},
		{model: &db.PasoRuta{}, field: "cargo_id"},
		{model: &db.Tarea{}, field: "cargo_asignado_id"},
		{model: &db.DocumentoRadicado{}, field: "cargo_actual_id"},
		{model: &db.ReglaMontoRuta{}, field: "cargo_aprobador_id"},
	}
	for _, reference := range references {
		var count int64
		if err := s.db.Unscoped().Model(reference.model).Where(reference.field+" = ?", id).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return errors.New("no se puede eliminar un cargo en uso; desactívelo")
		}
	}
	return s.db.Delete(&db.Cargo{}, id).Error
}

func (s *Service) toDTO(c db.Cargo) CargoDTO {
	return CargoDTO{
		ID:        c.ID,
		Nombre:    c.Nombre,
		Activo:    c.Activo,
		CreatedAt: c.CreatedAt,
	}
}

func (s *Service) toDTOList(cargos []db.Cargo) []CargoDTO {
	var dtos []CargoDTO
	for _, c := range cargos {
		dtos = append(dtos, s.toDTO(c))
	}
	return dtos
}
