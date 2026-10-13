package regla_nit_ruta

import (
	"testing"

	"sigefae/internal/db"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestValidarReglaNormalizaNITYValidaInsercion(t *testing.T) {
	cargoID := uint(3)
	regla := db.ReglaNITRuta{
		NIT: " 900.123.456-7 ", CargoAprobadorID: &cargoID,
		PosicionInsercion: "ANTES_FINAL", Prioridad: 1,
	}

	if err := validarRegla(&regla); err != nil {
		t.Fatalf("validarRegla() error = %v", err)
	}
	if regla.NIT != "9001234567" {
		t.Fatalf("NIT normalizado = %q, esperaba %q", regla.NIT, "9001234567")
	}

	regla.PosicionInsercion = "AL_FINAL"
	if err := validarRegla(&regla); err == nil {
		t.Fatal("validarRegla() aceptó una posición no soportada")
	}
}

func TestUpdatePreservaEstadoActivoYPrioridadCero(t *testing.T) {
	database, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := database.AutoMigrate(&db.ReglaNITRuta{}); err != nil {
		t.Fatal(err)
	}

	cargoID := uint(3)
	regla := db.ReglaNITRuta{
		NIT: "9001234567", CargoAprobadorID: &cargoID,
		PosicionInsercion: "ULTIMO", Prioridad: 2, Activo: true,
	}
	if err := database.Create(&regla).Error; err != nil {
		t.Fatal(err)
	}

	actualizada := db.ReglaNITRuta{
		NIT: "9001234567", CargoAprobadorID: &cargoID,
		PosicionInsercion: "PRIMERO", Prioridad: 0,
	}
	if err := New(database).Update(regla.ID, &actualizada); err != nil {
		t.Fatal(err)
	}

	var guardada db.ReglaNITRuta
	if err := database.First(&guardada, regla.ID).Error; err != nil {
		t.Fatal(err)
	}
	if !guardada.Activo {
		t.Fatal("la actualización desactivó la regla")
	}
	if guardada.Prioridad != 0 {
		t.Fatalf("prioridad = %d, esperaba 0", guardada.Prioridad)
	}
}
