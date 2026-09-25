import { formatCurrency } from "../../helpers/formatters";

export default function ModalRadicar({
  showRadicarModal, setShowRadicarModal, radicarForm, radicando,
  normasRepartoAutoMsg, tiposRadicacion, rutas, tiposPago = [], metodosPago,
  normasRepartoCatalogo, proyectosCatalogo = [],
  normaFiltroSede, setNormaFiltroSede, normaFiltroArea, setNormaFiltroArea,
  normaSeleccionadaId, setNormaSeleccionadaId, normaPorcentajeInput, setNormaPorcentajeInput,
  normaValorInput, setNormaValorInput, normaProyectoInput, setNormaProyectoInput,
  normaDescripcionInput, setNormaDescripcionInput,
  // Inputs de activos fijos
  activoNombreInput, setActivoNombreInput, activoSucursalInput, setActivoSucursalInput,
  activoProyectoInput, setActivoProyectoInput, activoPorcentajeInput, setActivoPorcentajeInput,
  activoValorInput, setActivoValorInput, activoDescripcionInput, setActivoDescripcionInput,
  subtotalDoc, sedesDisponibles, areasDisponibles, normasFiltradas,
  totalPorcentajeNormas = 0, totalPorcentajeActivos = 0, totalPorcentajeCombinado = 0,
  normasPredeterminadas, usarNormasPredeterminadas,
  aceptarNormasPredeterminadas, rechazarNormasPredeterminadas,
  handleRadicarChange, handleAgregarNormaModal, handleNormaRepartoChange, handleRemoveNormaReparto,
  handleAgregarActivoModal, handleActivoFijoChange, handleRemoveActivoFijo,
  handleRadicarSubmit
}) {
  if (!showRadicarModal) return null;

  const handleNormaPctChange = (e) => {
    const val = e.target.value;
    setNormaPorcentajeInput(val);
    if (subtotalDoc > 0 && val) {
      setNormaValorInput(((parseFloat(val) / 100) * subtotalDoc).toFixed(2));
    } else {
      setNormaValorInput("");
    }
  };

  const handleNormaValChange = (e) => {
    const val = e.target.value;
    setNormaValorInput(val);
    if (subtotalDoc > 0 && val) {
      setNormaPorcentajeInput(((parseFloat(val) / subtotalDoc) * 100).toFixed(2));
    } else {
      setNormaPorcentajeInput("");
    }
  };

  const handleActivoPctChange = (e) => {
    const val = e.target.value;
    setActivoPorcentajeInput(val);
    if (subtotalDoc > 0 && val) {
      setActivoValorInput(((parseFloat(val) / 100) * subtotalDoc).toFixed(2));
    } else {
      setActivoValorInput("");
    }
  };

  const handleActivoValChange = (e) => {
    const val = e.target.value;
    setActivoValorInput(val);
    if (subtotalDoc > 0 && val) {
      setActivoPorcentajeInput(((parseFloat(val) / subtotalDoc) * 100).toFixed(2));
    } else {
      setActivoPorcentajeInput("");
    }
  };

  const estaExacto100 = Math.abs(totalPorcentajeCombinado - 100) < 0.01;
  const excede100 = totalPorcentajeCombinado > 100.01;

  return (
    <div className="modal-overlay" onClick={() => setShowRadicarModal(false)}>
      <div className="modal-content" onClick={(e) => e.stopPropagation()} style={{ maxWidth: "680px", maxHeight: "88vh", display: "flex", flexDirection: "column" }}>
        <div className="modal-header">
          <h3><i className="fa-solid fa-stamp"></i> Radicar Documento</h3>
          <button className="modal-close" onClick={() => setShowRadicarModal(false)}><i className="fa-solid fa-xmark"></i></button>
        </div>

        <div className="modal-body" style={{ overflowY: "auto", flex: 1, paddingRight: "8px" }}>
          
          {/* Datalist global para autocompletar proyectos */}
          <datalist id="proyectos-radicar-list">
            {proyectosCatalogo.map(p => (
              <option key={p.id} value={p.nombre}>
                {p.zona ? `[${p.zona}] ` : ""}{p.nombre}
              </option>
            ))}
          </datalist>

          {/* ── SECCIÓN 1: DATOS GENERALES ── */}
          <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: "12px", marginBottom: "12px" }}>
            <div className="modal-field">
              <label>Tipo de Radicación <span className="required">*</span></label>
              <select name="tipo_radicacion_id" value={radicarForm.tipo_radicacion_id} onChange={handleRadicarChange} className="doc-input">
                <option value="">Seleccione...</option>
                {tiposRadicacion.map(tr => <option key={tr.id} value={tr.id}>{tr.nombre}</option>)}
              </select>
            </div>
            <div className="modal-field">
              <label>Ruta <span className="required">*</span></label>
              <select name="ruta_id" value={radicarForm.ruta_id} onChange={handleRadicarChange} className="doc-input">
                <option value="">Seleccione...</option>
                <optgroup label="BUCARAMANGA">
                  {rutas.filter(r => !r.zona || r.zona === "BUCARAMANGA").map(r => <option key={r.id} value={r.id}>{r.nombre} - {r.area}</option>)}
                </optgroup>
                <optgroup label="MALAMBO">
                  {rutas.filter(r => r.zona === "MALAMBO").map(r => <option key={r.id} value={r.id}>{r.nombre} - {r.area}</option>)}
                </optgroup>
              </select>
            </div>
          </div>

          <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: "12px", marginBottom: "12px" }}>
            <div className="modal-field">
              <label>Tipo de Pago <span className="required">*</span></label>
              <select
                name="tipo_pago_id"
                value={radicarForm.tipo_pago_id || ""}
                onChange={(e) => {
                  handleRadicarChange({ target: { name: "tipo_pago_id", value: e.target.value } });
                  handleRadicarChange({ target: { name: "metodo_pago_id", value: "" } });
                }}
                className="doc-input"
              >
                <option value="">Seleccione...</option>
                {tiposPago.map(tp => <option key={tp.id} value={tp.id}>{tp.nombre}</option>)}
              </select>
            </div>
            <div className="modal-field">
              <label>Método de Pago <span className="required">*</span></label>
              <select
                name="metodo_pago_id"
                value={radicarForm.metodo_pago_id || ""}
                onChange={handleRadicarChange}
                className="doc-input"
                disabled={!radicarForm.tipo_pago_id}
              >
                <option value="">{!radicarForm.tipo_pago_id ? "Primero seleccione un tipo de pago..." : "Seleccione..."}</option>
                {metodosPago
                  .filter(mp => String(mp.tipo_pago_id) === String(radicarForm.tipo_pago_id))
                  .map(mp => <option key={mp.id} value={mp.id}>{mp.nombre}</option>)
                }
              </select>
            </div>
          </div>

          {/* Info Ruta Malambo/Bucaramanga */}
          {(() => {
            const rutaSeleccionada = rutas.find(r => String(r.id) === String(radicarForm.ruta_id));
            const isMalambo = rutaSeleccionada ? rutaSeleccionada.zona === "MALAMBO" : false;
            return (
              <div className="modal-field" style={{ padding: "10px 12px", borderRadius: 8, background: "#f8fafc", border: "1px solid #e2e8f0", marginBottom: "14px" }}>
                <p style={{ margin: 0, fontSize: "0.85em", color: "#475569", fontWeight: 500 }}>
                  <i className="fa-solid fa-circle-info" style={{ color: "#3b82f6", marginRight: 6 }}></i>
                  {isMalambo
                    ? "Se aplicará la configuración de Sede Malambo: el 2º paso de revisión se asignará a auxadmonnorte@harinerapardo.co"
                    : "Se aplicará la configuración de Sede Bucaramanga: el 2º paso de revisión se asignará a analistaadmonoriente@harinerapardo.co"}
                </p>
              </div>
            );
          })()}

          {/* Aviso de normas predeterminadas */}
          {normasPredeterminadas.length > 0 && usarNormasPredeterminadas === null && (
            <div className="modal-field" style={{
              padding: "14px", borderRadius: 10, marginBottom: "14px",
              background: "linear-gradient(135deg, #fef3c7, #fde68a)",
              border: "1px solid #f59e0b", color: "#92400e"
            }}>
              <div style={{ display: "flex", alignItems: "center", gap: 8, marginBottom: 6 }}>
                <i className="fa-solid fa-lightbulb" style={{ fontSize: 18, color: "#f59e0b" }}></i>
                <strong style={{ fontSize: 14 }}>Normas de reparto predeterminadas disponibles</strong>
              </div>
              <p style={{ margin: "0 0 10px", fontSize: 12.5, lineHeight: 1.4 }}>
                Este proveedor tiene <strong>{normasPredeterminadas.length} norma{normasPredeterminadas.length !== 1 ? 's' : ''}</strong> guardada{normasPredeterminadas.length !== 1 ? 's' : ''} para esta ruta:
                {" "}({normasPredeterminadas.map(n => n.norma_reparto?.codigo || `Norma ${n.norma_reparto_id}`).join(", ")}).
              </p>
              <div style={{ display: "flex", gap: 10 }}>
                <button type="button" onClick={aceptarNormasPredeterminadas} style={{
                  padding: "6px 14px", borderRadius: 6, border: "none",
                  background: "#f59e0b", color: "#fff", fontWeight: 600,
                  fontSize: 12, cursor: "pointer", display: "flex", alignItems: "center", gap: 6
                }}>
                  <i className="fa-solid fa-check"></i> Cargar automáticamente
                </button>
                <button type="button" onClick={rechazarNormasPredeterminadas} style={{
                  padding: "6px 14px", borderRadius: 6, border: "1px solid #d97706",
                  background: "#fff", color: "#92400e", fontWeight: 600,
                  fontSize: 12, cursor: "pointer"
                }}>
                  No, dejar en blanco
                </button>
              </div>
            </div>
          )}

          {usarNormasPredeterminadas === true && (
            <div className="modal-field" style={{
              padding: "8px 12px", borderRadius: 8, marginBottom: "14px",
              background: "#d1fae5", border: "1px solid #10b981",
              color: "#065f46", fontSize: 12.5, display: "flex", alignItems: "center", gap: 8
            }}>
              <i className="fa-solid fa-check-circle"></i>
              <span>Normas de reparto cargadas automáticamente.</span>
            </div>
          )}

          {normasRepartoAutoMsg && (
            <div className="modal-field" style={{ padding: "8px 12px", borderRadius: 8, background: "#fef3c7", border: "1px solid #f59e0b", color: "#92400e", fontSize: "0.88em", fontWeight: 600, marginBottom: "14px" }}>{normasRepartoAutoMsg}</div>
          )}

          {/* ── BARRA RESUMEN DE PORCENTAJE COMBINADO ── */}
          <div style={{
            padding: "10px 14px", borderRadius: 8, marginBottom: "16px",
            background: excede100 ? "#fee2e2" : estaExacto100 ? "#d1fae5" : "#f1f5f9",
            border: `1px solid ${excede100 ? "#fca5a5" : estaExacto100 ? "#6ee7b7" : "#cbd5e1"}`,
            display: "flex", justifyContent: "space-between", alignItems: "center"
          }}>
            <div>
              <span style={{ fontSize: "0.85em", fontWeight: 700, color: excede100 ? "#991b1b" : estaExacto100 ? "#065f46" : "#334155" }}>
                <i className={`fa-solid ${excede100 ? "fa-circle-exclamation" : estaExacto100 ? "fa-check-circle" : "fa-chart-pie"}`} style={{ marginRight: 6 }}></i>
                Total Asignado (Normas {totalPorcentajeNormas.toFixed(2)}% + Activos {totalPorcentajeActivos.toFixed(2)}%)
              </span>
            </div>
            <span style={{
              fontSize: "0.95em", fontWeight: 800,
              color: excede100 ? "#dc2626" : estaExacto100 ? "#059669" : "#0284c7"
            }}>
              {totalPorcentajeCombinado.toFixed(2)}%{subtotalDoc > 0 && ` (${formatCurrency((totalPorcentajeCombinado / 100) * subtotalDoc)})`} / 100%
            </span>
          </div>

          {/* ── SECCIÓN 2: NORMAS DE REPARTO ── */}
          <div className="doc-section" style={{ border: "1px solid #e2e8f0", padding: "14px", borderRadius: "10px", marginBottom: "16px", background: "#ffffff" }}>
            <h4 style={{ margin: "0 0 10px 0", fontSize: "0.95em", color: "#1e293b", display: "flex", justifyContent: "space-between", alignItems: "center" }}>
              <span><i className="fa-solid fa-chart-pie" style={{ color: "#3b82f6" }}></i> Normas de Reparto ({(radicarForm.normas_reparto || []).length})</span>
              <span style={{ fontSize: "0.8em", color: "#64748b" }}>Suma: {totalPorcentajeNormas.toFixed(2)}%{subtotalDoc > 0 && ` (${formatCurrency((totalPorcentajeNormas / 100) * subtotalDoc)})`}</span>
            </h4>

            {subtotalDoc > 0 && (
              <p style={{ fontSize: "0.78em", color: "#64748b", margin: "0 0 10px" }}>
                Subtotal base: <strong>{formatCurrency(subtotalDoc)}</strong> (ingresa % o $ para auto-calcular)
              </p>
            )}

            <div style={{ display: "flex", gap: 8, marginBottom: 8 }}>
              <select className="doc-input" style={{ flex: 1, fontSize: "0.82em" }} value={normaFiltroSede} onChange={(e) => { setNormaFiltroSede(e.target.value); setNormaFiltroArea(""); setNormaSeleccionadaId(""); }}>
                <option value="">Todas las sedes</option>{sedesDisponibles.map(s => <option key={s} value={s}>{s}</option>)}
              </select>
              <select className="doc-input" style={{ flex: 1, fontSize: "0.82em" }} value={normaFiltroArea} onChange={(e) => { setNormaFiltroArea(e.target.value); setNormaSeleccionadaId(""); }} disabled={!normaFiltroSede}>
                <option value="">Todas las áreas</option>{areasDisponibles.map(a => <option key={a} value={a}>{a}</option>)}
              </select>
            </div>

            <div style={{ display: "flex", gap: 8, marginBottom: 8, alignItems: "center", flexWrap: "wrap" }}>
              <select className="doc-input" style={{ flex: 2, minWidth: 150, fontSize: "0.82em" }} value={normaSeleccionadaId} onChange={(e) => setNormaSeleccionadaId(e.target.value)} disabled={!normaFiltroSede || !normaFiltroArea}>
                <option value="">Seleccione norma...</option>{normasFiltradas.map(n => <option key={n.id} value={String(n.id)}>{n.codigo} — {n.nombre}</option>)}
              </select>
              <input type="number" min="0" max="100" step="0.01" placeholder="%" className="doc-input" style={{ flex: 0.5, textAlign: "right", minWidth: 65, fontSize: "0.82em" }} value={normaPorcentajeInput} onChange={handleNormaPctChange} />
              <input type="number" min="0" step="0.01" placeholder="Valor $" className="doc-input" style={{ flex: 0.7, textAlign: "right", minWidth: 85, fontSize: "0.82em" }} value={normaValorInput} onChange={handleNormaValChange} />
            </div>

            <div style={{ display: "flex", gap: 8, marginBottom: 12 }}>
              <input
                type="text"
                list="proyectos-radicar-list"
                className="doc-input"
                style={{ flex: 1, fontSize: "0.82em", borderColor: !normaProyectoInput ? "#ef4444" : undefined }}
                placeholder="Proyecto * (selecciona o escribe)"
                value={normaProyectoInput}
                onChange={(e) => setNormaProyectoInput(e.target.value)}
              />
              <input
                type="text"
                className="doc-input"
                style={{ flex: 1.5, fontSize: "0.82em" }}
                maxLength={100}
                placeholder="Descripción (opcional)"
                value={normaDescripcionInput}
                onChange={(e) => setNormaDescripcionInput(e.target.value)}
              />
              <button className="doc-btn doc-btn-secondary" onClick={handleAgregarNormaModal} disabled={!normaSeleccionadaId || (!normaPorcentajeInput && !normaValorInput)} style={{ padding: "6px 14px", fontSize: "0.82em", flexShrink: 0 }}>
                <i className="fa-solid fa-plus"></i> Agregar
              </button>
            </div>

            {/* Lista de Normas Agregadas */}
            {(radicarForm.normas_reparto || []).length === 0 ? (
              <p style={{ fontSize: "0.82em", color: "#94a3b8", margin: 0, fontStyle: "italic" }}>No se han asignado normas de reparto.</p>
            ) : (
              <div style={{ display: "flex", flexDirection: "column", gap: 6 }}>
                {radicarForm.normas_reparto.map((norma, idx) => {
                  const info = normasRepartoCatalogo.find(n => String(n.id) === String(norma.norma_reparto_id));
                  const pctVal = parseFloat(norma.porcentaje) || 0;
                  const calculatedVal = subtotalDoc > 0 ? (pctVal / 100) * subtotalDoc : parseFloat(norma.valor) || 0;
                  return (
                    <div key={idx} style={{ display: "flex", flexDirection: "column", gap: 4, padding: "8px 10px", background: "#f8fafc", borderRadius: 6, border: "1px solid #e2e8f0" }}>
                      <div style={{ display: "flex", gap: 8, alignItems: "center" }}>
                        <div style={{ flex: 2, display: "flex", flexDirection: "column" }}>
                          <span style={{ fontSize: "0.82em", fontWeight: 600, color: "#1e293b" }}>{info ? `${info.codigo} — ${info.nombre}` : "Norma"}</span>
                          <span style={{ fontSize: "0.72em", color: "#64748b" }}>{info ? `${info.sucursal} / ${info.departamento}` : ""}</span>
                        </div>
                        <input type="number" min="0" max="100" step="0.01" value={norma.porcentaje} onChange={(e) => handleNormaRepartoChange(idx, "porcentaje", e.target.value)} className="doc-input" style={{ width: 65, textAlign: "right", fontSize: "0.82em" }} />
                        <span style={{ fontSize: "0.82em", fontWeight: 700 }}>%</span>
                        <span style={{ fontSize: "0.8e", color: "#059669", fontWeight: 600, minWidth: 85, textAlign: "right" }}>{formatCurrency(calculatedVal)}</span>
                        <button className="btn-icon btn-toggle" onClick={() => handleRemoveNormaReparto(idx)} title="Quitar" style={{ width: 26, height: 26, flexShrink: 0 }}><i className="fa-solid fa-xmark"></i></button>
                      </div>
                      <div style={{ display: "flex", gap: 8 }}>
                        <input type="text" list="proyectos-radicar-list" className="doc-input" placeholder="Proyecto" style={{ flex: 1, fontSize: "0.78em", padding: "3px 6px" }} value={norma.proyecto || ""} onChange={(e) => handleNormaRepartoChange(idx, "proyecto", e.target.value)} />
                        <input type="text" className="doc-input" placeholder="Descripción" maxLength={100} style={{ flex: 1.5, fontSize: "0.78em", padding: "3px 6px" }} value={norma.descripcion || ""} onChange={(e) => handleNormaRepartoChange(idx, "descripcion", e.target.value)} />
                      </div>
                    </div>
                  );
                })}
              </div>
            )}
          </div>

          {/* ── SECCIÓN 3: ACTIVOS FIJOS ── */}
          <div className="doc-section" style={{ border: "1px solid #e2e8f0", padding: "14px", borderRadius: "10px", background: "#ffffff" }}>
            <h4 style={{ margin: "0 0 10px 0", fontSize: "0.95em", color: "#1e293b", display: "flex", justifyContent: "space-between", alignItems: "center" }}>
              <span><i className="fa-solid fa-boxes-stacked" style={{ color: "#8b5cf6" }}></i> Activos Fijos ({(radicarForm.activos_fijos || []).length})</span>
              <span style={{ fontSize: "0.8em", color: "#64748b" }}>Suma: {totalPorcentajeActivos.toFixed(2)}%{subtotalDoc > 0 && ` (${formatCurrency((totalPorcentajeActivos / 100) * subtotalDoc)})`}</span>
            </h4>

            <div style={{ display: "flex", gap: 8, marginBottom: 8, flexWrap: "wrap" }}>
              <input
                type="text"
                className="doc-input"
                style={{ flex: 1.5, minWidth: 140, fontSize: "0.82em" }}
                placeholder="Nombre del Activo Fijo *"
                value={activoNombreInput}
                onChange={(e) => setActivoNombreInput(e.target.value)}
              />
              <select
                className="doc-input"
                style={{ flex: 1, minWidth: 120, fontSize: "0.82em" }}
                value={activoSucursalInput}
                onChange={(e) => setActivoSucursalInput(e.target.value)}
              >
                <option value="">Sucursal *</option>
                {sedesDisponibles.map(s => <option key={s} value={s}>{s}</option>)}
              </select>
              <input type="number" min="0" max="100" step="0.01" placeholder="%" className="doc-input" style={{ flex: 0.5, textAlign: "right", minWidth: 65, fontSize: "0.82em" }} value={activoPorcentajeInput} onChange={handleActivoPctChange} />
              <input type="number" min="0" step="0.01" placeholder="Valor $" className="doc-input" style={{ flex: 0.7, textAlign: "right", minWidth: 85, fontSize: "0.82em" }} value={activoValorInput} onChange={handleActivoValChange} />
            </div>

            <div style={{ display: "flex", gap: 8, marginBottom: 12 }}>
              <input
                type="text"
                list="proyectos-radicar-list"
                className="doc-input"
                style={{ flex: 1, fontSize: "0.82em" }}
                placeholder="Proyecto (opcional)"
                value={activoProyectoInput}
                onChange={(e) => setActivoProyectoInput(e.target.value)}
              />
              <input
                type="text"
                className="doc-input"
                style={{ flex: 1.5, fontSize: "0.82em" }}
                maxLength={100}
                placeholder="Descripción (opcional)"
                value={activoDescripcionInput}
                onChange={(e) => setActivoDescripcionInput(e.target.value)}
              />
              <button className="doc-btn doc-btn-secondary" onClick={handleAgregarActivoModal} disabled={!activoNombreInput || !activoSucursalInput || (!activoPorcentajeInput && !activoValorInput)} style={{ padding: "6px 14px", fontSize: "0.82em", flexShrink: 0 }}>
                <i className="fa-solid fa-plus"></i> Agregar
              </button>
            </div>

            {/* Lista de Activos Agregados */}
            {(radicarForm.activos_fijos || []).length === 0 ? (
              <p style={{ fontSize: "0.82em", color: "#94a3b8", margin: 0, fontStyle: "italic" }}>No se han asignado activos fijos.</p>
            ) : (
              <div style={{ display: "flex", flexDirection: "column", gap: 6 }}>
                {radicarForm.activos_fijos.map((activo, idx) => {
                  const pctVal = parseFloat(activo.porcentaje) || 0;
                  const calculatedVal = subtotalDoc > 0 ? (pctVal / 100) * subtotalDoc : parseFloat(activo.valor) || 0;
                  return (
                    <div key={idx} style={{ display: "flex", flexDirection: "column", gap: 4, padding: "8px 10px", background: "#f8fafc", borderRadius: 6, border: "1px solid #e2e8f0" }}>
                      <div style={{ display: "flex", gap: 8, alignItems: "center" }}>
                        <div style={{ flex: 2, display: "flex", flexDirection: "column" }}>
                          <span style={{ fontSize: "0.82em", fontWeight: 600, color: "#1e293b" }}>{activo.nombre}</span>
                          <span style={{ fontSize: "0.72em", color: "#64748b" }}>{activo.sucursal}</span>
                        </div>
                        <input type="number" min="0" max="100" step="0.01" value={activo.porcentaje} onChange={(e) => handleActivoFijoChange(idx, "porcentaje", e.target.value)} className="doc-input" style={{ width: 65, textAlign: "right", fontSize: "0.82em" }} />
                        <span style={{ fontSize: "0.82em", fontWeight: 700 }}>%</span>
                        <span style={{ fontSize: "0.8em", color: "#059669", fontWeight: 600, minWidth: 85, textAlign: "right" }}>{formatCurrency(calculatedVal)}</span>
                        <button className="btn-icon btn-toggle" onClick={() => handleRemoveActivoFijo(idx)} title="Quitar" style={{ width: 26, height: 26, flexShrink: 0 }}><i className="fa-solid fa-xmark"></i></button>
                      </div>
                      <div style={{ display: "flex", gap: 8 }}>
                        <input type="text" list="proyectos-radicar-list" className="doc-input" placeholder="Proyecto" style={{ flex: 1, fontSize: "0.78em", padding: "3px 6px" }} value={activo.proyecto || ""} onChange={(e) => handleActivoFijoChange(idx, "proyecto", e.target.value)} />
                        <input type="text" className="doc-input" placeholder="Descripción" maxLength={100} style={{ flex: 1.5, fontSize: "0.78em", padding: "3px 6px" }} value={activo.descripcion || ""} onChange={(e) => handleActivoFijoChange(idx, "descripcion", e.target.value)} />
                      </div>
                    </div>
                  );
                })}
              </div>
            )}
          </div>

        </div>

        <div className="modal-footer">
          <button className="doc-btn doc-btn-secondary" onClick={() => setShowRadicarModal(false)} disabled={radicando}>Cancelar</button>
          <button className="doc-btn doc-btn-primary" onClick={handleRadicarSubmit} disabled={radicando || excede100}>
            <i className="fa-solid fa-stamp"></i> {radicando ? "Radicando..." : "Radicar Documento"}
          </button>
        </div>
      </div>
    </div>
  );
}