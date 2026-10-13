package documento_radicado

import (
	"sigefae/internal/db"
)

// ListMisTareas retorna los radicados en los que el usuario está involucrado actualmente
// (por ID o por cargo) o en los que participó previamente (basado en trazabilidad).
func (s *Service) ListMisTareas(user db.Usuario) ([]db.DocumentoRadicado, error) {
	var radicados []db.DocumentoRadicado

	query := s.db.
		Preload("DocumentoComercial").
		Preload("DocumentoComercial.Proveedor").
		Preload("DocumentoComercial.Receptor").
		Preload("TipoRadicacion").
		Preload("Ruta").
		Preload("MetodoPago").
		Preload("Estado").
		Preload("PasoActual").
		Preload("UsuarioActual").
		Preload("CargoActual").
		Preload("Archivos").
		Preload("Qr").
		Preload("NormasReparto").
		Preload("NormasReparto.NormaReparto").
		Joins("LEFT JOIN trazabilidad tz ON tz.documento_radicado_id = documento_radicado.id")

	if user.CargoID != nil {
		query = query.Where("documento_radicado.usuario_actual_id = ? OR documento_radicado.cargo_actual_id = ? OR tz.usuario_id = ?", user.ID, *user.CargoID, user.ID)
	} else {
		query = query.Where("documento_radicado.usuario_actual_id = ? OR tz.usuario_id = ?", user.ID, user.ID)
	}

	if err := query.
		Group("documento_radicado.id").
		Order("documento_radicado.fecha_radicacion DESC").
		Find(&radicados).Error; err != nil {
		return nil, err
	}
	return radicados, nil
}
