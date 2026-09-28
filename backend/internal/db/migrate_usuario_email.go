package db

import (
	"strings"

	"gorm.io/gorm"
)

// MigrateUsuarioEmailIndex removes legacy single-column unique indexes from usuario.email.
func MigrateUsuarioEmailIndex(database *gorm.DB) error {
	indexes, err := database.Migrator().GetIndexes(&Usuario{})
	if err != nil {
		return err
	}

	for _, index := range indexes {
		unique, ok := index.Unique()
		columns := index.Columns()
		if ok && unique && len(columns) == 1 && strings.EqualFold(columns[0], "email") {
			if err := database.Migrator().DropIndex(&Usuario{}, index.Name()); err != nil {
				return err
			}
		}
	}
	return nil
}
