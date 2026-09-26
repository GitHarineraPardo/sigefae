package db

import (
	"fmt"

	"gorm.io/gorm"
)

func MigrateCargoForeignKeys(database *gorm.DB) error {
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
			if err := database.Migrator().CreateConstraint(relation.model, relation.field); err != nil {
				return fmt.Errorf("no se pudo crear la clave foránea de %T.%s: %w", relation.model, relation.field, err)
			}
		}
	}
	return nil
}
