import { formatCurrency } from "../helpers/formatters";

export default function RenderNormasReparto({
  normasRepartoRadicado, radicadoId, readOnly = false,
  esAdmin, puedeGestionarRecurso, openNormaModal, handleEliminarNorma,
  activosFijosRadicado, openActivoModal, handleEliminarActivoFijo,
  subtotalRadicado = 0
}) {
  const normas = normasRepartoRadicado || [];
  const activos = activosFijosRadicado || [];
  const subtotal = parseFloat(subtotalRadicado) || 0;

  const totalNormasPct = normas.reduce((s, n) => s + (parseFloat(n.porcentaje) || 0), 0);
  const totalActivosPct = activos.reduce((s, a) => s + (parseFloat(a.porcentaje) || 0), 0);
  const totalPct = totalNormasPct + totalActivosPct;
  const excede100 = totalPct > 100;

  return (
    <div>
      {/* ── Normas de Reparto ── */}
      <div className="doc-section">
        <h4 style={{ margin: 0, display: 'flex', alignItems: 'center', gap: 8, justifyContent: 'space-between' }}>
          <span style={{ display: 'inline-flex', alignItems: 'center', gap: 8 }}>
            <i className="fa-solid fa-chart-pie"></i> Normas de Reparto ({normas.length})
          </span>
          {!readOnly && (
            <button className="doc-btn doc-btn-secondary" onClick={() => openNormaModal(radicadoId)} style={{ padding: "6px 12px", fontSize: "0.8em" }}>
              <i className="fa-solid fa-plus"></i> Agregar Norma
            </button>
          )}
        </h4>

        {normas.length === 0 ? (
          <p style={{ color: "#6b7280", fontSize: "0.9em" }}>No hay normas de reparto asignadas.</p>
        ) : (
          <table className="doc-items-table">
            <thead>
              <tr>
                <th>Código</th>
                <th>Nombre</th>
                <th>Proyecto</th>
                <th>Descripción</th>
                <th>Sede</th>
                <th>Área</th>
                <th style={{ textAlign: "right" }}>%</th>
                {subtotal > 0 && <th style={{ textAlign: "right" }}>Valor ($)</th>}
                {!readOnly && <th style={{ width: 90, textAlign: "center" }}>Acciones</th>}
              </tr>
            </thead>
            <tbody>
              {normas.map((n) => {
                const pct = parseFloat(n.porcentaje) || 0;
                const valorCalculado = subtotal > 0 ? (pct / 100) * subtotal : 0;
                return (
                  <tr key={n.id}>
                    <td><strong>{n.norma_reparto?.codigo || n.codigo}</strong></td>
                    <td>{n.norma_reparto?.nombre || n.nombre}</td>
                    <td style={{ color: n.proyecto ? "#1e293b" : "#94a3b8", fontStyle: n.proyecto ? "normal" : "italic" }}>
                      {n.proyecto || "—"}
                    </td>
                    <td style={{ color: n.descripcion ? "#1e293b" : "#94a3b8", fontStyle: n.descripcion ? "normal" : "italic", maxWidth: 200, whiteSpace: "nowrap", overflow: "hidden", textOverflow: "ellipsis" }}>
                      {n.descripcion || "—"}
                    </td>
                    <td>{n.norma_reparto?.sucursal || n.sucursal}</td>
                    <td>{n.norma_reparto?.departamento || n.departamento}</td>
                    <td style={{ textAlign: "right", fontWeight: 700 }}>{pct.toFixed(2)}%</td>
                    {subtotal > 0 && (
                      <td style={{ textAlign: "right", fontWeight: 600, color: "#059669" }}>
                        {formatCurrency(valorCalculado)}
                      </td>
                    )}
                    {!readOnly && (
                      <td style={{ textAlign: "center" }}>
                        {esAdmin || puedeGestionarRecurso(n.creado_por_id || n.creado_por?.id) ? (
                          <div style={{ display: "flex", gap: 6, justifyContent: "center" }}>
                            <button className="btn-icon btn-edit" onClick={() => openNormaModal(radicadoId, n)} title="Editar">
                              <i className="fa-solid fa-pen"></i>
                            </button>
                            <button className="btn-icon btn-toggle" onClick={() => handleEliminarNorma(n, radicadoId)} title="Eliminar">
                              <i className="fa-solid fa-xmark"></i>
                            </button>
                          </div>
                        ) : null}
                      </td>
                    )}
                  </tr>
                );
              })}
            </tbody>
          </table>
        )}
      </div>

      {/* ── Activos Fijos ── */}
      <div className="doc-section" style={{ marginTop: 12 }}>
        <h4 style={{ margin: 0, display: 'flex', alignItems: 'center', gap: 8, justifyContent: 'space-between' }}>
          <span style={{ display: 'inline-flex', alignItems: 'center', gap: 8 }}>
            <i className="fa-solid fa-boxes-stacked"></i> Activos Fijos ({activos.length})
          </span>
          {!readOnly && openActivoModal && (
            <button className="doc-btn doc-btn-secondary" onClick={() => openActivoModal(radicadoId)} style={{ padding: "6px 12px", fontSize: "0.8em" }}>
              <i className="fa-solid fa-plus"></i> Agregar Activo Fijo
            </button>
          )}
        </h4>

        {activos.length === 0 ? (
          <p style={{ color: "#6b7280", fontSize: "0.9em" }}>No hay activos fijos asignados.</p>
        ) : (
          <table className="doc-items-table">
            <thead>
              <tr>
                <th>Nombre</th>
                <th>Sucursal</th>
                <th>Proyecto</th>
                <th>Descripción</th>
                <th style={{ textAlign: "right" }}>%</th>
                {subtotal > 0 && <th style={{ textAlign: "right" }}>Valor ($)</th>}
                {!readOnly && <th style={{ width: 90, textAlign: "center" }}>Acciones</th>}
              </tr>
            </thead>
            <tbody>
              {activos.map((a) => {
                const pct = parseFloat(a.porcentaje) || 0;
                const valorCalculado = subtotal > 0 ? (pct / 100) * subtotal : 0;
                return (
                  <tr key={a.id}>
                    <td><strong>{a.nombre}</strong></td>
                    <td>{a.sucursal}</td>
                    <td style={{ color: a.proyecto ? "#1e293b" : "#94a3b8", fontStyle: a.proyecto ? "normal" : "italic" }}>
                      {a.proyecto || "—"}
                    </td>
                    <td style={{ color: a.descripcion ? "#1e293b" : "#94a3b8", fontStyle: a.descripcion ? "normal" : "italic", maxWidth: 200, whiteSpace: "nowrap", overflow: "hidden", textOverflow: "ellipsis" }}>
                      {a.descripcion || "—"}
                    </td>
                    <td style={{ textAlign: "right", fontWeight: 700 }}>{pct.toFixed(2)}%</td>
                    {subtotal > 0 && (
                      <td style={{ textAlign: "right", fontWeight: 600, color: "#059669" }}>
                        {formatCurrency(valorCalculado)}
                      </td>
                    )}
                    {!readOnly && (
                      <td style={{ textAlign: "center" }}>
                        {esAdmin || puedeGestionarRecurso?.(a.creado_por_id || a.creado_por?.id) ? (
                          <div style={{ display: "flex", gap: 6, justifyContent: "center" }}>
                            <button className="btn-icon btn-edit" onClick={() => openActivoModal(radicadoId, a)} title="Editar">
                              <i className="fa-solid fa-pen"></i>
                            </button>
                            <button className="btn-icon btn-toggle" onClick={() => handleEliminarActivoFijo?.(a, radicadoId)} title="Eliminar">
                              <i className="fa-solid fa-xmark"></i>
                            </button>
                          </div>
                        ) : null}
                      </td>
                    )}
                  </tr>
                );
              })}
            </tbody>
          </table>
        )}
      </div>

      {/* ── Total combinado ── */}
      {(normas.length > 0 || activos.length > 0) && (
        <div style={{
          marginTop: 10,
          padding: "10px 14px",
          background: excede100 ? "#fee2e2" : "#f1f5f9",
          borderRadius: 8,
          display: "flex",
          justifyContent: "space-between",
          alignItems: "center",
          fontSize: "0.9em",
          fontWeight: 700,
          color: excede100 ? "#991b1b" : "#334155",
          border: `1px solid ${excede100 ? "#fca5a5" : "#cbd5e1"}`
        }}>
          <span>
            <i className={`fa-solid ${excede100 ? "fa-triangle-exclamation" : "fa-chart-pie"}`} style={{ marginRight: 6 }}></i>
            Total (Normas {totalNormasPct.toFixed(2)}% + Activos {totalActivosPct.toFixed(2)}%)
            {subtotal > 0 && ` — ${formatCurrency((totalPct / 100) * subtotal)} de ${formatCurrency(subtotal)}`}
          </span>
          <span>{totalPct.toFixed(2)} % {excede100 ? `(Supera el máximo por ${(totalPct - 100).toFixed(2)} %)` : `(Disponible: ${(100 - totalPct).toFixed(2)} %)`}</span>
        </div>
      )}
    </div>
  );
}
