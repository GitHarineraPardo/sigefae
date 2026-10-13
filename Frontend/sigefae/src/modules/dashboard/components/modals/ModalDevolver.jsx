export default function ModalDevolver({
  showDevolverModal, setShowDevolverModal, devolverForm, setDevolverForm,
  devolviendo, tareasFlujo, handleDevolverTarea, selectedRadicadoId, selectedTareaId
}) {
  if (!showDevolverModal) return null;

  // Determinar la tarea "activa" — puede ser la En Proceso (caso normal)
  // o la última Completada (caso Contabilidad devolviendo desde finanzas)
  const tareaEnProceso = tareasFlujo.find(t => t.estado?.nombre === "En Proceso");
  const tareasCompletadas = tareasFlujo.filter(t => t.estado?.nombre === "Completada");
  const ultimaCompletada = tareasCompletadas.length > 0 ? tareasCompletadas[tareasCompletadas.length - 1] : null;
  const tareaActiva = tareaEnProceso || ultimaCompletada;

  // Tareas destino:
  // - Si hay tarea En Proceso: mostrar completadas anteriores (flujo normal)
  // - Si NO hay tarea En Proceso (Contabilidad): mostrar TODAS las completadas excepto la última (que se usa como fuente)
  const tareasDestino = tareasFlujo.filter(t => {
    if (t.estado?.nombre !== "Completada") return false;
    if (tareaEnProceso) {
      // Flujo normal: solo las anteriores a la tarea en proceso
      return t.id < tareaEnProceso.id;
    } else {
      // Contabilidad: todas las completadas excepto la última (que es la "activa")
      return t.id !== ultimaCompletada?.id;
    }
  });

  const getAsignadoLabel = (tarea) => {
    if (tarea.usuario_asignado?.nombre) {
      return {
        nombre: tarea.usuario_asignado.nombre,
        email: tarea.usuario_asignado.email || ""
      };
    }
    if (tarea.cargo_asignado?.nombre) {
      return {
        nombre: `Cargo: ${tarea.cargo_asignado.nombre}`,
        email: ""
      };
    }
    if (typeof tarea.cargo_asignado === "string" && tarea.cargo_asignado) {
      return {
        nombre: `Cargo: ${tarea.cargo_asignado}`,
        email: ""
      };
    }
    return { nombre: "Sin usuario", email: "" };
  };

  return (
    <div className="modal-overlay" onClick={() => setShowDevolverModal(false)}>
      <div className="modal-content" onClick={(e) => e.stopPropagation()}>
        <div className="modal-header">
          <h3><i className="fa-solid fa-reply"></i> Devolver Tarea</h3>
          <button className="modal-close" onClick={() => setShowDevolverModal(false)}><i className="fa-solid fa-xmark"></i></button>
        </div>
        <div className="modal-body">
          <div className="modal-field">
            <label>Devolver a Paso / Usuario <span className="required">*</span></label>
            <select value={devolverForm.tarea_destino_id} onChange={(e) => setDevolverForm(prev => ({ ...prev, tarea_destino_id: e.target.value }))} className="doc-input">
              <option value="">Seleccione paso de destino...</option>
              {tareasDestino.map(t => {
                const asignado = getAsignadoLabel(t);
                return (
                  <option key={t.id} value={t.id}>
                    {t.descripcion} — {asignado.nombre}{asignado.email ? ` (${asignado.email})` : ""} [Completada]
                  </option>
                );
              })}
            </select>
          </div>
          <div className="modal-field">
            <label>Motivo de la Devolución <span className="required">*</span></label>
            <textarea value={devolverForm.observacion} onChange={(e) => setDevolverForm(prev => ({ ...prev, observacion: e.target.value }))} className="doc-input" rows={3} placeholder="Describa claramente por qué devuelve esta tarea y qué debe ser corregido..." style={{ width: "100%", fontFamily: "inherit", resize: "vertical" }} />
          </div>
          <div className="modal-field" style={{ display: "flex", alignItems: "center", gap: 8, marginTop: 12 }}>
            <input type="checkbox" id="retorno_directo" checked={devolverForm.retorno_directo} onChange={(e) => setDevolverForm(prev => ({ ...prev, retorno_directo: e.target.checked }))} />
            <label htmlFor="retorno_directo" style={{ fontSize: "0.9em", color: "#374151", cursor: "pointer", margin: 0 }}>Regresar directamente a mi paso una vez corregido (Smart Return)</label>
          </div>
        </div>
        <div className="modal-footer">
          <button className="doc-btn doc-btn-secondary" onClick={() => setShowDevolverModal(false)} disabled={devolviendo}>Cancelar</button>
          <button className="doc-btn doc-btn-primary" onClick={() => {
            const radId = selectedRadicadoId || selectedTareaId;
            if (tareaActiva && radId) handleDevolverTarea(tareaActiva.id, radId);
          }} disabled={devolviendo} style={{ background: "var(--pardo-red)", borderColor: "var(--pardo-red)" }}>
            <i className="fa-solid fa-reply"></i> {devolviendo ? "Devolviendo..." : "Confirmar Devolución"}
          </button>
        </div>
      </div>
    </div>
  );
}