package db

import "log"

func Migrate() error {
	// Intentar eliminar índices únicos antiguos de columna única 'nombre' en la tabla 'ruta'
	indexesToDrop := []string{
		"nombre",
		"nombre_2",
		"uni_ruta_nombre",
		"idx_ruta_nombre",
		"uk_ruta_nombre",
		"idx_nombre",
	}
	for _, idx := range indexesToDrop {
		if DB.Migrator().HasIndex(&Ruta{}, idx) {
			if err := DB.Migrator().DropIndex(&Ruta{}, idx); err != nil {
				log.Printf("[Migrate] Aviso al eliminar índice antiguo '%s' en tabla ruta: %v", idx, err)
			}
		}
		// Ejecutar también SQL directo para asegurar eliminación en MySQL
		_ = DB.Exec("ALTER TABLE ruta DROP INDEX " + idx).Error
	}

	if err := DB.AutoMigrate(BaseModels()...); err != nil {
		return err
	}
	if err := DB.AutoMigrate(CircularModels()...); err != nil {
		return err
	}
	if err := MigrateCargoIDs(DB); err != nil {
		return err
	}
	return MigrateCargoForeignKeys(DB)
}
