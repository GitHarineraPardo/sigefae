package api

import (
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestNewRegistersCargoRoutes(t *testing.T) {
	database, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}

	router := New(database)
	registered := make(map[string]bool)
	for _, route := range router.Routes() {
		registered[route.Method+" "+route.Path] = true
	}

	for _, route := range []string{
		"GET /api/cargos",
		"GET /api/cargos/activos",
		"GET /api/cargos/todos",
		"POST /api/cargos",
		"PATCH /api/cargos/:id",
		"PATCH /api/cargos/:id/activo",
		"DELETE /api/cargos/:id",
	} {
		if !registered[route] {
			t.Errorf("cargo route %q was not registered", route)
		}
	}
}
