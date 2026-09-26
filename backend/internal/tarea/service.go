package tarea

import (
	"errors"
	"time"

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

func (s *Service) validateCargo(cargoID *uint) error {
	if cargoID == nil {
		return nil
	}
	var cargo db.Cargo
	if err := s.db.Where("activo = ?", true).First(&cargo, *cargoID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("el cargo no existe o está inactivo")
		}
		return err
	}
	return nil
}

func (s *Service) Create(req CreateDTO) (*Response, error) {
	if err := s.validateCargo(req.CargoAsignadoID); err != nil {
		return nil, err
	}

	// ==========================
	// Validar Documento Radicado
	// ==========================

	var documento db.DocumentoRadicado

	err := s.db.First(
		&documento,
		req.DocumentoRadicadoID,
	).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("documento radicado no encontrado")
	}

	if err != nil {
		return nil, err
	}

	// ==========================
	// Validar Usuario si viene
	// ==========================

	var usuario db.Usuario

	if req.UsuarioAsignadoID != nil {
		err = s.db.First(
			&usuario,
			*req.UsuarioAsignadoID,
		).Error

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("usuario no encontrado")
		}

		if err != nil {
			return nil, err
		}
	}

	// ==========================
	// Validar Estado
	// ==========================

	var estado db.EstadoTarea

	err = s.db.First(
		&estado,
		req.EstadoID,
	).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("estado de tarea no encontrado")
	}

	if err != nil {
		return nil, err
	}

	// ==========================
	// Crear
	// ==========================

	tarea := db.Tarea{
		DocumentoRadicadoID: req.DocumentoRadicadoID,
		UsuarioAsignadoID:   req.UsuarioAsignadoID,
		CargoAsignadoID:     req.CargoAsignadoID,
		EstadoID:            req.EstadoID,
		Descripcion:         req.Descripcion,
		FechaAsignacion:     time.Now(),
		FechaLimite:         req.FechaLimite,
	}

	if err := s.db.Create(&tarea).Error; err != nil {
		return nil, err
	}
	if err := s.db.Preload("CargoAsignadoCargo").Preload("UsuarioAsignado").Preload("Estado").First(&tarea, tarea.ID).Error; err != nil {
		return nil, err
	}

	response := toResponse(tarea)

	return &response, nil
}

func (s *Service) List() ([]Response, error) {

	var tareas []db.Tarea

	if err := s.db.
		Preload("CargoAsignadoCargo").
		Preload("UsuarioAsignado").
		Preload("Estado").
		Order("created_at DESC").
		Find(&tareas).Error; err != nil {

		return nil, err
	}

	response := make([]Response, 0, len(tareas))

	for _, tarea := range tareas {
		response = append(response, toResponse(tarea))
	}

	return response, nil
}

func (s *Service) Update(id uint, req UpdateDTO) (*Response, error) {
	if err := s.validateCargo(req.CargoAsignadoID); err != nil {
		return nil, err
	}

	var tarea db.Tarea

	err := s.db.First(&tarea, id).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("tarea no encontrada")
	}

	if err != nil {
		return nil, err
	}

	// ==========================
	// Validar Usuario si viene
	// ==========================

	var usuario db.Usuario

	if req.UsuarioAsignadoID != nil {
		err = s.db.First(
			&usuario,
			*req.UsuarioAsignadoID,
		).Error

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("usuario no encontrado")
		}

		if err != nil {
			return nil, err
		}
	}

	// ==========================
	// Validar Estado
	// ==========================

	var estado db.EstadoTarea

	err = s.db.First(
		&estado,
		req.EstadoID,
	).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("estado de tarea no encontrado")
	}

	if err != nil {
		return nil, err
	}

	// ==========================
	// Actualizar
	// ==========================

	if err := s.db.Model(&tarea).Updates(map[string]any{
		"usuario_asignado_id": req.UsuarioAsignadoID,
		"cargo_asignado_id":   req.CargoAsignadoID,
		"estado_id":           req.EstadoID,
		"descripcion":         req.Descripcion,
		"fecha_inicio":        req.FechaInicio,
		"fecha_limite":        req.FechaLimite,
		"fecha_finalizacion":  req.FechaFinalizacion,
	}).Error; err != nil {

		return nil, err
	}

	if err := s.db.Preload("CargoAsignadoCargo").Preload("UsuarioAsignado").Preload("Estado").First(&tarea, tarea.ID).Error; err != nil {
		return nil, err
	}

	response := toResponse(tarea)

	return &response, nil
}
