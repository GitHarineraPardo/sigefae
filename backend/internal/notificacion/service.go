package notificacion

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"

	"sigefae/internal/db"
)

type Service struct {
	db              *gorm.DB
	pardo360URL     string
	pardo360Project uint
	httpClient      *http.Client
}

func New(database *gorm.DB) *Service {
	projectID, _ := strconv.ParseUint(os.Getenv("PARDO360_NOTIFICACIONES_PROYECTO_ID"), 10, 64)
	return &Service{
		db:              database,
		pardo360URL:     strings.TrimSpace(os.Getenv("PARDO360_NOTIFICACIONES_URL")),
		pardo360Project: uint(projectID),
		httpClient:      &http.Client{Timeout: 8 * time.Second},
	}
}

func (s *Service) Create(dto CreateDTO) (*Response, error) {

	// ==========================
	// Validar usuario
	// ==========================

	var usuario db.Usuario

	err := s.db.First(&usuario, dto.UsuarioID).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("el usuario no existe")
	}

	if err != nil {
		return nil, err
	}

	// ==========================
	// Validar documento radicado (opcional)
	// ==========================

	if dto.DocumentoRadicadoID != nil {

		var documento db.DocumentoRadicado

		err = s.db.First(&documento, *dto.DocumentoRadicadoID).Error

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("el documento radicado no existe")
		}

		if err != nil {
			return nil, err
		}
	}

	// ==========================
	// Validar estado
	// ==========================

	switch dto.Estado {
	case "Pendiente", "Enviada", "Leida", "Expirada":
	default:
		return nil, errors.New("estado de notificación inválido")
	}

	// ==========================
	// Validar tipo
	// ==========================

	switch dto.Tipo {
	case "Recordatorio", "Asignacion", "Sistema", "Rechazo", "Finalizado", "Devolucion":
	default:
		return nil, errors.New("tipo de notificación inválido")
	}

	// ==========================
	// Crear
	// ==========================

	notificacion := db.Notificacion{
		UsuarioID:           dto.UsuarioID,
		DocumentoRadicadoID: dto.DocumentoRadicadoID,
		Mensaje:             dto.Mensaje,
		Estado:              dto.Estado,
		Tipo:                dto.Tipo,
		FechaCreacion:       dto.FechaCreacion,
		FechaEnvio:          dto.FechaEnvio,
		FechaLectura:        dto.FechaLectura,
	}

	if err := s.db.Create(&notificacion).Error; err != nil {
		return nil, err
	}

	if err := s.db.
		Preload("Usuario").
		Preload("DocumentoRadicado").
		First(&notificacion, notificacion.ID).Error; err != nil {

		return nil, err
	}

	response := toResponse(notificacion)

	return &response, nil
}
func (s *Service) List() ([]Response, error) {

	var notificaciones []db.Notificacion

	if err := s.db.
		Preload("Usuario").
		Preload("DocumentoRadicado").
		Order("fecha_creacion DESC").
		Find(&notificaciones).Error; err != nil {

		return nil, err
	}

	response := make([]Response, 0, len(notificaciones))

	for _, notificacion := range notificaciones {
		response = append(response, toResponse(notificacion))
	}

	return response, nil
}
func (s *Service) Update(id uint, dto UpdateDTO) (*Response, error) {

	// ==========================
	// Validar notificación
	// ==========================

	var notificacion db.Notificacion

	err := s.db.First(&notificacion, id).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("la notificación no existe")
	}

	if err != nil {
		return nil, err
	}

	// ==========================
	// Validar usuario
	// ==========================

	var usuario db.Usuario

	err = s.db.First(&usuario, dto.UsuarioID).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("el usuario no existe")
	}

	if err != nil {
		return nil, err
	}

	// ==========================
	// Validar documento radicado (opcional)
	// ==========================

	if dto.DocumentoRadicadoID != nil {

		var documento db.DocumentoRadicado

		err = s.db.First(&documento, *dto.DocumentoRadicadoID).Error

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("el documento radicado no existe")
		}

		if err != nil {
			return nil, err
		}
	}

	// ==========================
	// Validar estado
	// ==========================

	switch dto.Estado {
	case "Pendiente", "Enviada", "Leida", "Expirada":
	default:
		return nil, errors.New("estado de notificación inválido")
	}

	// ==========================
	// Validar tipo
	// ==========================

	switch dto.Tipo {
	case "Recordatorio", "Asignacion", "Sistema", "Rechazo", "Finalizado", "Devolucion":
	default:
		return nil, errors.New("tipo de notificación inválido")
	}

	// ==========================
	// Actualizar
	// ==========================

	if err := s.db.Model(&notificacion).Updates(map[string]any{
		"usuario_id":            dto.UsuarioID,
		"documento_radicado_id": dto.DocumentoRadicadoID,
		"mensaje":               dto.Mensaje,
		"estado":                dto.Estado,
		"tipo":                  dto.Tipo,
		"fecha_creacion":        dto.FechaCreacion,
		"fecha_envio":           dto.FechaEnvio,
		"fecha_lectura":         dto.FechaLectura,
	}).Error; err != nil {

		return nil, err
	}

	if err := s.db.
		Preload("Usuario").
		Preload("DocumentoRadicado").
		First(&notificacion, id).Error; err != nil {

		return nil, err
	}

	response := toResponse(notificacion)

	return &response, nil
}
func (s *Service) ListByUsuario(usuarioID uint) ([]Response, error) {
	var notificaciones []db.Notificacion
	if err := s.db.
		Preload("Usuario").
		Preload("DocumentoRadicado").
		Where("usuario_id = ?", usuarioID).
		Order("fecha_creacion DESC").
		Find(&notificaciones).Error; err != nil {
		return nil, err
	}
	response := make([]Response, 0, len(notificaciones))
	for _, n := range notificaciones {
		response = append(response, toResponse(n))
	}
	return response, nil
}

func (s *Service) MarkAsRead(id uint) (*Response, error) {
	var notificacion db.Notificacion
	if err := s.db.First(&notificacion, id).Error; err != nil {
		return nil, err
	}
	now := time.Now()
	notificacion.Estado = "Leida"
	notificacion.FechaLectura = &now
	if err := s.db.Save(&notificacion).Error; err != nil {
		return nil, err
	}
	s.db.Preload("Usuario").Preload("DocumentoRadicado").First(&notificacion, id)
	resp := toResponse(notificacion)
	return &resp, nil
}
func (s *Service) CreateFromEvent(dto CreateDTO) (*Response, error) {
	// Normalizar estado
	switch dto.Estado {
	case "Pendiente", "Enviada", "Leida", "Expirada":
	default:
		dto.Estado = "Pendiente"
	}
	// Normalizar tipo
	switch dto.Tipo {
	case "Recordatorio", "Asignacion", "Sistema", "Rechazo", "Finalizado", "Devolucion":
	default:
		dto.Tipo = "Sistema"
	}

	n := db.Notificacion{
		UsuarioID:           dto.UsuarioID,
		DocumentoRadicadoID: dto.DocumentoRadicadoID,
		Mensaje:             dto.Mensaje,
		Estado:              dto.Estado,
		Tipo:                dto.Tipo,
		FechaCreacion:       dto.FechaCreacion,
	}
	if err := s.db.Create(&n).Error; err != nil {
		return nil, err
	}

	s.sendToPardo360(dto)

	s.db.Preload("Usuario").Preload("DocumentoRadicado").First(&n, n.ID)
	resp := toResponse(n)
	return &resp, nil
}

type pardo360Notification struct {
	ProjectID      uint   `json:"proyecto_id"`
	RecipientID    uint   `json:"destinatario_id,omitempty"`
	RecipientEmail string `json:"destinatario_email,omitempty"`
	Heading        string `json:"encabezado"`
	Description    string `json:"descripcion"`
	Important      int    `json:"importante"`
}

func (s *Service) sendToPardo360(dto CreateDTO) {
	if s.pardo360URL == "" || s.pardo360Project == 0 {
		return
	}

	payload := pardo360Notification{
		ProjectID:   s.pardo360Project,
		RecipientID: dto.UsuarioID,
		Heading:     notificationHeading(dto.Tipo),
		Description: dto.Mensaje,
		Important:   notificationImportance(dto.Tipo),
	}
	var user db.Usuario
	if err := s.db.Select("email").First(&user, dto.UsuarioID).Error; err == nil && strings.TrimSpace(user.Email) != "" {
		payload.RecipientID = 0
		payload.RecipientEmail = strings.TrimSpace(user.Email)
	}
	body, err := json.Marshal(payload)
	if err != nil {
		log.Printf("notificaciones Pardo 360: no se pudo serializar el evento: %v", err)
		return
	}

	req, err := http.NewRequest(http.MethodPost, s.pardo360URL, bytes.NewReader(body))
	if err != nil {
		log.Printf("notificaciones Pardo 360: URL inválida: %v", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	res, err := s.httpClient.Do(req)
	if err != nil {
		log.Printf("notificaciones Pardo 360: error enviando a %s: %v", s.pardo360URL, err)
		return
	}
	defer res.Body.Close()

	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusMultipleChoices {
		log.Printf("notificaciones Pardo 360: endpoint respondió HTTP %d", res.StatusCode)
	}
}

func notificationHeading(tipo string) string {
	switch tipo {
	case "Asignacion":
		return "Nueva asignación en SIGEFAE"
	case "Recordatorio":
		return "Recordatorio de SIGEFAE"
	case "Rechazo":
		return "Documento rechazado en SIGEFAE"
	case "Devolucion":
		return "Documento devuelto en SIGEFAE"
	case "Finalizado":
		return "Documento finalizado en SIGEFAE"
	default:
		return "Actualización en SIGEFAE"
	}
}

func notificationImportance(tipo string) int {
	switch tipo {
	case "Rechazo", "Devolucion":
		return 1
	default:
		return 0
	}
}

func (s *Service) Delete(id uint) error {
	var notificacion db.Notificacion
	if err := s.db.First(&notificacion, id).Error; err != nil {
		return err
	}
	return s.db.Delete(&notificacion).Error
}
