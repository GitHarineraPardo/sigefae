package auth

import (
	"net/http"

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
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
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

	user, token, err := h.service.Login(
		request.Email,
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
	Identifier string `json:"identifier" binding:"required"`
	Secret     string `json:"secret" binding:"required"`
}

func (h *Handler) SSO(c *gin.Context) {
	var request SSORequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Secret validation (you can change this string or move to env vars)
	if request.Secret != "SIGEFAE_INTERNAL_SSO_SECRET_2026" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid secret"})
		return
	}

	user, token, err := h.service.SSOLogin(request.Identifier)
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
