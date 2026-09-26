package main

import (
	"fmt"
	"log"

	"sigefae/env"
	"sigefae/internal/db"
)

func main() {
	cfg, err := env.Load(".env")
	if err != nil {
		log.Fatal(err)
	}

	if _, err := db.Connect(cfg); err != nil {
		log.Fatal("Error conectando a la base de datos:", err)
	}
	if err := db.Migrate(); err != nil {
		log.Fatal("Error migrando cargos:", err)
	}

	fmt.Println("Migración relacional de cargos completada.")
}
