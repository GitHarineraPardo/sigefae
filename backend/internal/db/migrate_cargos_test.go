package db

import (
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestMigrateCargoForeignKeys(t *testing.T) {
	database, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{DisableForeignKeyConstraintWhenMigrating: true})
	if err != nil {
		t.Fatal(err)
	}
	if err := database.AutoMigrate(BaseModels()...); err != nil {
		t.Fatal(err)
	}
	if err := database.AutoMigrate(CircularModels()...); err != nil {
		t.Fatal(err)
	}
	if err := MigrateCargoForeignKeys(database); err != nil {
		t.Fatal(err)
	}

	relations := []struct {
		model any
		field string
	}{
		{model: &Usuario{}, field: "Cargo"},
		{model: &PasoRuta{}, field: "Cargo"},
		{model: &Tarea{}, field: "CargoAsignadoCargo"},
		{model: &DocumentoRadicado{}, field: "CargoActual"},
		{model: &ReglaMontoRuta{}, field: "CargoAprobador"},
	}
	for _, relation := range relations {
		if !database.Migrator().HasConstraint(relation.model, relation.field) {
			t.Errorf("missing foreign key for %T.%s", relation.model, relation.field)
		}
	}
}

func TestMigrateCargoIDsBackfillsLegacyValuesIdempotently(t *testing.T) {
	database, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := database.AutoMigrate(&Cargo{}); err != nil {
		t.Fatal(err)
	}

	legacyTables := []struct {
		table   string
		textCol string
		idCol   string
		legacy  string
	}{
		{table: "usuario", textCol: "cargo", idCol: "cargo_id", legacy: " Auxiliar Compras Oriente "},
		{table: "paso_ruta", textCol: "cargo", idCol: "cargo_id", legacy: "Analista Compras Oriente"},
		{table: "tarea", textCol: "cargo_asignado", idCol: "cargo_asignado_id", legacy: "Analista Compras Norte"},
		{table: "documento_radicado", textCol: "cargo_actual", idCol: "cargo_actual_id", legacy: "Auxiliar Compras Norte"},
		{table: "regla_monto_ruta", textCol: "cargo_aprobador", idCol: "cargo_aprobador_id", legacy: "Otro cargo"},
	}

	for _, table := range legacyTables {
		statement := "CREATE TABLE " + table.table + " (id integer primary key, " + table.textCol + " text, " + table.idCol + " integer, deleted_at datetime, updated_at datetime)"
		if err := database.Exec(statement).Error; err != nil {
			t.Fatal(err)
		}
		if err := database.Exec("INSERT INTO "+table.table+" (id, "+table.textCol+") VALUES (1, ?)", table.legacy).Error; err != nil {
			t.Fatal(err)
		}
	}

	if err := MigrateCargoIDs(database); err != nil {
		t.Fatal(err)
	}
	if err := MigrateCargoIDs(database); err != nil {
		t.Fatal(err)
	}

	for _, table := range legacyTables {
		var got uint
		if err := database.Table(table.table).Select(table.idCol).Where("id = ?", 1).Scan(&got).Error; err != nil {
			t.Fatal(err)
		}
		if got == 0 {
			t.Errorf("%s.%s was not populated", table.table, table.idCol)
		}
	}

	var assignedID uint
	if err := database.Table("usuario").Select("cargo_id").Where("id = ?", 1).Scan(&assignedID).Error; err != nil {
		t.Fatal(err)
	}
	var cargo Cargo
	if err := database.First(&cargo, assignedID).Error; err != nil {
		t.Fatal(err)
	}
	if cargo.Nombre != "Auxiliar Compras Oriente" {
		t.Fatalf("trimmed cargo name = %q", cargo.Nombre)
	}
}

type legacyUsuarioEmail struct {
	ID    uint   `gorm:"primaryKey"`
	Email string `gorm:"uniqueIndex"`
}

func (legacyUsuarioEmail) TableName() string { return "usuario" }

func TestMigrateUsuarioEmailIndexAllowsDuplicateEmails(t *testing.T) {
	database, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := database.AutoMigrate(&legacyUsuarioEmail{}); err != nil {
		t.Fatal(err)
	}
	if err := MigrateUsuarioEmailIndex(database); err != nil {
		t.Fatal(err)
	}

	if err := database.AutoMigrate(&Usuario{}); err != nil {
		t.Fatal(err)
	}
	users := []Usuario{
		{Nombre: "Persona A", Email: "compartido@example.com"},
		{Nombre: "Persona B", Email: "compartido@example.com"},
	}
	if err := database.Create(&users).Error; err != nil {
		t.Fatalf("duplicate email should be allowed: %v", err)
	}

	idA := "erp-101"
	idB := "erp-102"
	users[0].IDExterno = &idA
	users[1].IDExterno = &idA
	if err := database.Create(&users).Error; err == nil {
		t.Fatal("duplicate external ID should be rejected")
	}
	users[1].IDExterno = &idB
	if err := database.Save(&users[1]).Error; err != nil {
		t.Fatalf("distinct external ID should be allowed: %v", err)
	}
}
