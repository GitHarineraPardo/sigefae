package evento_dian

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"sigefae/internal/db"
	"sigefae/procesos_aplicacion/proveedor_inbound"
)

type Service struct {
	db       *gorm.DB
	ptClient *proveedor_inbound.Client
}

func NewService(database *gorm.DB, ptClient *proveedor_inbound.Client) *Service {
	return &Service{
		db:       database,
		ptClient: ptClient,
	}
}

// ListByRadicadoID obtiene la lista histórica de eventos emitidos para el radicado dado
func (s *Service) ListByRadicadoID(radicadoID uint) ([]Response, error) {
	var rad db.DocumentoRadicado
	if err := s.db.Preload("DocumentoComercial").First(&rad, radicadoID).Error; err != nil {
		return nil, errors.New("radicado no encontrado")
	}

	var eventos []db.EventoRadian
	if err := s.db.
		Preload("Usuario").
		Where("documento_comercial_id = ?", rad.DocumentoComercialID).
		Order("created_at ASC").
		Find(&eventos).Error; err != nil {
		return nil, err
	}

	res := make([]Response, 0, len(eventos)+1)
	hasAcuse := false
	for _, e := range eventos {
		if e.CodigoDIAN == "030" || strings.EqualFold(e.TipoEvento, "ACKNOWLEDGED") {
			hasAcuse = true
		}
		userName := "Sistema"
		if e.Usuario != nil && e.Usuario.Nombre != "" {
			userName = e.Usuario.Nombre
		}
		res = append(res, Response{
			ID:                   e.ID,
			DocumentoComercialID: e.DocumentoComercialID,
			TipoEvento:           e.TipoEvento,
			CodigoDIAN:           e.CodigoDIAN,
			FechaEnvio:           e.FechaEnvio,
			Exitoso:              e.Exitoso,
			MensajeRespuesta:     e.MensajeRespuesta,
			UsuarioID:            e.UsuarioID,
			UsuarioNombre:        userName,
			Observacion:          e.Observacion,
			CreatedAt:            e.CreatedAt,
		})
	}

	// Si no hay evento 030 manual pero la factura ya vino acusada por el proveedor (buzón Inbound)
	if !hasAcuse && rad.DocumentoComercial != nil && strings.EqualFold(strings.TrimSpace(rad.DocumentoComercial.EventoDIAN), "ACKNOWLEDGED") {
		fechaEnvio := rad.DocumentoComercial.CreatedAt
		if rad.DocumentoComercial.FechaRecepcionDIAN != nil {
			fechaEnvio = *rad.DocumentoComercial.FechaRecepcionDIAN
		}
		acuseAuto := Response{
			ID:                   0,
			DocumentoComercialID: rad.DocumentoComercial.ID,
			TipoEvento:           "ACKNOWLEDGED",
			CodigoDIAN:           "030",
			FechaEnvio:           fechaEnvio,
			Exitoso:              true,
			MensajeRespuesta:     "Acuse registrado por Buzón Inbound",
			UsuarioNombre:        "Buzón Inbound (Automático)",
			Observacion:          "Acuse de recibo generado por el proveedor tecnológico al recibir la factura",
			CreatedAt:            fechaEnvio,
		}
		res = append([]Response{acuseAuto}, res...)
	}

	return res, nil
}

// EmitirEvento valida permisos de usuario, estado de la ruta y envía el evento al Proveedor Tecnológico
func (s *Service) EmitirEvento(radicadoID uint, usuarioID uint, req EmitirEventoDTO) (*Response, error) {
	// 1. Validar Usuario y Cargo
	var user db.Usuario
	if err := s.db.Preload("Cargo").Preload("Rol").First(&user, usuarioID).Error; err != nil {
		return nil, errors.New("usuario no encontrado")
	}

	esAuxOriente := user.Cargo != nil && strings.EqualFold(strings.TrimSpace(user.Cargo.Nombre), "Auxiliar Admon Oriente")

	if !esAuxOriente {
		return nil, errors.New("solo el cargo Auxiliar Admon Oriente tiene autorización para emitir eventos ante la DIAN")
	}

	// 2. Validar Radicado y Estado de la Ruta
	var rad db.DocumentoRadicado
	if err := s.db.
		Preload("DocumentoComercial").
		Preload("DocumentoComercial.Receptor").
		First(&rad, radicadoID).Error; err != nil {
		return nil, errors.New("radicado no encontrado")
	}

	if rad.DocumentoComercial == nil {
		return nil, errors.New("el radicado no tiene un documento comercial asociado")
	}

	if rad.DocumentoComercial.LDF == nil || strings.TrimSpace(*rad.DocumentoComercial.LDF) == "" {
		return nil, errors.New("el documento comercial no cuenta con LDF (Identificador Proveedor Tecnológico)")
	}
	ldf := strings.TrimSpace(*rad.DocumentoComercial.LDF)

	// La ruta debe haber finalizado
	if rad.EstadoPosesion != "Completado" {
		// Validar si existe alguna tarea pendiente o en proceso
		var tareasIncompletas int64
		s.db.Model(&db.Tarea{}).
			Joins("JOIN estado_tarea ON tarea.estado_id = estado_tarea.id").
			Where("tarea.documento_radicado_id = ? AND estado_tarea.nombre NOT IN ('Completada', 'Omitida')", rad.ID).
			Count(&tareasIncompletas)

		if tareasIncompletas > 0 {
			return nil, errors.New("no es posible emitir eventos ante la DIAN hasta que la ruta de aprobación esté completamente finalizada")
		}
	}

	// 3. Normalizar tipo de evento
	tipo := strings.ToUpper(strings.TrimSpace(req.Tipo))
	var codigoDIAN, tipoEvento string

	switch tipo {
	case "030", "ACUSE", "ACKNOWLEDGED":
		codigoDIAN = "030"
		tipoEvento = "ACKNOWLEDGED"
	case "032", "RECIBO", "RECEIVEDGOODS":
		codigoDIAN = "032"
		tipoEvento = "RECEIVEDGOODS"
	case "034", "ACEPTACION", "ACCEPTED":
		codigoDIAN = "034"
		tipoEvento = "ACCEPTED"
	case "031", "RECLAMO", "REJECTED":
		codigoDIAN = "031"
		tipoEvento = "REJECTED"
	default:
		return nil, fmt.Errorf("tipo de evento '%s' no reconocido (use 030, 032, 034 o 031)", req.Tipo)
	}

	// 3.1. Validar pre-requisitos de eventos DIAN
	var eventosPrevios []db.EventoRadian
	s.db.Where("documento_comercial_id = ?", rad.DocumentoComercialID).Find(&eventosPrevios)

	for _, ep := range eventosPrevios {
		if ep.CodigoDIAN == codigoDIAN {
			return nil, fmt.Errorf("el evento %s (%s) ya fue emitido previamente para esta factura", codigoDIAN, tipoEvento)
		}
	}

	tieneAcuse := rad.DocumentoComercial != nil && strings.EqualFold(strings.TrimSpace(rad.DocumentoComercial.EventoDIAN), "ACKNOWLEDGED")
	tieneReciboBienes := rad.DocumentoComercial != nil && strings.EqualFold(strings.TrimSpace(rad.DocumentoComercial.EventoDIAN), "RECEIVEDGOODS")

	for _, ep := range eventosPrevios {
		if ep.CodigoDIAN == "030" || strings.EqualFold(ep.TipoEvento, "ACKNOWLEDGED") {
			tieneAcuse = true
		}
		if ep.CodigoDIAN == "032" || strings.EqualFold(ep.TipoEvento, "RECEIVEDGOODS") {
			tieneReciboBienes = true
		}
	}

	if codigoDIAN == "032" && !tieneAcuse {
		return nil, errors.New("no es posible emitir el Recibo de Bienes o Servicios (032) sin haber registrado previamente el Acuse de Recibo (030)")
	}

	if (codigoDIAN == "034" || codigoDIAN == "031") && !tieneReciboBienes {
		return nil, errors.New("no es posible emitir la Aceptación (034) o Reclamo (031) sin haber registrado previamente el Recibo de Bienes o Servicios (032)")
	}

	// 4. Datos del receptor
	nitReceptor := "890116102"
	nombreReceptor := "HARINERA PARDO S.A.S."
	if rad.DocumentoComercial.Receptor != nil {
		if rad.DocumentoComercial.Receptor.NumeroDocumento != "" {
			nitReceptor = rad.DocumentoComercial.Receptor.NumeroDocumento
		}
		if rad.DocumentoComercial.Receptor.Nombre != "" {
			nombreReceptor = rad.DocumentoComercial.Receptor.Nombre
		}
	}

	medio := req.Medio
	if medio == "" {
		medio = fmt.Sprintf("%s@3beta.factureinbox.co", nitReceptor)
	}

	motivo := req.Motivo
	if motivo == "" {
		switch codigoDIAN {
		case "030":
			motivo = "Acuse"
		case "032":
			motivo = "Otro"
		case "034":
			motivo = "Aceptación"
		case "031":
			motivo = "Servicio no prestado"
		}
	}

	obs := req.Observacion
	if obs == "" {
		switch codigoDIAN {
		case "030":
			obs = "Factura electrónica recibida"
		case "032":
			obs = "Bienes o servicios recibidos a satisfacción"
		case "034":
			obs = "Factura aprobada y aceptada para pago"
		case "031":
			obs = "Reclamo registrado por disconformidad"
		}
	}

	comments := req.Comentario
	if comments == "" {
		comments = obs
	}

	// 5. Enviar petición al Proveedor Tecnológico si el cliente está configurado
	var respuestaPT string
	var errPT error

	if s.ptClient != nil {
		switch codigoDIAN {
		case "030":
			acuseReq := proveedor_inbound.AcuseEventRequest{
				Motive:         motivo,
				Observation:    obs,
				Comments:       comments,
				SourceDelivery: "INBOX",
				Canal:          "ERP",
				Medio:          medio,
			}
			respuestaPT, errPT = s.ptClient.EmitirAcuse(ldf, acuseReq)

		case "032":
			dv := "9"
			if len(nitReceptor) > 1 {
				dv = string(nitReceptor[len(nitReceptor)-1])
			}
			commentsPtr := &comments
			reciboReq := proveedor_inbound.ReciboGoodsEventRequest{
				Motive:                    motivo,
				Observation:               obs,
				Comments:                  commentsPtr,
				SourceDelivery:            "INBOX",
				Canal:                     "INBOX",
				Medio:                     medio,
				ReceiverDocumentType:      "NIT",
				ReceiverDocumentNumber:    nitReceptor,
				ReceiverVerificationDigit: dv,
				ReceiverName:              nombreReceptor,
				ReceiverLastName:          nombreReceptor,
				ReceiveDateTime:           time.Now().Format("2006-01-02T15:04:05-07:00"),
			}
			respuestaPT, errPT = s.ptClient.EmitirReciboBienes(ldf, reciboReq)

		case "034":
			aceptReq := proveedor_inbound.AceptacionEventRequest{
				Motive:         motivo,
				Observation:    obs,
				Comments:       comments,
				SourceDelivery: "INBOX",
				Canal:          "INBOX",
				Medio:          medio,
			}
			respuestaPT, errPT = s.ptClient.EmitirAceptacion(ldf, aceptReq)

		case "031":
			codigoMotivo := req.CodigoMotivo
			if codigoMotivo == "" {
				codigoMotivo = "01"
			}
			reclamoReq := proveedor_inbound.ReclamoEventRequest{
				Motive:         motivo,
				Observation:    obs,
				Comments:       comments,
				SourceDelivery: "INBOX",
				Canal:          "INBOX",
				Medio:          medio,
				CodigoMotivo:   codigoMotivo,
			}
			respuestaPT, errPT = s.ptClient.EmitirReclamo(ldf, reclamoReq)
		}

		if errPT != nil {
			return nil, fmt.Errorf("error del Proveedor Tecnológico DIAN: %w", errPT)
		}
	} else {
		respuestaPT = "Evento simulado (PT API no configurada)"
	}

	// 6. Guardar en Base de Datos
	eventoRadian := db.EventoRadian{
		DocumentoComercialID: rad.DocumentoComercial.ID,
		TipoEvento:           tipoEvento,
		CodigoDIAN:           codigoDIAN,
		FechaEnvio:           time.Now(),
		Exitoso:              true,
		MensajeRespuesta:     respuestaPT,
		UsuarioID:            &user.ID,
		Observacion:          obs,
		CreatedAt:            time.Now(),
	}

	if err := s.db.Create(&eventoRadian).Error; err != nil {
		return nil, fmt.Errorf("error registrando auditoría del evento: %w", err)
	}

	// Actualizar evento_dian en el Documento Comercial
	_ = s.db.Model(rad.DocumentoComercial).Update("evento_dian", tipoEvento).Error

	// Registrar en Trazabilidad general del radicado
	s.db.Create(&db.Trazabilidad{
		DocumentoRadicadoID: rad.ID,
		UsuarioID:           user.ID,
		Accion:              "Evento DIAN Emitido",
		Descripcion:         fmt.Sprintf("Evento %s (%s) emitido ante la DIAN por Auxiliar Admon Oriente. Observación: %s", codigoDIAN, tipoEvento, obs),
		Fecha:               time.Now(),
	})

	return &Response{
		ID:                   eventoRadian.ID,
		DocumentoComercialID: eventoRadian.DocumentoComercialID,
		TipoEvento:           eventoRadian.TipoEvento,
		CodigoDIAN:           eventoRadian.CodigoDIAN,
		FechaEnvio:           eventoRadian.FechaEnvio,
		Exitoso:              eventoRadian.Exitoso,
		MensajeRespuesta:     eventoRadian.MensajeRespuesta,
		UsuarioID:            eventoRadian.UsuarioID,
		UsuarioNombre:        user.Nombre,
		Observacion:          eventoRadian.Observacion,
		CreatedAt:            eventoRadian.CreatedAt,
	}, nil
}
