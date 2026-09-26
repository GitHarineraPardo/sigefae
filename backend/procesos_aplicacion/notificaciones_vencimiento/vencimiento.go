package notificaciones_vencimiento

import (
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"
	"sigefae/internal/db"
)

// Start inicia un worker que corre cada hora y revisa radicados próximos a vencer
func Start(database *gorm.DB) {
	// Ejecutar la primera vez inmediatamente
	checkVencimientos(database)

	ticker := time.NewTicker(1 * time.Hour) // Se ejecuta cada hora
	for {
		<-ticker.C
		err := checkVencimientos(database)
		if err != nil {
			log.Printf("[Vencimientos] Error checking: %v", err)
		}
	}
}

func checkVencimientos(database *gorm.DB) error {
	var radicados []db.DocumentoRadicado

	ahora := time.Now()
	limite := ahora.AddDate(0, 0, 3) // Cerca de vencer = menos de 3 días (o ya vencido)

	err := database.
		Preload("DocumentoComercial").
		Where("estado_posesion NOT IN (?)", []string{"Finalizado", "Rechazado", "Anulado", "Contabilizado", "Pagado"}).
		Find(&radicados).Error

	if err != nil {
		return err
	}

	for _, rad := range radicados {
		if rad.DocumentoComercial == nil || rad.DocumentoComercial.FechaVencimiento == nil {
			continue
		}

		vencimiento := *rad.DocumentoComercial.FechaVencimiento

		if vencimiento.Before(limite) {
			// Revisar si ya mandamos una notificación en las últimas 3 horas
			var ultimaNotif db.Notificacion
			err := database.Where("documento_radicado_id = ? AND tipo = ? AND fecha_creacion > ?",
				rad.ID, "Recordatorio", ahora.Add(-3*time.Hour)).
				Order("fecha_creacion DESC").
				First(&ultimaNotif).Error

			if err == nil {
				// Ya se mandó una notificación en las últimas 3 horas para este radicado, saltar
				continue
			}

			mensaje := fmt.Sprintf("ALERTA: El radicado %s está próximo a vencer o vencido (Vencimiento: %s)",
				rad.NumeroRadicado, vencimiento.Format("2006-01-02"))

			// 1. Notificar a la persona que lo tiene actualmente
			if rad.UsuarioActualID != nil {
				database.Create(&db.Notificacion{
					UsuarioID:           *rad.UsuarioActualID,
					DocumentoRadicadoID: &rad.ID,
					Mensaje:             mensaje,
					Estado:              "Pendiente",
					Tipo:                "Recordatorio",
					FechaCreacion:       ahora,
				})
			}

			// 2. Notificar a los administradores
			var admins []db.Usuario
			database.Joins("JOIN rols ON rols.id = usuarios.id_rol").Where("rols.nombre = ?", "Administrador").Find(&admins)
			for _, admin := range admins {
				if rad.UsuarioActualID != nil && admin.ID == *rad.UsuarioActualID {
					continue
				}
				database.Create(&db.Notificacion{
					UsuarioID:           admin.ID,
					DocumentoRadicadoID: &rad.ID,
					Mensaje:             mensaje,
					Estado:              "Pendiente",
					Tipo:                "Recordatorio",
					FechaCreacion:       ahora,
				})
			}
		}
	}
	return nil
}
