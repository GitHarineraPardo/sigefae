package paso_ruta

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
	// Validar Ruta
	// ==========================

	var ruta db.Ruta

	err := s.db.
		First(&ruta, req.RutaID).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("la ruta no existe")
	}

	if err != nil {
		return nil, err
	}

	// ==========================
	// Validar Usuario
	// ==========================

	var usuario db.Usuario

	err = s.db.
		First(&usuario, req.UsuarioID).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("el usuario no existe")
	}

	if err != nil {
		return nil, err
	}

	// ==========================
	// Validar Orden
	// ==========================

	var existing db.PasoRuta

	err = s.db.
		Where("ruta_id = ? AND orden = ?", req.RutaID, req.Orden).
		First(&existing).Error

	if err == nil {
		return nil, errors.New("ya existe un paso con ese orden")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// ==========================
	// Crear Paso
	// ==========================

	paso := db.PasoRuta{
		RutaID:    req.RutaID,
		Orden:     req.Orden,
		Nombre:    req.Nombre,
		UsuarioID: req.UsuarioID,
		Activo:    true,
	}

	if err := s.db.Create(&paso).Error; err != nil {
		return nil, err
	}

	// ==========================
	// Cargar relaciones
	// ==========================

	if err := s.db.
		Preload("Ruta").
		Preload("Ruta.Area").
		Preload("Usuario").
		First(&paso, paso.ID).Error; err != nil {

		return nil, err
	}

	response := toResponse(paso)

	return &response, nil
}

func (s *Service) List() ([]Response, error) {

	var pasos []db.PasoRuta

	if err := s.db.
		Preload("Ruta").
		Preload("Ruta.Area").
		Preload("Usuario").
		Order("ruta_id ASC").
		Order("orden ASC").
		Find(&pasos).Error; err != nil {

		return nil, err
	}

	response := make([]Response, 0, len(pasos))

	for _, paso := range pasos {
		response = append(response, toResponse(paso))
	}

	return response, nil
}

func (s *Service) UpdateStatus(id uint, activo bool) error {

	var paso db.PasoRuta

	err := s.db.
		First(&paso, id).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.New("paso no encontrado")
	}

	if err != nil {
		return err
	}

	paso.Activo = activo

	return s.db.Save(&paso).Error
}
func (s *Service) Update(id uint, req UpdateRequest) (*Response, error) {

	// ==========================
	// Buscar paso (SIN PRELOAD)
	// ==========================

	var paso db.PasoRuta

	err := s.db.First(&paso, id).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("paso no encontrado")
	}

	if err != nil {
		return nil, err
	}

	// ==========================
	// Validar Ruta
	// ==========================

	var ruta db.Ruta

	err = s.db.First(&ruta, req.RutaID).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("la ruta no existe")
	}

	if err != nil {
		return nil, err
	}

	// ==========================
	// Validar Usuario
	// ==========================

	var usuario db.Usuario

	err = s.db.First(&usuario, req.UsuarioID).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("el usuario no existe")
	}

	if err != nil {
		return nil, err
	}

	// ==========================
	// Validar orden repetido
	// ==========================

	var existing db.PasoRuta

	err = s.db.
		Where(
			"ruta_id = ? AND orden = ? AND id <> ?",
			req.RutaID,
			req.Orden,
			id,
		).
		First(&existing).Error

	if err == nil {
		return nil, errors.New("ya existe un paso con ese orden en la ruta")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// ==========================
	// Actualizar
	// ==========================

	if err := s.db.Model(&paso).Updates(map[string]any{
		"ruta_id":    req.RutaID,
		"orden":      req.Orden,
		"nombre":     req.Nombre,
		"usuario_id": req.UsuarioID,
	}).Error; err != nil {

		return nil, err
	}

	// ==========================
	// Recargar relaciones
	// ==========================

	if err := s.db.
		Preload("Ruta").
		Preload("Ruta.Area").
		Preload("Usuario").
		First(&paso, paso.ID).Error; err != nil {

		return nil, err
	}

	response := toResponse(paso)

	return &response, nil
}
