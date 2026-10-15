package gasto_proveedor

import (
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"sigefae/internal/db"
)

func setupTestDB(t *testing.T) *gorm.DB {
	database, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect to in-memory database: %v", err)
	}

	err = database.AutoMigrate(
		&db.Rol{},
		&db.Cargo{},
		&db.Usuario{},
		&db.Proveedor{},
		&db.DocumentoComercial{},
		&db.DetalleDocumentoComercial{},
		&db.DocumentoRadicado{},
		&db.Tarea{},
		&db.Trazabilidad{},
		&db.NormaReparto{},
		&db.RadicadoNormaReparto{},
	)
	if err != nil {
		t.Fatalf("failed to migrate models: %v", err)
	}

	return database
}

func TestGetGastosPorCargo(t *testing.T) {
	database := setupTestDB(t)
	svc := New(database)

	cargoLogistica := db.Cargo{Nombre: "Auxiliar Logística", Activo: true}
	database.Create(&cargoLogistica)

	cargoCompras := db.Cargo{Nombre: "Jefe de Compras", Activo: true}
	database.Create(&cargoCompras)

	userLogistica := db.Usuario{Nombre: "Juan Logística", CargoID: &cargoLogistica.ID, Activo: true}
	database.Create(&userLogistica)

	prov1 := db.Proveedor{RazonSocial: "Empaques y Cajas S.A.", NumeroDocumento: "900111222-1"}
	database.Create(&prov1)

	prov2 := db.Proveedor{RazonSocial: "Transportes del Oriente", NumeroDocumento: "900333444-2"}
	database.Create(&prov2)

	doc1 := db.DocumentoComercial{
		NumeroDocumento: "FAC-001",
		IDProveedor:     prov1.ID,
		Total:           1000000,
		Subtotal:        840336,
		Iva:             159664,
		FechaDocumento:  time.Now(),
		Activo:          true,
	}
	database.Create(&doc1)

	det1 := db.DetalleDocumentoComercial{
		DocumentoComercialID: doc1.ID,
		Descripcion:          "Cajas de cartón corrugado",
		Cantidad:             100,
		ValorUnit:            10000,
		Total:                1000000,
		Activo:               true,
	}
	database.Create(&det1)

	doc2 := db.DocumentoComercial{
		NumeroDocumento: "FAC-002",
		IDProveedor:     prov2.ID,
		Total:           2000000,
		Subtotal:        2000000,
		Iva:             0,
		FechaDocumento:  time.Now(),
		Activo:          true,
	}
	database.Create(&doc2)

	rad1 := db.DocumentoRadicado{
		NumeroRadicado:       "RAD-1001",
		DocumentoComercialID: doc1.ID,
		EstadoPosesion:       "Completado",
		FechaRadicacion:      time.Now(),
	}
	database.Create(&rad1)

	// Tarea aprobada por Auxiliar Logística en rad1
	tarea1 := db.Tarea{
		DocumentoRadicadoID: rad1.ID,
		CargoAsignadoID:     &cargoLogistica.ID,
		EstadoID:            3, // Completada
	}
	database.Create(&tarea1)

	rad2 := db.DocumentoRadicado{
		NumeroRadicado:       "RAD-1002",
		DocumentoComercialID: doc2.ID,
		EstadoPosesion:       "Completado",
		FechaRadicacion:      time.Now(),
	}
	database.Create(&rad2)

	// Tarea en rad2 aprobada SOLO por Jefe de Compras, no por Logística
	tarea2 := db.Tarea{
		DocumentoRadicadoID: rad2.ID,
		CargoAsignadoID:     &cargoCompras.ID,
		EstadoID:            3,
	}
	database.Create(&tarea2)

	// Test: Usuario de logística consulta sus gastos
	resp, err := svc.GetGastos(userLogistica, "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.CantidadFacturas != 1 {
		t.Fatalf("esperaba 1 factura aprobada por logística, obtuvo %d", resp.CantidadFacturas)
	}

	if resp.TotalGeneral != 1000000 {
		t.Fatalf("esperaba total 1000000, obtuvo %f", resp.TotalGeneral)
	}

	if len(resp.Proveedores) != 1 {
		t.Fatalf("esperaba 1 proveedor, obtuvo %d", len(resp.Proveedores))
	}

	if resp.Proveedores[0].RazonSocial != "Empaques y Cajas S.A." {
		t.Fatalf("proveedor inesperado: %s", resp.Proveedores[0].RazonSocial)
	}

	if len(resp.Items) != 1 {
		t.Fatalf("esperaba 1 ítem en desglose, obtuvo %d", len(resp.Items))
	}

	if resp.Items[0].Descripcion != "Cajas de cartón corrugado" {
		t.Fatalf("ítem inesperado: %s", resp.Items[0].Descripcion)
	}
}

func TestGetGastosAdminYFechas(t *testing.T) {
	database := setupTestDB(t)
	svc := New(database)

	cargo1 := db.Cargo{Nombre: "Mantenimiento", Activo: true}
	database.Create(&cargo1)

	adminRole := db.Rol{Nombre: "Superadministrador", Activo: true}
	database.Create(&adminRole)

	userAdmin := db.Usuario{Nombre: "Admin User", RolID: adminRole.ID, Rol: &adminRole, CargoID: &cargo1.ID, Activo: true}
	database.Create(&userAdmin)

	cargo2 := db.Cargo{Nombre: "Compras", Activo: true}
	database.Create(&cargo2)

	prov := db.Proveedor{RazonSocial: "Ferretería Central", NumeroDocumento: "800555666-3"}
	database.Create(&prov)

	prov2 := db.Proveedor{RazonSocial: "Proveedor de Compras", NumeroDocumento: "800555667-4"}
	database.Create(&prov2)

	doc := db.DocumentoComercial{
		NumeroDocumento: "FAC-999",
		IDProveedor:     prov.ID,
		Total:           500000,
		FechaDocumento:  time.Now(),
		Activo:          true,
	}
	database.Create(&doc)

	rad := db.DocumentoRadicado{
		NumeroRadicado:       "RAD-999",
		DocumentoComercialID: doc.ID,
		EstadoPosesion:       "Completado",
		FechaRadicacion:      time.Date(2026, 5, 10, 10, 0, 0, 0, time.UTC),
	}
	database.Create(&rad)

	doc2 := db.DocumentoComercial{
		NumeroDocumento: "FAC-1000",
		IDProveedor:     prov2.ID,
		Total:           900000,
		FechaDocumento:  time.Now(),
		Activo:          true,
	}
	database.Create(&doc2)

	rad2 := db.DocumentoRadicado{
		NumeroRadicado:       "RAD-1000",
		DocumentoComercialID: doc2.ID,
		EstadoPosesion:       "Completado",
		FechaRadicacion:      time.Date(2026, 5, 11, 10, 0, 0, 0, time.UTC),
	}
	database.Create(&rad2)

	tarea := db.Tarea{
		DocumentoRadicadoID: rad.ID,
		CargoAsignadoID:     &cargo1.ID,
		EstadoID:            3,
	}
	database.Create(&tarea)

	tarea2 := db.Tarea{
		DocumentoRadicadoID: rad2.ID,
		CargoAsignadoID:     &cargo2.ID,
		EstadoID:            3,
	}
	database.Create(&tarea2)

	// El cargo del usuario determina el alcance, incluso para un superadministrador.
	respAdmin, err := svc.GetGastos(userAdmin, "", "")
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if respAdmin.CantidadFacturas != 1 {
		t.Fatalf("esperaba 1 factura para admin, obtuvo %d", respAdmin.CantidadFacturas)
	}

	// Filtro por fecha fuera de rango
	respFueraRango, err := svc.GetGastos(userAdmin, "2026-06-01", "2026-06-30")
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if respFueraRango.CantidadFacturas != 0 {
		t.Fatalf("esperaba 0 facturas fuera de rango, obtuvo %d", respFueraRango.CantidadFacturas)
	}

	// Filtro por fecha dentro de rango
	respEnRango, err := svc.GetGastos(userAdmin, "2026-05-01", "2026-05-31")
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if respEnRango.CantidadFacturas != 1 {
		t.Fatalf("esperaba 1 factura dentro de rango, obtuvo %d", respEnRango.CantidadFacturas)
	}
}

func TestGetGastosSinCargoNoExponeOtrosCargos(t *testing.T) {
	database := setupTestDB(t)
	svc := New(database)

	cargo := db.Cargo{Nombre: "Compras", Activo: true}
	database.Create(&cargo)
	adminRole := db.Rol{Nombre: "Superadministrador", Activo: true}
	database.Create(&adminRole)
	user := db.Usuario{Nombre: "Admin sin cargo", RolID: adminRole.ID, Rol: &adminRole, Activo: true}
	database.Create(&user)

	proveedor := db.Proveedor{RazonSocial: "Proveedor privado", NumeroDocumento: "900000001"}
	database.Create(&proveedor)
	documento := db.DocumentoComercial{
		NumeroDocumento: "FAC-PRIVADA",
		IDProveedor:     proveedor.ID,
		Total:           100000,
		FechaDocumento:  time.Now(),
		Activo:          true,
	}
	database.Create(&documento)
	radicado := db.DocumentoRadicado{
		NumeroRadicado:       "RAD-PRIVADO",
		DocumentoComercialID: documento.ID,
		EstadoPosesion:       "Completado",
		FechaRadicacion:      time.Now(),
	}
	database.Create(&radicado)
	tarea := db.Tarea{
		DocumentoRadicadoID: radicado.ID,
		CargoAsignadoID:     &cargo.ID,
		UsuarioAsignadoID:   &user.ID,
		EstadoID:            3,
	}
	database.Create(&tarea)

	resp, err := svc.GetGastos(user, "", "")
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if resp.CantidadFacturas != 0 {
		t.Fatalf("esperaba 0 facturas sin cargo asignado, obtuvo %d", resp.CantidadFacturas)
	}
	if resp.CargoNombre != "Sin cargo asignado" {
		t.Fatalf("cargo inesperado para usuario sin cargo: %q", resp.CargoNombre)
	}
}
