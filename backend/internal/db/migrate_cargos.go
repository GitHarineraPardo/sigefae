package db

import (
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
)

type cargoLegacySource struct {
	model      any
	legacyName string
	idName     string
}

// MigrateCargoIDs copies legacy cargo text into relational cargo references.
// It is safe to run repeatedly and intentionally leaves the legacy columns intact.
func MigrateCargoIDs(database *gorm.DB) error {
	sources := []cargoLegacySource{
		{model: &Usuario{}, legacyName: "cargo", idName: "cargo_id"},
		{model: &PasoRuta{}, legacyName: "cargo", idName: "cargo_id"},
		{model: &Tarea{}, legacyName: "cargo_asignado", idName: "cargo_asignado_id"},
		{model: &DocumentoRadicado{}, legacyName: "cargo_actual", idName: "cargo_actual_id"},
		{model: &ReglaMontoRuta{}, legacyName: "cargo_aprobador", idName: "cargo_aprobador_id"},
	}

	return database.Transaction(func(tx *gorm.DB) error {
		ensureCargo := func(name string) (Cargo, error) {
			var cargo Cargo
			err := tx.Unscoped().Where("nombre = ?", name).First(&cargo).Error
			if errors.Is(err, gorm.ErrRecordNotFound) {
				cargo = Cargo{Nombre: name, Activo: true, CreatedAt: time.Now(), UpdatedAt: time.Now()}
				err = tx.Create(&cargo).Error
			} else if err == nil && cargo.DeletedAt.Valid {
				cargo.DeletedAt = gorm.DeletedAt{}
				cargo.Activo = true
				err = tx.Unscoped().Save(&cargo).Error
			}
			return cargo, err
		}

		for _, name := range []string{
			"Auxiliar Compras Oriente",
			"Auxiliar Compras Norte",
			"Analista Compras Norte",
			"Analista Compras Oriente",
		} {
			if _, err := ensureCargo(name); err != nil {
				return err
			}
		}

		for _, source := range sources {
			if !tx.Migrator().HasTable(source.model) ||
				!tx.Migrator().HasColumn(source.model, source.legacyName) ||
				!tx.Migrator().HasColumn(source.model, source.idName) {
				continue
			}

			var names []string
			query := tx.Model(source.model).
				Distinct("TRIM(" + source.legacyName + ")").
				Where(source.legacyName + " IS NOT NULL AND TRIM(" + source.legacyName + ") <> ''")
			if err := query.Pluck(source.legacyName, &names).Error; err != nil {
				return err
			}

			for _, rawName := range names {
				name := strings.TrimSpace(rawName)
				if name == "" {
					continue
				}

				cargo, err := ensureCargo(name)
				if err != nil {
					return err
				}

				update := "TRIM(" + source.legacyName + ") = ? AND (" + source.idName + " IS NULL OR " + source.idName + " = 0)"
				if err := tx.Model(source.model).Where(update, name).Update(source.idName, cargo.ID).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
}
