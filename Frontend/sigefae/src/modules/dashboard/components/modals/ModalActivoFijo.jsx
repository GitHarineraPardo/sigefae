import { formatCurrency } from "../../helpers/formatters";

export default function ModalActivoFijo({
  showActivoModal, setShowActivoModal, activoEditandoId,
  activoFormDetalle, setActivoFormDetalle,
  activoModalRadicadoId, proyectosDisponibles, sedesDisponibles,
  handleGuardarActivoFijo, subtotalRadicado
}) {
  if (!showActivoModal) return null;

  const subtotal = parseFloat(subtotalRadicado) || 0;

  const handlePctChange = (e) => {
    const val = e.target.value;
    const newForm = { ...activoFormDetalle, porcentaje: val };
    if (subtotal > 0 && val) {
      newForm.valor = ((parseFloat(val) / 100) * subtotal).toFixed(2);
    } else {
      newForm.valor = "";
    }
    setActivoFormDetalle(newForm);
  };

  const handleValChange = (e) => {
    const val = e.target.value;
    const newForm = { ...activoFormDetalle, valor: val };
    if (subtotal > 0 && val) {
      newForm.porcentaje = ((parseFloat(val) / subtotal) * 100).toFixed(2);
    } else {
      newForm.porcentaje = "";
    }
    setActivoFormDetalle(newForm);
  };

  return (
    <div className="modal-overlay" onClick={() => setShowActivoModal(false)}>
      <div className="modal-content" onClick={(e) => e.stopPropagation()}>
        <div className="modal-header">
          <h3>
            <i className="fa-solid fa-boxes-stacked"></i>{" "}
            {activoEditandoId ? "Editar" : "Agregar"} Activo Fijo
          </h3>
          <button className="modal-close" onClick={() => setShowActivoModal(false)}>
            <i className="fa-solid fa-xmark"></i>
          </button>
        </div>
        <div className="modal-body">
          {subtotal > 0 && (
            <p style={{ fontSize: "0.85em", color: "#64748b", margin: "0 0 12px" }}>
              <i className="fa-solid fa-info-circle"></i> Subtotal:{" "}
              <strong>{formatCurrency(subtotal)}</strong> — Ingresa % o valor, el otro se calcula automáticamente.
            </p>
          )}

          {/* Nombre */}
          <div className="modal-field">
            <label>Nombre del Activo <span className="required">*</span></label>
            <input
              type="text"
              className="doc-input"
              style={{ borderColor: !activoFormDetalle.nombre?.trim() ? "#ef4444" : undefined }}
              value={activoFormDetalle.nombre || ""}
              onChange={(e) => setActivoFormDetalle(prev => ({ ...prev, nombre: e.target.value }))}
              placeholder="Ej: Maquinaria industrial, Vehículo..."
            />
          </div>

          {/* Sucursal */}
          <div className="modal-field" style={{ marginTop: 10 }}>
            <label>Sucursal <span className="required">*</span></label>
            <select
              className="doc-input"
              style={{ borderColor: !activoFormDetalle.sucursal ? "#ef4444" : undefined }}
              value={activoFormDetalle.sucursal || ""}
              onChange={(e) => setActivoFormDetalle(prev => ({ ...prev, sucursal: e.target.value }))}
            >
              <option value="">Seleccione sucursal...</option>
              {sedesDisponibles.map(s => <option key={s} value={s}>{s}</option>)}
            </select>
          </div>

          {/* Proyecto + Descripción */}
          <div style={{ display: "flex", gap: 10, marginTop: 10 }}>
            <div className="modal-field" style={{ flex: 1 }}>
              <label>Proyecto</label>
              <input
                type="text"
                list="activo-proyectos-list"
                className="doc-input"
                value={activoFormDetalle.proyecto || ""}
                onChange={(e) => setActivoFormDetalle(prev => ({ ...prev, proyecto: e.target.value }))}
                placeholder="Seleccione o escriba un proyecto..."
              />
              <datalist id="activo-proyectos-list">
                {(proyectosDisponibles || []).map(p => (
                  <option key={p.id} value={p.nombre} />
                ))}
              </datalist>
            </div>
            <div className="modal-field" style={{ flex: 2 }}>
              <label>
                Descripción{" "}
                <small style={{ float: "right", color: (activoFormDetalle.descripcion || "").length >= 100 ? "#ef4444" : "#6b7280" }}>
                  {(activoFormDetalle.descripcion || "").length}/100
                </small>
              </label>
              <input
                type="text"
                className="doc-input"
                maxLength={100}
                value={activoFormDetalle.descripcion || ""}
                onChange={(e) => setActivoFormDetalle(prev => ({ ...prev, descripcion: e.target.value }))}
                placeholder="Descripción (máx 100 caracteres)..."
              />
            </div>
          </div>

          {/* Porcentaje + Valor */}
          <div style={{ display: "flex", gap: 10, marginTop: 10 }}>
            <div className="modal-field" style={{ flex: 1 }}>
              <label>Porcentaje <span className="required">*</span></label>
              <input
                type="number"
                min="0"
                max="100"
                step="0.01"
                className="doc-input"
                value={activoFormDetalle.porcentaje || ""}
                onChange={handlePctChange}
                placeholder="%"
              />
            </div>
            <div className="modal-field" style={{ flex: 1 }}>
              <label>Valor ($)</label>
              <input
                type="number"
                min="0"
                step="0.01"
                className="doc-input"
                value={activoFormDetalle.valor || ""}
                onChange={handleValChange}
                placeholder="Valor $"
              />
            </div>
          </div>

          <p style={{ fontSize: "0.8em", color: "#64748b", marginTop: 10, padding: "8px 10px", background: "#f0fdf4", borderRadius: 6, border: "1px solid #bbf7d0" }}>
            <i className="fa-solid fa-triangle-exclamation" style={{ color: "#16a34a" }}></i>{" "}
            La suma de <strong>Normas de Reparto + Activos Fijos</strong> debe ser <strong>100%</strong> al guardar.
          </p>
        </div>
        <div className="modal-footer">
          <button className="doc-btn doc-btn-secondary" onClick={() => setShowActivoModal(false)}>
            Cancelar
          </button>
          <button
            className="doc-btn doc-btn-primary"
            onClick={() => handleGuardarActivoFijo(activoModalRadicadoId)}
          >
            <i className="fa-solid fa-floppy-disk"></i>{" "}
            {activoEditandoId ? "Actualizar" : "Agregar"}
          </button>
        </div>
      </div>
    </div>
  );
}
