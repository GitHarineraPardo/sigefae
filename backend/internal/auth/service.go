package auth

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

	return &Service{
		db: database,
	}
}
func (s *Service) Login(email, idExterno, password string) (*db.Usuario, string, error) {

	var user db.Usuario

	query := s.db.
		Preload("Rol").
		Preload("Cargo")
	if strings.TrimSpace(idExterno) != "" {
		err := query.Where("id_externo = ?", strings.TrimSpace(idExterno)).First(&user).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, "", errors.New("correo o contraseña incorrectos")
			}
			return nil, "", err
		}
	} else {
		var users []db.Usuario
		if err := query.Where("email = ?", email).Find(&users).Error; err != nil {
			return nil, "", err
		}
		if len(users) == 0 {
			return nil, "", errors.New("correo o contraseña incorrectos")
		}
		if len(users) > 1 {
			return nil, "", errors.New("el correo identifica varias cuentas; inicie sesión con su ID externo")
		}
		user = users[0]
	}

	if !user.Activo {
		return nil, "", errors.New("usuario inactivo")
	}

	if err := Check(user.HashContrasena, password); err != nil {
		return nil, "", errors.New("correo o contraseña incorrectos")
	}

	token, err := GenerateToken(
		user.ID,
		user.RolID,
	)

	if err != nil {
		return nil, "", err
	}

	return &user, token, nil
}

func (s *Service) SSOLogin(idExterno string) (*db.Usuario, string, error) {
	var user db.Usuario

	err := s.db.
		Preload("Rol").
		Preload("Cargo").
		Where("id_externo = ?", strings.TrimSpace(idExterno)).
		First(&user).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, "", errors.New("usuario no encontrado")
		}
		return nil, "", err
	}

	if !user.Activo {
		return nil, "", errors.New("usuario inactivo")
	}

	token, err := GenerateToken(
		user.ID,
		user.RolID,
	)

	if err != nil {
		return nil, "", err
	}

	return &user, token, nil
}
