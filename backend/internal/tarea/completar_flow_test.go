package tarea

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"sigefae/internal/db"
)

func setupTestDB(t *testing.T) *gorm.DB {
	database, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("error opening sqlite: %v", err)
	}

	err = database.AutoMigrate(
		&db.Rol{},
		&db.Cargo{},
		&db.Usuario{},
		&db.DocumentoRadicado{},
		&db.Tarea{},
		&db.EstadoTarea{},
		&db.Trazabilidad{},
		&db.RadicadoNormaReparto{},
		&db.RadicadoActivoFijo{},
	)
	if err != nil {
		t.Fatalf("error migrating: %v", err)
	}

	// Estados de tarea
	database.Create(&db.EstadoTarea{ID: 1, Nombre: "Pendiente", Activo: true})
	database.Create(&db.EstadoTarea{ID: 2, Nombre: "En Proceso", Activo: true})
	database.Create(&db.EstadoTarea{ID: 3, Nombre: "Completada", Activo: true})

	return database
}

func TestCompletarAuxiliarOrientePrimerTurno(t *testing.T) {
	gin.SetMode(gin.TestMode)
	database := setupTestDB(t)

	cargoAux := db.Cargo{Nombre: "Auxiliar Admon Oriente", Activo: true}
	database.Create(&cargoAux)

	cargoAprobador1 := db.Cargo{Nombre: "Analista de Compras Oriente", Activo: true}
	database.Create(&cargoAprobador1)

	cargoAprobador2 := db.Cargo{Nombre: "Coordinador Logístico Oriente", Activo: true}
	database.Create(&cargoAprobador2)

	rol := db.Rol{Nombre: "Usuario", Activo: true}
	database.Create(&rol)

	userAux := db.Usuario{
		Nombre:   "Jennyfer Osman",
		Email:    "aux@harinera.co",
		RolID:    rol.ID,
		CargoID:  &cargoAux.ID,
		Activo:   true,
	}
	database.Create(&userAux)
	database.Preload("Rol").Preload("Cargo").First(&userAux, userAux.ID)

	radicado := db.DocumentoRadicado{
		NumeroRadicado:  "BU-2026-00001",
		FechaRadicacion: time.Now(),
		EstadoPosesion:  "EnProceso",
		CargoActualID:   &cargoAux.ID,
	}
	database.Create(&radicado)

	// Crear flujo: [Aux Oriente (1)] -> [Analista Compras (2)] -> [Coordinador Log (3)] -> [Aux Oriente Cierre (4)]
	t1 := db.Tarea{DocumentoRadicadoID: radicado.ID, CargoAsignadoID: &cargoAux.ID, EstadoID: 2, Descripcion: "Revisión Auxiliar Admon Oriente"}
	t2 := db.Tarea{DocumentoRadicadoID: radicado.ID, CargoAsignadoID: &cargoAprobador1.ID, EstadoID: 1, Descripcion: "Aprobación Analista Compras"}
	t3 := db.Tarea{DocumentoRadicadoID: radicado.ID, CargoAsignadoID: &cargoAprobador2.ID, EstadoID: 1, Descripcion: "Aprobación Coordinador Logístico"}
	t4 := db.Tarea{DocumentoRadicadoID: radicado.ID, CargoAsignadoID: &cargoAux.ID, EstadoID: 1, Descripcion: "Revisión Auxiliar Admon Oriente"}
	database.Create(&t1)
	database.Create(&t2)
	database.Create(&t3)
	database.Create(&t4)

	handler := NewHandler(New(database), nil, database)
	router := gin.New()
	router.PATCH("/tarea/:id/completar", func(c *gin.Context) {
		c.Set("user", userAux)
		c.Set("user_id", userAux.ID)
		handler.Completar(c)
	})

	// El Auxiliar Oriente decide DESMARCAR t2 (Analista Compras) y DEJAR solo t3 (Coordinador)
	payload := CompletarDTO{
		TareasSeleccionadasIDs: &[]uint{t3.ID},
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("PATCH", "/tarea/1/completar", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("esperaba status 200, obtuvo %d: %s", w.Code, w.Body.String())
	}

	// 1. Tarea 1 debe quedar Completada
	var t1DB db.Tarea
	database.First(&t1DB, t1.ID)
	if t1DB.EstadoID != 3 {
		t.Fatalf("tarea 1 debió quedar completada, estado = %d", t1DB.EstadoID)
	}

	// 2. Tarea 2 debió ser excluida (soft deleted)
	var t2DB db.Tarea
	err := database.First(&t2DB, t2.ID).Error
	if err == nil {
		t.Fatalf("tarea 2 debió ser eliminada (soft delete)")
	}

	// 3. Tarea 3 debe existir y ser activada (En Proceso = 2)
	var t3DB db.Tarea
	if err := database.First(&t3DB, t3.ID).Error; err != nil {
		t.Fatalf("tarea 3 debió mantenerse: %v", err)
	}
	if t3DB.EstadoID != 2 {
		t.Fatalf("tarea 3 debió activarse (En Proceso = 2), estado = %d", t3DB.EstadoID)
	}

	// 4. Tarea 4 (cierre) debe permanecer activa y pendiente
	var t4DB db.Tarea
	if err := database.First(&t4DB, t4.ID).Error; err != nil {
		t.Fatalf("tarea 4 de cierre debió mantenerse: %v", err)
	}
	if t4DB.EstadoID != 1 {
		t.Fatalf("tarea 4 debió seguir pendiente, estado = %d", t4DB.EstadoID)
	}

	// 5. Debe existir registro en trazabilidad
	var traz db.Trazabilidad
	if err := database.Where("documento_radicado_id = ? AND accion = ?", radicado.ID, "Ajuste de Flujo").First(&traz).Error; err != nil {
		t.Fatalf("debió crearse trazabilidad de Ajuste de Flujo: %v", err)
	}
}

func TestCompletarOtroCargoNoPuedeAjustarFlujo(t *testing.T) {
	gin.SetMode(gin.TestMode)
	database := setupTestDB(t)

	cargoOtro := db.Cargo{Nombre: "Analista de Compras Oriente", Activo: true}
	database.Create(&cargoOtro)

	rol := db.Rol{Nombre: "Usuario", Activo: true}
	database.Create(&rol)

	userOtro := db.Usuario{
		Nombre:   "Otro Usuario",
		Email:    "otro@harinera.co",
		RolID:    rol.ID,
		CargoID:  &cargoOtro.ID,
		Activo:   true,
	}
	database.Create(&userOtro)
	database.Preload("Rol").Preload("Cargo").First(&userOtro, userOtro.ID)

	radicado := db.DocumentoRadicado{NumeroRadicado: "BU-2026-00002", EstadoPosesion: "EnProceso"}
	database.Create(&radicado)

	t1 := db.Tarea{DocumentoRadicadoID: radicado.ID, CargoAsignadoID: &cargoOtro.ID, EstadoID: 2, Descripcion: "Paso 1"}
	t2 := db.Tarea{DocumentoRadicadoID: radicado.ID, CargoAsignadoID: &cargoOtro.ID, EstadoID: 1, Descripcion: "Paso 2"}
	database.Create(&t1)
	database.Create(&t2)

	handler := NewHandler(New(database), nil, database)
	router := gin.New()
	router.PATCH("/tarea/:id/completar", func(c *gin.Context) {
		c.Set("user", userOtro)
		c.Set("user_id", userOtro.ID)
		handler.Completar(c)
	})

	// Intenta mandar selección de tareas
	payload := CompletarDTO{
		TareasSeleccionadasIDs: &[]uint{t2.ID},
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("PATCH", "/tarea/1/completar", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("esperaba status 403 Forbidden para otro cargo, obtuvo %d", w.Code)
	}
}
