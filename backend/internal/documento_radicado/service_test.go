package documento_radicado

import (
	"testing"

	"sigefae/internal/db"
	"sigefae/internal/regla_nit_ruta"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestReglasMontoIncluyeGlobalesYDeLaRutaSeleccionada(t *testing.T) {
	database, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := database.AutoMigrate(&db.ReglaMontoRuta{}); err != nil {
		t.Fatal(err)
	}

	rutaID := uint(7)
	otraRutaID := uint(9)
	reglas := []db.ReglaMontoRuta{
		{Activo: true},
		{RutaID: &rutaID, Activo: true},
		{RutaID: &otraRutaID, Activo: true},
	}
	if err := database.Create(&reglas).Error; err != nil {
		t.Fatal(err)
	}

	var aplicables []db.ReglaMontoRuta
	if err := reglasMontoActivasPorRuta(database, rutaID).Find(&aplicables).Error; err != nil {
		t.Fatal(err)
	}
	if len(aplicables) != 2 {
		t.Fatalf("se encontraron %d reglas, esperaba 2 (global y de la ruta)", len(aplicables))
	}
}

func TestOrdenarReglasRutaEsDeterministico(t *testing.T) {
	reglas := []reglaRutaAplicable{
		{ID: 4, Tipo: "NIT", Prioridad: 2},
		{ID: 9, Tipo: "MONTO", Prioridad: 1},
		{ID: 3, Tipo: "NIT", Prioridad: 1},
		{ID: 2, Tipo: "MONTO", Prioridad: 1},
	}

	ordenarReglasRuta(reglas)

	esperados := []struct {
		tipo string
		id   uint
	}{
		{tipo: "MONTO", id: 2},
		{tipo: "MONTO", id: 9},
		{tipo: "NIT", id: 3},
		{tipo: "NIT", id: 4},
	}

	for i, esperado := range esperados {
		if reglas[i].Tipo != esperado.tipo || reglas[i].ID != esperado.id {
			t.Fatalf("regla %d = (%s, %d), esperaba (%s, %d)", i, reglas[i].Tipo, reglas[i].ID, esperado.tipo, esperado.id)
		}
	}
}

func TestNormalizarNITIgnoraSeparadores(t *testing.T) {
	if got := regla_nit_ruta.NormalizarNIT(" 900.123.456-7 "); got != "9001234567" {
		t.Fatalf("NormalizarNIT() = %q, esperaba %q", got, "9001234567")
	}
	if got := regla_nit_ruta.NormalizarNIT("sin dígitos"); got != "" {
		t.Fatalf("NormalizarNIT() = %q, esperaba vacío", got)
	}
}
