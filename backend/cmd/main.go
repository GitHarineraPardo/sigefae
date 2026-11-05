package main

import (
	"fmt"
	"log"
	"time"

	"sigefae/env"
	"sigefae/internal/api"
	"sigefae/internal/db"
	"sigefae/procesos_aplicacion/notificaciones_vencimiento"
	"sigefae/procesos_aplicacion/proveedor_inbound"
)

func main() {

	// ========================================
	// Configuración
	// ========================================

	cfg, err := env.Load(".env")
	if err != nil {
		log.Fatal(err)
	}

	// ========================================
	// Base de datos
	// ========================================

	database, err := db.Connect(cfg)
	if err != nil {
		log.Fatal(err)
	}

	if err := db.Migrate(); err != nil {
		log.Fatal(err)
	}

	if err := db.Seed(db.DB); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Environment loaded successfully.")
	fmt.Println("Database connected successfully.")

	// ========================================
	// API & Tareas en Segundo Plano
	// ========================================

	// Iniciar la sincronización directa con el Proveedor Tecnológico DIAN
	var ptService *proveedor_inbound.Service
	if cfg.PtAPIBaseURL != "" && cfg.PtAPIUser != "" {
		ptClient := proveedor_inbound.NewClient(
			cfg.PtAPIBaseURL,
			cfg.PtAPIUser,
			cfg.PtAPIPassword,
			cfg.PtAPIXWhoListarDescargar,
			cfg.PtAPIXWhoAuthEventos,
		)
		ptService = proveedor_inbound.New(database, ptClient)
		go proveedor_inbound.StartBackgroundTaskWithService(ptService, 10*time.Minute)
	}

	// Iniciar la tarea de notificaciones de vencimiento
	go notificaciones_vencimiento.Start(database)

	router := api.New(database, ptService)

	log.Println("Servidor iniciado en http://localhost:8080")

	log.Fatal(router.Run(":8080"))
}
