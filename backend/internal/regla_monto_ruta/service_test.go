package regla_monto_ruta

import (
	"testing"

	"sigefae/internal/db"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestUpdateCanChangeRouteScopedRuleToGlobal(t *testing.T) {
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
	rule := db.ReglaMontoRuta{
		RutaID:            &rutaID,
		MontoMinimoSmmlv:  10,
		MontoMaximoSmmlv:  20,
		PosicionInsercion: "ULTIMO",
		Activo:            true,
	}
	if err := database.Create(&rule).Error; err != nil {
		t.Fatal(err)
	}

	if err := New(database).Update(rule.ID, &db.ReglaMontoRuta{
		MontoMinimoSmmlv:  5,
		MontoMaximoSmmlv:  0,
		PosicionInsercion: "ULTIMO",
	}); err != nil {
		t.Fatal(err)
	}

	var updated db.ReglaMontoRuta
	if err := database.First(&updated, rule.ID).Error; err != nil {
		t.Fatal(err)
	}
	if updated.RutaID != nil {
		t.Fatalf("RutaID = %d, esperaba nil para regla global", *updated.RutaID)
	}
}
