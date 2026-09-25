import { formatCurrency } from "../../helpers/formatters";

export default function ModalNormaReparto({
  showNormaModal, setShowNormaModal, normaEditandoId, normaFormDetalle, setNormaFormDetalle,
  normaModalRadicadoId, normaFiltroSede, setNormaFiltroSede, normaFiltroArea, setNormaFiltroArea,
  normasRepartoCatalogo, proyectosDisponibles, sedesDisponibles, areasDisponibles, handleGuardarNormaDetalle,
  subtotalRadicado
}) {
  if (!showNormaModal) return null;

  const subtotal = parseFloat(subtotalRadicado) || 0;

  const normasFiltradas = normasRepartoCatalogo.filter(n => {
    if (normaFiltroSede && n.sucursal !== normaFiltroSede) return false;
    if (normaFiltroArea && n.departamento !== normaFiltroArea) return false;
    return true;
  });

  const handlePctChange = (e) => {
    const val = e.target.value;
    const newForm = { ...normaFormDetalle, porcentaje: val };
    if (subtotal > 0 && val) {
      newForm.valor = ((parseFloat(val) / 100) * subtotal).toFixed(2);
    } else {
      newForm.valor = "";
    }
    setNormaFormDetalle(newForm);
  };

  const handleValChange = (e) => {
    const val = e.target.value;
    const newForm = { ...normaFormDetalle, valor: val };
    if (subtotal > 0 && val) {
      newForm.porcentaje = ((parseFloat(val) / subtotal) * 100).toFixed(2);
    } else {
      newForm.porcentaje = "";
    }
    setNormaFormDetalle(newForm);
  };

  const normaSeleccionadaInfo = normasRepartoCatalogo.find(n => String(n.id) === String(normaFormDetalle.norma_reparto_id));

  return (
    <div className="modal-overlay" onClick={() => setShowNormaModal(false)}>
      <div className="modal-content" onClick={(e) => e.stopPropagation()}>
        <div className="modal-header">
          <h3><i className="fa-solid fa-chart-pie"></i> {normaEditandoId ? "Editar" : "Agregar"} Norma de Reparto</h3>
          <button className="modal-close" onClick={() => setShowNormaModal(false)}><i className="fa-solid fa-xmark"></i></button>
        </div>
        <div className="modal-body">
          {subtotal > 0 && (
            <p style={{ fontSize: "0.85em", color: "#64748b", margin: "0 0 12px" }}>
              <i className="fa-solid fa-info-circle"></i> Subtotal: <strong>{formatCurrency(subtotal)}</strong> — Ingresa % o valor, el otro se calcula automáticamente.
            </p>
          )}
          <div style={{ display: "flex", gap: 8, marginBottom: 10 }}>
            <select className="doc-input" style={{ flex: 1 }} value={normaFiltroSede} onChange={(e) => { setNormaFiltroSede(e.target.value); setNormaFiltroArea(""); }}>
              <option value="">Todas las sedes</option>{sedesDisponibles.map(s => <option key={s} value={s}>{s}</option>)}
            </select>
            <select className="doc-input" style={{ flex: 1 }} value={normaFiltroArea} onChange={(e) => setNormaFiltroArea(e.target.value)}>
              <option value="">Todas las áreas</option>{areasDisponibles.map(a => <option key={a} value={a}>{a}</option>)}
            </select>
          </div>
          <div className="modal-field">
            <label>Norma <span className="required">*</span></label>
            <select className="doc-input" value={normaFormDetalle.norma_reparto_id} onChange={(e) => setNormaFormDetalle(prev => ({ ...prev, norma_reparto_id: e.target.value }))}>
              <option value="">Seleccione norma...</option>
              {normasFiltradas.map(n => <option key={n.id} value={String(n.id)}>{n.codigo} — {n.nombre} {n.proyecto ? `[${n.proyecto}]` : ''} ({n.sucursal} / {n.departamento})</option>)}
            </select>
          </div>

          <div style={{ display: "flex", gap: 10, marginTop: 10 }}>
            <div className="modal-field" style={{ flex: 1 }}>
              <label>Proyecto <span className="required">*</span></label>
              <input
                type="text"
                list="proyectos-list"
                className="doc-input"
                style={{ borderColor: !normaFormDetalle.proyecto?.trim() ? "#ef4444" : undefined }}
                value={normaFormDetalle.proyecto || ""}
                onChange={(e) => setNormaFormDetalle(prev => ({ ...prev, proyecto: e.target.value }))}
                placeholder="Nombre del proyecto (requerido)..."
              />
              <datalist id="proyectos-list">
                {proyectosDisponibles.map(p => (
                  <option key={p.id} value={p.nombre} />
                ))}
              </datalist>
            </div>
            <div className="modal-field" style={{ flex: 2 }}>
              <label>Descripción <small style={{ float: "right", color: (normaFormDetalle.descripcion || "").length >= 100 ? "#ef4444" : "#6b7280" }}>{(normaFormDetalle.descripcion || "").length}/100</small></label>
              <input
                type="text"
                className="doc-input"
                maxLength={100}
                value={normaFormDetalle.descripcion || ""}
                onChange={(e) => setNormaFormDetalle(prev => ({ ...prev, descripcion: e.target.value }))}
                placeholder="Descripción (máx 100 caracteres)..."
              />
            </div>
          </div>
          <div style={{ display: "flex", gap: 10, marginTop: 10 }}>
            <div className="modal-field" style={{ flex: 1 }}>
              <label>Porcentaje <span className="required">*</span></label>
              <input type="number" min="0" max="100" step="0.01" className="doc-input" value={normaFormDetalle.porcentaje} onChange={handlePctChange} placeholder="%" />
            </div>
            <div className="modal-field" style={{ flex: 1 }}>
              <label>Valor ($)</label>
              <input type="number" min="0" step="0.01" className="doc-input" value={normaFormDetalle.valor || ""} onChange={handleValChange} placeholder="Valor $" />
            </div>
          </div>
        </div>
        <div className="modal-footer">
          <button className="doc-btn doc-btn-secondary" onClick={() => setShowNormaModal(false)}>Cancelar</button>
          <button className="doc-btn doc-btn-primary" onClick={() => handleGuardarNormaDetalle(normaModalRadicadoId)}>
            <i className="fa-solid fa-floppy-disk"></i> {normaEditandoId ? "Actualizar" : "Agregar"}
          </button>
        </div>
      </div>
    </div>
  );
}