package ruta

import (
	"errors"
	"strings"

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

	nombre := strings.TrimSpace(req.Nombre)
	zona := strings.TrimSpace(req.Zona)
	if zona == "" {
		zona = "BUCARAMANGA"
	}

	var existing db.Ruta

	err := s.db.
		Where("LOWER(TRIM(nombre)) = LOWER(?) AND area_id = ? AND LOWER(TRIM(zona)) = LOWER(?)", nombre, req.AreaID, zona).
		First(&existing).Error

	if err == nil {
		return nil, errors.New("ya existe una ruta con ese nombre, área y zona")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	var area db.Area

	err = s.db.
		First(&area, req.AreaID).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("el área no existe")
	}

	if err != nil {
		return nil, err
	}

	ruta := db.Ruta{
		Nombre:  nombre,
		Zona:    zona,
		Version: 1,
		AreaID:  req.AreaID,
		Activo:  true,
	}

	if err := s.db.Create(&ruta).Error; err != nil {
		return nil, err
	}

	if err := s.db.
		Preload("Area").
		First(&ruta, ruta.ID).Error; err != nil {

		return nil, err
	}

	response := toResponse(ruta)

	return &response, nil
}
func (s *Service) List() ([]Response, error) {

	var rutas []db.Ruta

	if err := s.db.
		Preload("Area").
		Order("nombre ASC").
		Find(&rutas).Error; err != nil {

		return nil, err
	}

	response := make([]Response, 0, len(rutas))

	for _, ruta := range rutas {
		response = append(response, toResponse(ruta))
	}

	return response, nil
}
func (s *Service) UpdateStatus(id uint, activo bool) error {

	var ruta db.Ruta

	err := s.db.
		First(&ruta, id).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.New("ruta no encontrada")
	}

	if err != nil {
		return err
	}

	ruta.Activo = activo

	return s.db.Save(&ruta).Error
}
func (s *Service) Update(id uint, req UpdateRequest) (*Response, error) {

	var ruta db.Ruta

	// ==========================
	// Validar que exista la ruta
	// ==========================

	err := s.db.First(&ruta, id).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("ruta no encontrada")
	}

	if err != nil {
		return nil, err
	}

	// ==========================
	// Validar área
	// ==========================

	var area db.Area

	err = s.db.First(&area, req.AreaID).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("el área no existe")
	}

	if err != nil {
		return nil, err
	}

	// ==========================
	// Actualizar zona si se envía
	// ==========================

	zonaUpdate := strings.TrimSpace(req.Zona)
	if zonaUpdate == "" {
		zonaUpdate = ruta.Zona // mantener zona actual si no se cambia
	}
	nombreUpdate := strings.TrimSpace(req.Nombre)

	// ==========================
	// Validar nombre repetido (nombre+area+zona)
	// ==========================

	var existing db.Ruta

	err = s.db.
		Where(
			"LOWER(TRIM(nombre)) = LOWER(?) AND area_id = ? AND LOWER(TRIM(zona)) = LOWER(?) AND id <> ?",
			nombreUpdate,
			req.AreaID,
			zonaUpdate,
			id,
		).
		First(&existing).Error

	if err == nil {
		return nil, errors.New("ya existe una ruta con ese nombre, área y zona")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// ==========================
	// Actualizar
	// ==========================

	updatesMap := map[string]any{
		"nombre":  nombreUpdate,
		"area_id": req.AreaID,
		"zona":    zonaUpdate,
	}

	err = s.db.
		Model(&db.Ruta{}).
		Where("id = ?", id).
		Updates(updatesMap).Error

	if err != nil {
		return nil, err
	}

	// ==========================
	// Obtener actualizado
	// ==========================

	err = s.db.
		Preload("Area").
		First(&ruta, id).Error

	if err != nil {
		return nil, err
	}

	response := toResponse(ruta)

	return &response, nil
}
