package evento_dian

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"sigefae/internal/db"
)

func setupTestDB(t *testing.T) *gorm.DB {
	database, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	err = database.AutoMigrate(
		&db.Rol{},
		&db.Cargo{},
		&db.Usuario{},
		&db.Receptor{},
		&db.DocumentoComercial{},
		&db.DocumentoRadicado{},
		&db.EstadoTarea{},
		&db.Tarea{},
		&db.EventoRadian{},
		&db.Trazabilidad{},
	)
	assert.NoError(t, err)

	return database
}

func TestEmitirEvento_Permisos(t *testing.T) {
	database := setupTestDB(t)
	svc := NewService(database, nil)

	// Crear rol y cargo diferente
	rolAprobador := db.Rol{Nombre: "Aprobador"}
	database.Create(&rolAprobador)
	cargoAnalista := db.Cargo{Nombre: "Analista Compras", Activo: true}
	database.Create(&cargoAnalista)

	userNoAutorizado := db.Usuario{
		Nombre: "Juan Perez",
		Email:  "juan@test.com",
		RolID:  rolAprobador.ID,
		CargoID: &cargoAnalista.ID,
		Activo:  true,
	}
	database.Create(&userNoAutorizado)

	// Intentar emitir evento con usuario no autorizado
	_, err := svc.EmitirEvento(1, userNoAutorizado.ID, EmitirEventoDTO{Tipo: "030"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "solo el cargo Auxiliar Admon Oriente")

	// Superadministrador sin el cargo Auxiliar Admon Oriente tampoco debe poder emitir
	rolSuperAdmin := db.Rol{Nombre: "Superadministrador"}
	database.Create(&rolSuperAdmin)
	userAdminSinCargo := db.Usuario{
		Nombre:  "Admin Sistema",
		Email:   "admin@test.com",
		RolID:   rolSuperAdmin.ID,
		CargoID: &cargoAnalista.ID,
		Activo:  true,
	}
	database.Create(&userAdminSinCargo)

	_, errAdmin := svc.EmitirEvento(1, userAdminSinCargo.ID, EmitirEventoDTO{Tipo: "030"})
	assert.Error(t, errAdmin)
	assert.Contains(t, errAdmin.Error(), "solo el cargo Auxiliar Admon Oriente")
}

func TestEmitirEvento_RutaIncompleta(t *testing.T) {
	database := setupTestDB(t)
	svc := NewService(database, nil)

	cargoAux := db.Cargo{Nombre: "Auxiliar Admon Oriente", Activo: true}
	database.Create(&cargoAux)
	rolUsuario := db.Rol{Nombre: "Usuario"}
	database.Create(&rolUsuario)

	userAux := db.Usuario{
		Nombre: "Maria Auxiliar",
		Email:  "maria@test.com",
		RolID:  rolUsuario.ID,
		CargoID: &cargoAux.ID,
		Activo:  true,
	}
	database.Create(&userAux)

	ldfVal := "LDF-TEST-123"
	docCom := db.DocumentoComercial{
		NumeroDocumento: "FACT-001",
		LDF:             &ldfVal,
		Activo:          true,
	}
	database.Create(&docCom)

	rad := db.DocumentoRadicado{
		NumeroRadicado:       "RAD-2026-001",
		DocumentoComercialID: docCom.ID,
		EstadoPosesion:       "EnProceso", // Ruta no completada
	}
	database.Create(&rad)

	estadoPendiente := db.EstadoTarea{Nombre: "Pendiente", Activo: true}
	database.Create(&estadoPendiente)

	tarea := db.Tarea{
		DocumentoRadicadoID: rad.ID,
		EstadoID:            estadoPendiente.ID,
		Descripcion:         "Tarea pendiente",
	}
	database.Create(&tarea)

	// Intentar emitir cuando la ruta sigue en proceso
	_, err := svc.EmitirEvento(rad.ID, userAux.ID, EmitirEventoDTO{Tipo: "032"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ruta de aprobación esté completamente finalizada")
}

func TestEmitirEvento_Exitoso(t *testing.T) {
	database := setupTestDB(t)
	svc := NewService(database, nil)

	cargoAux := db.Cargo{Nombre: "Auxiliar Admon Oriente", Activo: true}
	database.Create(&cargoAux)
	rolUsuario := db.Rol{Nombre: "Usuario"}
	database.Create(&rolUsuario)

	userAux := db.Usuario{
		Nombre: "Maria Auxiliar",
		Email:  "maria@test.com",
		RolID:  rolUsuario.ID,
		CargoID: &cargoAux.ID,
		Activo:  true,
	}
	database.Create(&userAux)

	ldfVal := "LDF-12345"
	receptor := db.Receptor{
		NumeroDocumento: "890116102",
		Nombre:          "HARINERA PARDO S.A.S.",
		Activo:          true,
	}
	database.Create(&receptor)

	docCom := db.DocumentoComercial{
		NumeroDocumento: "SETT-100",
		LDF:             &ldfVal,
		IDReceptor:      receptor.ID,
		Activo:          true,
	}
	database.Create(&docCom)

	rad := db.DocumentoRadicado{
		NumeroRadicado:       "RAD-100",
		DocumentoComercialID: docCom.ID,
		EstadoPosesion:       "Completado",
	}
	database.Create(&rad)

	// Emitir Acuse (030)
	resp, err := svc.EmitirEvento(rad.ID, userAux.ID, EmitirEventoDTO{
		Tipo:        "030",
		Observacion: "Acuse de recibo de prueba",
	})
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "030", resp.CodigoDIAN)
	assert.Equal(t, "ACKNOWLEDGED", resp.TipoEvento)
	assert.Equal(t, "Maria Auxiliar", resp.UsuarioNombre)

	// Verificar que se actualizó el doc comercial
	var docActualizado db.DocumentoComercial
	database.First(&docActualizado, docCom.ID)
	assert.Equal(t, "ACKNOWLEDGED", docActualizado.EventoDIAN)

	// Listar eventos
	lista, err := svc.ListByRadicadoID(rad.ID)
	assert.NoError(t, err)
	assert.Len(t, lista, 1)
	assert.Equal(t, "030", lista[0].CodigoDIAN)

	// Intentar emitir Aceptación (034) antes de Recibo de Bienes (032) -> Debe fallar
	_, errAcepPrevia := svc.EmitirEvento(rad.ID, userAux.ID, EmitirEventoDTO{
		Tipo: "034",
	})
	assert.Error(t, errAcepPrevia)
	assert.Contains(t, errAcepPrevia.Error(), "sin haber registrado previamente el Recibo de Bienes o Servicios (032)")

	// Emitir Recibo de Bienes (032) -> Debe ser exitoso
	resp2, err := svc.EmitirEvento(rad.ID, userAux.ID, EmitirEventoDTO{
		Tipo:        "032",
		Observacion: "Mercancía recibida en planta",
	})
	assert.NoError(t, err)
	assert.Equal(t, "032", resp2.CodigoDIAN)
	assert.Equal(t, "RECEIVEDGOODS", resp2.TipoEvento)

	// Emitir Aceptación (034) tras haber recibido bienes -> Debe ser exitoso
	resp3, err := svc.EmitirEvento(rad.ID, userAux.ID, EmitirEventoDTO{
		Tipo:        "034",
		Observacion: "Factura aprobada y aceptada",
	})
	assert.NoError(t, err)
	assert.Equal(t, "034", resp3.CodigoDIAN)
	assert.Equal(t, "ACCEPTED", resp3.TipoEvento)

	// Listar nuevamente: debe tener 3 eventos ordenados cronológicamente
	lista2, err := svc.ListByRadicadoID(rad.ID)
	assert.NoError(t, err)
	assert.Len(t, lista2, 3)
}

func TestEmitirEvento_FacturaVieneAcusadaInbound(t *testing.T) {
	database := setupTestDB(t)
	svc := NewService(database, nil)

	cargoAux := db.Cargo{Nombre: "Auxiliar Admon Oriente", Activo: true}
	database.Create(&cargoAux)
	rolUsuario := db.Rol{Nombre: "Usuario"}
	database.Create(&rolUsuario)

	userAux := db.Usuario{
		Nombre:  "Jennyfer Auxiliar",
		Email:   "jennyfer@test.com",
		RolID:   rolUsuario.ID,
		CargoID: &cargoAux.ID,
		Activo:  true,
	}
	database.Create(&userAux)

	ldfVal := "LDF-INBOUND-999"
	receptor := db.Receptor{
		NumeroDocumento: "890116102",
		Nombre:          "HARINERA PARDO S.A.S.",
		Activo:          true,
	}
	database.Create(&receptor)

	// Factura viene acusada por Inbound (EventoDIAN: ACKNOWLEDGED, sin fila en evento_radian)
	docCom := db.DocumentoComercial{
		NumeroDocumento: "SETT-INBOUND",
		LDF:             &ldfVal,
		IDReceptor:      receptor.ID,
		EventoDIAN:      "ACKNOWLEDGED",
		Activo:          true,
	}
	database.Create(&docCom)

	rad := db.DocumentoRadicado{
		NumeroRadicado:       "RAD-INBOUND",
		DocumentoComercialID: docCom.ID,
		EstadoPosesion:       "Completado",
	}
	database.Create(&rad)

	// 1. ListByRadicadoID debe reconocer el evento 030
	evs1, err := svc.ListByRadicadoID(rad.ID)
	assert.NoError(t, err)
	assert.Len(t, evs1, 1)
	assert.Equal(t, "030", evs1[0].CodigoDIAN)

	// 2. No debe permitir emitir 030 nuevamente
	_, errDuplicado := svc.EmitirEvento(rad.ID, userAux.ID, EmitirEventoDTO{Tipo: "030"})
	assert.Error(t, errDuplicado)
	assert.Contains(t, errDuplicado.Error(), "ya fue emitido o registrado previamente")

	// 3. Emitir Recibo de Bienes (032)
	resp032, err := svc.EmitirEvento(rad.ID, userAux.ID, EmitirEventoDTO{
		Tipo:        "032",
		Observacion: "Bienes recibidos a satisfacción",
	})
	assert.NoError(t, err)
	assert.Equal(t, "032", resp032.CodigoDIAN)

	// 4. Tras emitir 032, el evento 030 NO DEBE DESAPARECER. Deben estar 030 y 032!
	evs2, err := svc.ListByRadicadoID(rad.ID)
	assert.NoError(t, err)
	assert.Len(t, evs2, 2)
	assert.Equal(t, "030", evs2[0].CodigoDIAN)
	assert.Equal(t, "032", evs2[1].CodigoDIAN)

	// 5. Ahora emitir Aceptación (034)
	resp034, err := svc.EmitirEvento(rad.ID, userAux.ID, EmitirEventoDTO{
		Tipo:        "034",
		Observacion: "Factura aprobada y aceptada",
	})
	assert.NoError(t, err)
	assert.Equal(t, "034", resp034.CodigoDIAN)

	// 6. Lista final debe tener los 3 eventos: 030, 032 y 034
	evs3, err := svc.ListByRadicadoID(rad.ID)
	assert.NoError(t, err)
	assert.Len(t, evs3, 3)
	assert.Equal(t, "030", evs3[0].CodigoDIAN)
	assert.Equal(t, "032", evs3[1].CodigoDIAN)
	assert.Equal(t, "034", evs3[2].CodigoDIAN)
}

