package evento_dian

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
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

	if rad.DocumentoComercial == nil {
		return []Response{}, nil
	}

	var eventos []db.EventoRadian
	if err := s.db.
		Preload("Usuario").
		Where("documento_comercial_id = ?", rad.DocumentoComercialID).
		Order("created_at ASC").
		Find(&eventos).Error; err != nil {
		return nil, err
	}

	has030 := false
	has032 := false
	has034 := false
	has031 := false

	for _, e := range eventos {
		if e.CodigoDIAN == "030" || strings.EqualFold(e.TipoEvento, "ACKNOWLEDGED") {
			has030 = true
		}
		if e.CodigoDIAN == "032" || strings.EqualFold(e.TipoEvento, "RECEIVEDGOODS") {
			has032 = true
		}
		if e.CodigoDIAN == "034" || strings.EqualFold(e.TipoEvento, "ACCEPTED") {
			has034 = true
		}
		if e.CodigoDIAN == "031" || strings.EqualFold(e.TipoEvento, "REJECTED") {
			has031 = true
		}
	}

	docEvento := strings.ToUpper(strings.TrimSpace(rad.DocumentoComercial.EventoDIAN))
	fechaRecepcion := rad.DocumentoComercial.CreatedAt
	if rad.DocumentoComercial.FechaRecepcionDIAN != nil && !rad.DocumentoComercial.FechaRecepcionDIAN.IsZero() {
		fechaRecepcion = *rad.DocumentoComercial.FechaRecepcionDIAN
	}

	// Inferencia de eventos provenientes del Proveedor Tecnológico (Inbound) o requeridos por RADIAN:
	// Si el documento comercial indica un evento o si ya se emitieron eventos posteriores en el sistema:
	debeTener030 := has030 || has032 || has034 || has031 ||
		docEvento == "ACKNOWLEDGED" || docEvento == "RECEIVEDGOODS" || docEvento == "ACCEPTED" || docEvento == "REJECTED"
	debeTener032 := has032 || has034 || has031 ||
		docEvento == "RECEIVEDGOODS" || docEvento == "ACCEPTED" || docEvento == "REJECTED"
	debeTener034 := has034 || docEvento == "ACCEPTED"
	debeTener031 := has031 || docEvento == "REJECTED"

	// Auto-sanar / persistir en la tabla evento_radian los eventos previos del proveedor tecnológico
	// para que queden guardados de forma inmutable y con trazabilidad
	if debeTener030 && !has030 {
		ev030 := db.EventoRadian{
			DocumentoComercialID: rad.DocumentoComercial.ID,
			TipoEvento:           "ACKNOWLEDGED",
			CodigoDIAN:           "030",
			FechaEnvio:           fechaRecepcion,
			Exitoso:              true,
			MensajeRespuesta:     "Acuse registrado por Buzón Inbound / Proveedor Tecnológico",
			Observacion:          "Acuse de recibo generado automáticamente al recibir la factura",
			CreatedAt:            fechaRecepcion,
		}
		if err := s.db.Create(&ev030).Error; err == nil {
			eventos = append([]db.EventoRadian{ev030}, eventos...)
			has030 = true
		}
	}

	if debeTener032 && !has032 {
		ev032 := db.EventoRadian{
			DocumentoComercialID: rad.DocumentoComercial.ID,
			TipoEvento:           "RECEIVEDGOODS",
			CodigoDIAN:           "032",
			FechaEnvio:           fechaRecepcion,
			Exitoso:              true,
			MensajeRespuesta:     "Recibo de bienes registrado por Proveedor Tecnológico",
			Observacion:          "Bienes o servicios recibidos según registro del proveedor tecnológico",
			CreatedAt:            fechaRecepcion,
		}
		if err := s.db.Create(&ev032).Error; err == nil {
			eventos = append(eventos, ev032)
			has032 = true
		}
	}

	if debeTener034 && !has034 {
		ev034 := db.EventoRadian{
			DocumentoComercialID: rad.DocumentoComercial.ID,
			TipoEvento:           "ACCEPTED",
			CodigoDIAN:           "034",
			FechaEnvio:           fechaRecepcion,
			Exitoso:              true,
			MensajeRespuesta:     "Aceptación expresa registrada por Proveedor Tecnológico",
			Observacion:          "Factura aceptada según registro del proveedor tecnológico",
			CreatedAt:            fechaRecepcion,
		}
		if err := s.db.Create(&ev034).Error; err == nil {
			eventos = append(eventos, ev034)
			has034 = true
		}
	}

	if debeTener031 && !has031 {
		ev031 := db.EventoRadian{
			DocumentoComercialID: rad.DocumentoComercial.ID,
			TipoEvento:           "REJECTED",
			CodigoDIAN:           "031",
			FechaEnvio:           fechaRecepcion,
			Exitoso:              true,
			MensajeRespuesta:     "Reclamo registrado por Proveedor Tecnológico",
			Observacion:          "Reclamo según registro del proveedor tecnológico",
			CreatedAt:            fechaRecepcion,
		}
		if err := s.db.Create(&ev031).Error; err == nil {
			eventos = append(eventos, ev031)
			has031 = true
		}
	}

	// Ordenar eventos según secuencia lógica DIAN y cronología
	sort.SliceStable(eventos, func(i, j int) bool {
		peso := func(cod string) int {
			switch cod {
			case "030":
				return 1
			case "032":
				return 2
			case "034":
				return 3
			case "031":
				return 4
			default:
				return 5
			}
		}
		if peso(eventos[i].CodigoDIAN) != peso(eventos[j].CodigoDIAN) {
			return peso(eventos[i].CodigoDIAN) < peso(eventos[j].CodigoDIAN)
		}
		return eventos[i].CreatedAt.Before(eventos[j].CreatedAt)
	})

	res := make([]Response, 0, len(eventos))
	for _, e := range eventos {
		userName := "Buzón Inbound (Automático)"
		if e.Usuario != nil && e.Usuario.Nombre != "" {
			userName = e.Usuario.Nombre
		} else if e.UsuarioID != nil {
			userName = "Usuario #" + strconv.FormatUint(uint64(*e.UsuarioID), 10)
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

	docEvento := ""
	if rad.DocumentoComercial != nil {
		docEvento = strings.ToUpper(strings.TrimSpace(rad.DocumentoComercial.EventoDIAN))
	}

	tiene030 := docEvento == "ACKNOWLEDGED" || docEvento == "RECEIVEDGOODS" || docEvento == "ACCEPTED" || docEvento == "REJECTED"
	tiene032 := docEvento == "RECEIVEDGOODS" || docEvento == "ACCEPTED" || docEvento == "REJECTED"
	tiene034 := docEvento == "ACCEPTED"
	tiene031 := docEvento == "REJECTED"

	for _, ep := range eventosPrevios {
		switch ep.CodigoDIAN {
		case "030":
			tiene030 = true
		case "032":
			tiene030 = true
			tiene032 = true
		case "034":
			tiene030 = true
			tiene032 = true
			tiene034 = true
		case "031":
			tiene030 = true
			tiene032 = true
			tiene031 = true
		}
		if strings.EqualFold(ep.TipoEvento, "ACKNOWLEDGED") {
			tiene030 = true
		} else if strings.EqualFold(ep.TipoEvento, "RECEIVEDGOODS") {
			tiene030 = true
			tiene032 = true
		} else if strings.EqualFold(ep.TipoEvento, "ACCEPTED") {
			tiene030 = true
			tiene032 = true
			tiene034 = true
		} else if strings.EqualFold(ep.TipoEvento, "REJECTED") {
			tiene030 = true
			tiene032 = true
			tiene031 = true
		}
	}

	switch codigoDIAN {
	case "030":
		if tiene030 {
			return nil, fmt.Errorf("el evento 030 (ACKNOWLEDGED) ya fue emitido o registrado previamente para esta factura")
		}
	case "032":
		if tiene032 {
			return nil, fmt.Errorf("el evento 032 (RECEIVEDGOODS) ya fue emitido o registrado previamente para esta factura")
		}
		if !tiene030 {
			return nil, errors.New("no es posible emitir el Recibo de Bienes o Servicios (032) sin haber registrado previamente el Acuse de Recibo (030)")
		}
	case "034":
		if tiene034 {
			return nil, fmt.Errorf("el evento 034 (ACCEPTED) ya fue emitido previamente para esta factura")
		}
		if tiene031 {
			return nil, errors.New("la factura ya cuenta con un Reclamo (031); no es posible emitir Aceptación")
		}
		if !tiene032 {
			return nil, errors.New("no es posible emitir la Aceptación (034) sin haber registrado previamente el Recibo de Bienes o Servicios (032)")
		}
	case "031":
		if tiene031 {
			return nil, fmt.Errorf("el evento 031 (REJECTED) ya fue emitido previamente para esta factura")
		}
		if tiene034 {
			return nil, errors.New("la factura ya cuenta con Aceptación (034); no es posible emitir Reclamo")
		}
		if !tiene032 {
			return nil, errors.New("no es posible emitir el Reclamo (031) sin haber registrado previamente el Recibo de Bienes o Servicios (032)")
		}
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

	// Si se emitió 032, 034 o 031 y el acuse (030) previo no estaba registrado en evento_radian,
	// guardarlo para preservar la integridad histórica y auditoría permanente.
	if codigoDIAN != "030" {
		var count030 int64
		s.db.Model(&db.EventoRadian{}).
			Where("documento_comercial_id = ? AND (codigo_dian = '030' OR tipo_evento = 'ACKNOWLEDGED')", rad.DocumentoComercial.ID).
			Count(&count030)
		if count030 == 0 {
			fechaAcuse := rad.DocumentoComercial.CreatedAt
			if rad.DocumentoComercial.FechaRecepcionDIAN != nil && !rad.DocumentoComercial.FechaRecepcionDIAN.IsZero() {
				fechaAcuse = *rad.DocumentoComercial.FechaRecepcionDIAN
			}
			_ = s.db.Create(&db.EventoRadian{
				DocumentoComercialID: rad.DocumentoComercial.ID,
				TipoEvento:           "ACKNOWLEDGED",
				CodigoDIAN:           "030",
				FechaEnvio:           fechaAcuse,
				Exitoso:              true,
				MensajeRespuesta:     "Acuse registrado por Buzón Inbound / Proveedor Tecnológico",
				Observacion:          "Acuse de recibo generado automáticamente al recibir la factura",
				CreatedAt:            fechaAcuse,
			}).Error
		}
	}

	if codigoDIAN == "034" || codigoDIAN == "031" {
		var count032 int64
		s.db.Model(&db.EventoRadian{}).
			Where("documento_comercial_id = ? AND (codigo_dian = '032' OR tipo_evento = 'RECEIVEDGOODS')", rad.DocumentoComercial.ID).
			Count(&count032)
		if count032 == 0 {
			fechaRecibo := rad.DocumentoComercial.CreatedAt
			if rad.DocumentoComercial.FechaRecepcionDIAN != nil && !rad.DocumentoComercial.FechaRecepcionDIAN.IsZero() {
				fechaRecibo = *rad.DocumentoComercial.FechaRecepcionDIAN
			}
			_ = s.db.Create(&db.EventoRadian{
				DocumentoComercialID: rad.DocumentoComercial.ID,
				TipoEvento:           "RECEIVEDGOODS",
				CodigoDIAN:           "032",
				FechaEnvio:           fechaRecibo,
				Exitoso:              true,
				MensajeRespuesta:     "Recibo de bienes registrado por Proveedor Tecnológico",
				Observacion:          "Bienes o servicios recibidos según registro del proveedor tecnológico",
				CreatedAt:            fechaRecibo,
			}).Error
		}
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
