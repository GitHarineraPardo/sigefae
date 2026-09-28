package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"sigefae/internal/db"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {

	return &Handler{
		service: service,
	}
}

type LoginRequest struct {
	Email     string `json:"email" binding:"omitempty,email"`
	IDExterno string `json:"id_externo"`
	Password  string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token   string     `json:"token"`
	Usuario UsuarioDTO `json:"usuario"`
}

type UsuarioDTO struct {
	ID      uint   `json:"id"`
	Nombre  string `json:"nombre"`
	Email   string `json:"email"`
	CargoID *uint  `json:"cargo_id,omitempty"`
	Cargo   string `json:"cargo"`
	Rol     string `json:"rol"`
}

func (h *Handler) Login(c *gin.Context) {

	var request LoginRequest

	if err := c.ShouldBindJSON(&request); err != nil {

		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})

		return
	}
	if request.Email == "" && request.IDExterno == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email o id_externo es requerido"})
		return
	}

	user, token, err := h.service.Login(
		request.Email,
		request.IDExterno,
		request.Password,
	)

	if err != nil {

		c.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})

		return
	}

	response := LoginResponse{
		Token: token,
		Usuario: UsuarioDTO{
			ID:      user.ID,
			Nombre:  user.Nombre,
			Email:   user.Email,
			CargoID: user.CargoID,
			Cargo:   cargoNombre(user.Cargo),
		},
	}

	if user.Rol != nil {
		response.Usuario.Rol = user.Rol.Nombre
	}

	c.JSON(http.StatusOK, response)
}

type SSORequest struct {
	IDExterno  externalID `json:"id_externo"`
	Identifier externalID `json:"identifier"`
	Secret     string     `json:"secret" binding:"required"`
}

type externalID string

func (id *externalID) UnmarshalJSON(data []byte) error {
	var value string
	if err := json.Unmarshal(data, &value); err == nil {
		*id = externalID(value)
		return nil
	}

	var number json.Number
	if err := json.Unmarshal(data, &number); err != nil {
		return fmt.Errorf("el ID externo debe ser texto o número")
	}
	*id = externalID(number.String())
	return nil
}

func (h *Handler) SSO(c *gin.Context) {
	var request SSORequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	idExterno := string(request.IDExterno)
	if strings.TrimSpace(idExterno) == "" {
		idExterno = string(request.Identifier)
	}
	if strings.TrimSpace(idExterno) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id_externo es requerido"})
		return
	}

	// Secret validation (you can change this string or move to env vars)
	if request.Secret != "SIGEFAE_INTERNAL_SSO_SECRET_2026" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid secret"})
		return
	}

	user, token, err := h.service.SSOLogin(idExterno)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	response := LoginResponse{
		Token: token,
		Usuario: UsuarioDTO{
			ID:      user.ID,
			Nombre:  user.Nombre,
			Email:   user.Email,
			CargoID: user.CargoID,
			Cargo:   cargoNombre(user.Cargo),
		},
	}

	if user.Rol != nil {
		response.Usuario.Rol = user.Rol.Nombre
	}

	c.JSON(http.StatusOK, response)
}

func cargoNombre(cargo *db.Cargo) string {
	if cargo == nil {
		return ""
	}
	return cargo.Nombre
}
