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

func TestRadicarDocumento_RespetarNormasEnBlanco(t *testing.T) {
	database, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	err = database.AutoMigrate(
		&db.DocumentoComercial{},
		&db.DocumentoRadicado{},
		&db.TipoRadicacion{},
		&db.Ruta{},
		&db.PasoRuta{},
		&db.MetodoPago{},
		&db.EstadoDocumentoRadicado{},
		&db.EstadoTarea{},
		&db.Tarea{},
		&db.CodigoQr{},
		&db.Trazabilidad{},
		&db.ProveedorNormaReparto{},
		&db.RadicadoNormaReparto{},
		&db.NormaReparto{},
		&db.RadicadoActivoFijo{},
		&db.SalarioMinimo{},
		&db.ReglaMontoRuta{},
		&db.ReglaNITRuta{},
		&db.Proveedor{},
		&db.Usuario{},
		&db.Cargo{},
		&db.Archivo{},
		&db.ArchivoOrigen{},
		&db.Receptor{},
		&db.Area{},
		&db.Moneda{},
		&db.DetalleDocumentoComercial{},
	)
	if err != nil {
		t.Fatal(err)
	}

	// 1. Fixtures
	prov := db.Proveedor{ID: 10, RazonSocial: "Proveedor Test", NumeroDocumento: "900123456"}
	database.Create(&prov)

	docCom := db.DocumentoComercial{
		ID:              1,
		NumeroDocumento: "FACT-123",
		IDProveedor:     prov.ID,
		Total:           1000000,
		Subtotal:        840000,
	}
	database.Create(&docCom)

	tipoRad := db.TipoRadicacion{ID: 1, Nombre: "General"}
	database.Create(&tipoRad)

	ruta := db.Ruta{ID: 1, Nombre: "Ruta Compras"}
	database.Create(&ruta)

	metodoPago := db.MetodoPago{ID: 1, Nombre: "Transferencia"}
	database.Create(&metodoPago)

	estadoRad := db.EstadoDocumentoRadicado{ID: 1, Nombre: "Radicado"}
	database.Create(&estadoRad)

	estadoTar1 := db.EstadoTarea{ID: 1, Nombre: "Pendiente"}
	database.Create(&estadoTar1)
	estadoTar2 := db.EstadoTarea{ID: 2, Nombre: "En Proceso"}
	database.Create(&estadoTar2)

	cargoOriente := db.Cargo{ID: 1, Nombre: "Auxiliar Admon Oriente"}
	database.Create(&cargoOriente)

	pasoRuta := db.PasoRuta{
		ID:     1,
		RutaID: ruta.ID,
		Nombre: "Paso 1",
		Orden:  1,
		Activo: true,
	}
	database.Create(&pasoRuta)

	// Guardar normas por defecto del proveedor para esta ruta
	pnr := db.ProveedorNormaReparto{
		ProveedorID:    prov.ID,
		RutaID:         ruta.ID,
		NormaRepartoID: 55,
		Porcentaje:     100,
	}
	database.Create(&pnr)

	svc := New(database)

	// 2. Radicar con NormasReparto vacío (el usuario eligió "dejar en blanco")
	dtoEnBlanco := CreateDTO{
		NumeroRadicado:       "RAD-TEST-001",
		DocumentoComercialID: docCom.ID,
		TipoRadicacionID:     tipoRad.ID,
		RutaID:               ruta.ID,
		MetodoPagoID:         metodoPago.ID,
		NormasReparto:        []NormaRepartoInputDTO{}, // En blanco
	}

	radicado, err := svc.Create(dtoEnBlanco, 1)
	if err != nil {
		t.Fatalf("error creando radicado: %v", err)
	}

	// 3. Verificar que NO se crearon normas de reparto en el radicado
	var normasEnBD []db.RadicadoNormaReparto
	database.Where("documento_radicado_id = ?", radicado.ID).Find(&normasEnBD)

	if len(normasEnBD) != 0 {
		t.Fatalf("se crearon %d normas de reparto; se esperaba 0 porque el usuario eligió dejar en blanco", len(normasEnBD))
	}

	// 4. Radicar con NormasReparto explicitas (el usuario eligió "Cargar automáticamente" o configuró normas)
	docCom2 := db.DocumentoComercial{
		ID:              2,
		NumeroDocumento: "FACT-124",
		IDProveedor:     prov.ID,
		Total:           2000000,
		Subtotal:        1680000,
	}
	database.Create(&docCom2)

	dtoConNormas := CreateDTO{
		NumeroRadicado:       "RAD-TEST-002",
		DocumentoComercialID: docCom2.ID,
		TipoRadicacionID:     tipoRad.ID,
		RutaID:               ruta.ID,
		MetodoPagoID:         metodoPago.ID,
		NormasReparto: []NormaRepartoInputDTO{
			{
				NormaRepartoID: 55,
				Porcentaje:     100,
				Proyecto:       "PROY-1",
			},
		},
	}

	radicado2, err := svc.Create(dtoConNormas, 1)
	if err != nil {
		t.Fatalf("error creando segundo radicado: %v", err)
	}

	var normasGuardadas []db.RadicadoNormaReparto
	database.Where("documento_radicado_id = ?", radicado2.ID).Find(&normasGuardadas)

	if len(normasGuardadas) != 1 {
		t.Fatalf("se crearon %d normas; se esperaba 1", len(normasGuardadas))
	}
	if normasGuardadas[0].NormaRepartoID != 55 || normasGuardadas[0].Porcentaje != 100 {
		t.Fatalf("norma guardada no coincide: %+v", normasGuardadas[0])
	}
}


