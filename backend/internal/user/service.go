package user

import (
	"errors"

	"gorm.io/gorm"

	"sigefae/internal/auth"
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
	// Validar correo
	// ==========================

	var existing db.Usuario

	err := s.db.
		Where("email = ?", req.Email).
		First(&existing).Error

	if err == nil {
		return nil, errors.New("el correo ya existe")
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// ==========================
	// Validar rol
	// ==========================

	var role db.Rol

	err = s.db.
		First(&role, req.RolID).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("el rol no existe")
	}

	if err != nil {
		return nil, err
	}

	var cargo db.Cargo
	if err := s.db.Where("activo = ?", true).First(&cargo, req.CargoID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("el cargo no existe o está inactivo")
		}
		return nil, err
	}

	// ==========================
	// Hash contraseña
	// ==========================

	hash, err := auth.Hash(req.Contrasena)

	if err != nil {
		return nil, err
	}

	// ==========================
	// Crear usuario
	// ==========================

	user := db.Usuario{
		Nombre:         req.Nombre,
		Email:          req.Email,
		HashContrasena: hash,
		CargoID:        &req.CargoID,
		RolID:          req.RolID,
	}

	if err := s.db.Create(&user).Error; err != nil {
		return nil, err
	}

	// ==========================
	// Cargar relaciones
	// ==========================

	if err := s.db.
		Preload("Rol").
		Preload("Cargo").
		First(&user, user.ID).Error; err != nil {
		return nil, err
	}

	// ==========================
	// Respuesta
	// ==========================

	response := toResponse(user)

	return &response, nil
}

func (s *Service) List() ([]Response, error) {

	var users []db.Usuario

	if err := s.db.
		Preload("Rol").
		Preload("Cargo").
		Order("nombre ASC").
		Find(&users).Error; err != nil {

		return nil, err
	}

	response := make([]Response, 0, len(users))

	for _, user := range users {
		response = append(response, toResponse(user))
	}

	return response, nil
}

func (s *Service) GetByID(id uint) (*Response, error) {

	var user db.Usuario

	err := s.db.
		Preload("Rol").
		Preload("Cargo").
		First(&user, id).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("usuario no encontrado")
	}

	if err != nil {
		return nil, err
	}

	response := toResponse(user)

	return &response, nil
}

func (s *Service) Update(id uint, req UpdateRequest) (*Response, error) {

	var user db.Usuario

	err := s.db.
		First(&user, id).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("usuario no encontrado")
	}

	if err != nil {
		return nil, err
	}

	// ==========================
	// Validar correo repetido
	// ==========================

	var existing db.Usuario

	err = s.db.
		Where("email = ? AND id <> ?", req.Email, id).
		First(&existing).Error

	if err == nil {
		return nil, errors.New("el correo ya existe")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// ==========================
	// Validar rol
	// ==========================

	var role db.Rol

	err = s.db.
		First(&role, req.RolID).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("el rol no existe")
	}

	if err != nil {
		return nil, err
	}

	var cargo db.Cargo
	if err := s.db.Where("activo = ?", true).First(&cargo, req.CargoID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("el cargo no existe o está inactivo")
		}
		return nil, err
	}

	user.Nombre = req.Nombre
	user.Email = req.Email
	user.CargoID = &req.CargoID
	user.RolID = req.RolID

	if err := s.db.Save(&user).Error; err != nil {
		return nil, err
	}

	if err := s.db.
		Preload("Rol").
		Preload("Cargo").
		First(&user, user.ID).Error; err != nil {
		return nil, err
	}

	response := toResponse(user)

	return &response, nil
}

func (s *Service) UpdateStatus(id uint, activo bool) error {

	var user db.Usuario

	err := s.db.
		First(&user, id).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.New("usuario no encontrado")
	}

	if err != nil {
		return err
	}

	user.Activo = activo

	return s.db.Save(&user).Error
}

func (s *Service) UpdatePassword(id uint, password string) error {

	var user db.Usuario

	err := s.db.
		First(&user, id).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.New("usuario no encontrado")
	}

	if err != nil {
		return err
	}

	hash, err := auth.Hash(password)

	if err != nil {
		return err
	}

	user.HashContrasena = hash

	return s.db.Save(&user).Error
}
