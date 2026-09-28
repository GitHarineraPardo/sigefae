package auth

import (
	"encoding/json"
	"testing"

	"sigefae/internal/db"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestLoginAndSSOResolveDuplicateEmailsByExternalID(t *testing.T) {
	database, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := database.AutoMigrate(&db.Rol{}, &db.Cargo{}, &db.Usuario{}); err != nil {
		t.Fatal(err)
	}

	role := db.Rol{Nombre: "Aprobador", Activo: true}
	cargo := db.Cargo{Nombre: "Analista", Activo: true}
	if err := database.Create(&role).Error; err != nil {
		t.Fatal(err)
	}
	if err := database.Create(&cargo).Error; err != nil {
		t.Fatal(err)
	}

	passwordA, err := Hash("password-a")
	if err != nil {
		t.Fatal(err)
	}
	passwordB, err := Hash("password-b")
	if err != nil {
		t.Fatal(err)
	}
	externalIDA := "erp-user-101"
	externalIDB := "erp-user-102"
	users := []db.Usuario{
		{Nombre: "Persona A", Email: "compartido@example.com", IDExterno: &externalIDA, HashContrasena: passwordA, RolID: role.ID, CargoID: &cargo.ID, Activo: true},
		{Nombre: "Persona B", Email: "compartido@example.com", IDExterno: &externalIDB, HashContrasena: passwordB, RolID: role.ID, CargoID: &cargo.ID, Activo: true},
	}
	if err := database.Create(&users).Error; err != nil {
		t.Fatal(err)
	}

	service := New(database)
	user, _, err := service.Login("", externalIDB, "password-b")
	if err != nil {
		t.Fatalf("login by external ID failed: %v", err)
	}
	if user.ID != users[1].ID {
		t.Fatalf("login resolved user %d, want %d", user.ID, users[1].ID)
	}

	user, _, err = service.SSOLogin(externalIDA)
	if err != nil {
		t.Fatalf("SSO by external ID failed: %v", err)
	}
	if user.ID != users[0].ID {
		t.Fatalf("SSO resolved user %d, want %d", user.ID, users[0].ID)
	}

	if _, _, err := service.Login("compartido@example.com", "", "password-a"); err == nil {
		t.Fatal("login by duplicate email should require an external ID")
	}
}

func TestSSORequestAcceptsNumericAndTextExternalIDs(t *testing.T) {
	for _, input := range []string{
		`{"id_externo": 42, "secret": "test"}`,
		`{"identifier": "42", "secret": "test"}`,
	} {
		var request SSORequest
		if err := json.Unmarshal([]byte(input), &request); err != nil {
			t.Fatalf("could not decode %s: %v", input, err)
		}
		idExterno := string(request.IDExterno)
		if idExterno == "" {
			idExterno = string(request.Identifier)
		}
		if idExterno != "42" {
			t.Errorf("external ID = %q, want 42", idExterno)
		}
	}
}
