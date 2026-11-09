import { useState } from "react";
import { formatCurrency } from "../helpers/formatters";
import ModalSubirComprobante from "./modals/ModalSubirComprobante";

export default function RenderFinanzas({
  radicados,
  loadingRadicados,
  searchFinanzas,
  setSearchFinanzas,
  sortFinanzas,
  setSortFinanzas,
  openSaia,
  userRole,
  handleCausar,
  handlePagar,
  obtenerToken,
  handleComprobantesSubidosLocal
}) {
  const [activeSubTab, setActiveSubTab] = useState("en_proceso");
  const [fechaInicio, setFechaInicio] = useState("");
  const [fechaFin, setFechaFin] = useState("");
  const [modalComprobante, setModalComprobante] = useState({ open: false, radicadoId: null });

  // Filtrar radicados por búsqueda y estado
  const getFilteredRadicados = () => {
    // Solo mostramos facturas que ya llegaron a la fase contable (Completado el BPM)
    let list = radicados.filter(r => r.estado_posesion === "Completado");

    if (searchFinanzas.trim()) {
      const lower = searchFinanzas.toLowerCase();
      list = list.filter((r) => {
        const prov = r.documento_comercial?.proveedor?.razon_social?.toLowerCase() || "";
        const num = r.documento_comercial?.numero_documento?.toLowerCase() || "";
        const rad = r.numero_radicado?.toString().toLowerCase() || "";
        return prov.includes(lower) || num.includes(lower) || rad.includes(lower);
      });
    }

    // Ordenamiento inicial por fecha facturación
    if (sortFinanzas === "fecha_fact_desc") {
      list = [...list].sort((a, b) => new Date(b.documento_comercial?.fecha_documento || 0) - new Date(a.documento_comercial?.fecha_documento || 0));
    }

    // Filtro por fechas (usando string comparison seguro YYYY-MM-DD)
    if (fechaInicio) {
      list = list.filter(r => {
        const fechaStr = (r.documento_comercial?.fecha_documento || r.fecha_radicacion || "").substring(0, 10);
        return fechaStr >= fechaInicio;
      });
    }
    if (fechaFin) {
      list = list.filter(r => {
        const fechaStr = (r.documento_comercial?.fecha_documento || r.fecha_radicacion || "").substring(0, 10);
        return fechaStr <= fechaFin;
      });
    }

    // Filtro por pestañas "En Proceso" vs "Finalizados"
    list = list.filter(r => {
      // Requisitos para estar Finalizado en Finanzas:
      const cumpleTodos = r.causado === true && 
                          r.numero_egreso && r.numero_egreso.trim() !== "" && 
                          r.pagado === true && 
                          r.comprobantes_subidos === true;
      
      if (activeSubTab === "finalizados") return cumpleTodos;
      return !cumpleTodos; // "En Proceso"
    });

    if (sortFinanzas === "fecha_fact_asc") {
      list.sort((a, b) => new Date(a.documento_comercial?.fecha_documento || 0) - new Date(b.documento_comercial?.fecha_documento || 0));
    } else if (sortFinanzas === "total_desc") {
      list.sort((a, b) => (b.documento_comercial?.total || 0) - (a.documento_comercial?.total || 0));
    } else if (sortFinanzas === "total_asc") {
      list.sort((a, b) => (a.documento_comercial?.total || 0) - (b.documento_comercial?.total || 0));
    } else if (sortFinanzas === "proveedor") {
      list = [...list].sort((a, b) => {
        const pA = a.documento_comercial?.proveedor?.razon_social || "";
        const pB = b.documento_comercial?.proveedor?.razon_social || "";
        return pA.localeCompare(pB);
      });
    }

    return list;
  };

  const filteredList = getFilteredRadicados();

  const handleCausarChange = (rad) => {
    if (rad.causado && userRole !== "Superadministrador") return; // Solo admin puede desmarcar
    
    const confirmar = window.confirm(
      rad.causado 
        ? "¿Estás seguro que deseas revertir la causación de esta factura?" 
        : "¿Confirmas que esta factura ha sido causada?"
    );
    if (!confirmar) return;

    handleCausar(rad.id, !rad.causado, rad.numero_egreso || "");
  };

  const handleEgresoBlur = (e, rad) => {
    const val = e.target.value.trim();
    if (val === (rad.numero_egreso || "")) return;

    const mensaje = rad.numero_egreso 
      ? `¿Estás seguro que deseas cambiar el N° de Documento de '${rad.numero_egreso}' a '${val}'?`
      : `¿Confirmas que el N° de Documento es '${val}'?`;

    if (!window.confirm(mensaje)) {
      e.target.value = rad.numero_egreso || ""; // revert
      return;
    }

    handleCausar(rad.id, rad.causado, val);
  };

  const handlePagarChange = (rad) => {
    if (rad.pagado && userRole !== "Superadministrador") return; 
    
    const confirmar = window.confirm(
      rad.pagado 
        ? "¿Estás seguro que deseas revertir el pago de esta factura?" 
        : "¿Confirmas que esta factura ha sido pagada?"
    );
    if (!confirmar) return;

    handlePagar(rad.id, !rad.pagado);
  };

  const handleVerComprobante = (rad) => {
    openSaia(rad, "finanzas");
    // Saia should ideally open directly on anexos tab, but just opening it is fine since anexos is prominent.
  };

  return (
    <div className="fullwidth-list-container">
      <div className="fullwidth-list-header" style={{ flexDirection: "column", alignItems: "stretch", gap: "15px" }}>
        <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center" }}>
          <h3><i className="fa-solid fa-file-invoice-dollar"></i> Control Financiero ({filteredList.length})</h3>
          <div className="list-controls">
            <input type="text" placeholder="Buscar radicado, doc, proveedor..." value={searchFinanzas} onChange={e => setSearchFinanzas(e.target.value)} />
            <select value={sortFinanzas} onChange={e => setSortFinanzas(e.target.value)}>
              <option value="fecha_fact_desc">Fech. Fact. (Recientes)</option>
              <option value="fecha_fact_asc">Fech. Fact. (Antiguos)</option>
              <option value="total_desc">Mayor Total</option>
              <option value="total_asc">Menor Total</option>
              <option value="proveedor">Por Proveedor (A-Z)</option>
            </select>
          </div>
        </div>

        <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", borderBottom: "1px solid #e2e8f0", paddingBottom: "10px", flexWrap: "wrap", gap: "10px", overflow: "visible" }}>
          <div style={{ display: "flex", gap: "10px" }}>
            <button 
              className={`doc-btn ${activeSubTab === "en_proceso" ? "doc-btn-primary" : "doc-btn-secondary"}`} 
              onClick={() => setActiveSubTab("en_proceso")}
            >
              En Proceso
            </button>
            <button 
              className={`doc-btn ${activeSubTab === "finalizados" ? "doc-btn-primary" : "doc-btn-secondary"}`} 
              onClick={() => setActiveSubTab("finalizados")}
            >
              Finalizados
            </button>
          </div>
          
          <div style={{ display: "flex", gap: "10px", alignItems: "center", flexWrap: "wrap" }}>
            <span style={{ fontSize: "0.85em", color: "#64748b" }}><i className="fa-solid fa-filter"></i> Fechas:</span>
            <input type="date" className="doc-btn doc-btn-secondary" style={{ padding: "4px 8px", zIndex: 10, position: "relative" }} value={fechaInicio} onChange={e => setFechaInicio(e.target.value)} title="Fecha Inicio" />
            <span style={{ fontSize: "0.85em", color: "#64748b" }}>a</span>
            <input type="date" className="doc-btn doc-btn-secondary" style={{ padding: "4px 8px", zIndex: 10, position: "relative" }} value={fechaFin} onChange={e => setFechaFin(e.target.value)} title="Fecha Fin" />
          </div>
        </div>
      </div>
      <div className="fullwidth-list-body" style={{ overflowX: "auto" }}>
        {loadingRadicados ? (
          <p style={{ padding: "15px" }}>Cargando datos financieros...</p>
        ) : filteredList.length === 0 ? (
          <div className="correo-empty">
            <i className="fa-solid fa-inbox"></i>
            <p>No hay facturas para mostrar.</p>
          </div>
        ) : (
          <div style={{ padding: "0", width: "100%", height: "100%", display: "block" }}>
            <table className="finanzas-table" style={{ width: "100%", borderCollapse: "separate", borderSpacing: "0", textAlign: "left", minWidth: "900px", background: "white", borderRadius: "8px", overflow: "hidden", boxShadow: "0 1px 3px rgba(0,0,0,0.1)" }}>
              <thead>
                <tr style={{ backgroundColor: "#f8fafc", color: "#334155", fontSize: "0.85em", textTransform: "uppercase", letterSpacing: "0.05em" }}>
                  <th style={{ padding: "16px", borderBottom: "2px solid #e2e8f0" }}>Proveedor</th>
                  <th style={{ padding: "16px", borderBottom: "2px solid #e2e8f0" }}>N° Factura</th>
                  <th style={{ padding: "16px", borderBottom: "2px solid #e2e8f0" }}>Fecha Fact.</th>
                  <th style={{ padding: "16px", borderBottom: "2px solid #e2e8f0" }}>Fecha Venc.</th>
                  <th style={{ padding: "16px", borderBottom: "2px solid #e2e8f0" }}>Valores</th>
                  <th style={{ padding: "16px", borderBottom: "2px solid #e2e8f0" }}>Normas Aplicadas</th>
                  <th style={{ padding: "16px", borderBottom: "2px solid #e2e8f0" }}>Causación</th>
                  <th style={{ padding: "16px", borderBottom: "2px solid #e2e8f0" }}>N° Documento</th>
                  <th style={{ padding: "16px", borderBottom: "2px solid #e2e8f0" }}>Pago (Tesorería)</th>
                  <th style={{ padding: "16px", borderBottom: "2px solid #e2e8f0", textAlign: "center" }}>Acciones</th>
                </tr>
              </thead>
              <tbody>
                {filteredList.map((rad) => {
                  const doc = rad.documento_comercial;
                  const esContabilidad = userRole === "Contabilidad" || userRole === "Superadministrador";
                  const esTesoreria = userRole === "Tesorería" || userRole === "Superadministrador";
                  const checkDisabled = (!esContabilidad) || (rad.causado && userRole !== "Superadministrador");
                  const egresoDisabled = (!esContabilidad) || (rad.numero_egreso && userRole !== "Superadministrador");
                  const pagoDisabled = (!esTesoreria) || (rad.pagado && userRole !== "Superadministrador");

                  return (
                    <tr key={rad.id} style={{ transition: "background-color 0.2s", ":hover": { backgroundColor: "#f1f5f9" } }}>
                      <td style={{ padding: "16px", borderBottom: "1px solid #e2e8f0", color: "#1e293b", fontWeight: "500" }}>
                        {doc?.proveedor?.razon_social || "—"}
                      </td>
                      <td style={{ padding: "16px", borderBottom: "1px solid #e2e8f0" }}>
                        <span style={{ background: "#e2e8f0", padding: "4px 8px", borderRadius: "4px", fontSize: "0.9em", fontWeight: "600", color: "#334155" }}>
                          {doc?.numero_documento || "—"}
                        </span>
                      </td>
                      <td style={{ padding: "16px", borderBottom: "1px solid #e2e8f0", color: "#64748b", fontSize: "0.95em" }}>
                        {doc?.fecha_documento ? new Date(doc.fecha_documento).toLocaleDateString() : "—"}
                      </td>
                      <td style={{ padding: "16px", borderBottom: "1px solid #e2e8f0", color: "#64748b", fontSize: "0.95em" }}>
                        {doc?.fecha_vencimiento ? new Date(doc.fecha_vencimiento).toLocaleDateString() : "—"}
                      </td>
                      <td style={{ padding: "16px", borderBottom: "1px solid #e2e8f0", fontSize: "0.9em" }}>
                        <div style={{ display: "flex", flexDirection: "column", gap: "2px" }}>
                          <span style={{ color: "#64748b" }}>Sub: {formatCurrency(doc?.subtotal, doc?.moneda?.codigo)}</span>
                          <span style={{ color: "#64748b" }}>IVA: {formatCurrency(doc?.iva, doc?.moneda?.codigo)}</span>
                          <span style={{ fontWeight: "600", color: "#0f172a", marginTop: "2px" }}>Tot: {formatCurrency(doc?.total, doc?.moneda?.codigo)}</span>
                        </div>
                      </td>
                      <td style={{ padding: "16px", borderBottom: "1px solid #e2e8f0", fontSize: "0.85em", color: "#475569" }}>
                        {(!rad.normas_reparto || rad.normas_reparto.length === 0) ? (
                          <span style={{ fontStyle: "italic", color: "#94a3b8" }}>Sin normas</span>
                        ) : (
                          <div style={{ display: "flex", flexDirection: "column", gap: "4px" }}>
                            {rad.normas_reparto.map((nr, idx) => {
                              const proj = nr.proyecto || nr.norma_reparto?.proyecto;
                              const desc = nr.descripcion || nr.norma_reparto?.descripcion;
                              return (
                                <div key={idx} style={{ background: "#f8fafc", padding: "4px 8px", borderRadius: "4px", border: "1px solid #e2e8f0" }}>
                                  <strong>{nr.norma_reparto?.nombre || "N/A"}</strong> ({nr.porcentaje}%)
                                  <br/>
                                  <span style={{ color: "#1e293b", fontWeight: "600" }}>{formatCurrency((doc?.subtotal || 0) * (nr.porcentaje / 100), doc?.moneda?.codigo)}</span>
                                  {(proj || desc) && (
                                    <div style={{ fontSize: "0.8em", color: "#64748b", marginTop: "2px" }}>
                                      {proj ? <span><strong>Proj:</strong> {proj} </span> : null}
                                      {desc ? <span>| <strong>Desc:</strong> {desc}</span> : null}
                                    </div>
                                  )}
                                </div>
                              );
                            })}
                          </div>
                        )}
                      </td>
                      <td style={{ padding: "16px", borderBottom: "1px solid #e2e8f0" }}>
                        <div style={{ display: "flex", alignItems: "center", gap: "10px" }}>
                          <input
                            type="checkbox"
                            checked={rad.causado}
                            onChange={() => handleCausarChange(rad)}
                            disabled={checkDisabled}
                            style={{ 
                              cursor: checkDisabled ? "not-allowed" : "pointer",
                              width: "18px", height: "18px", accentColor: "var(--primary-color)"
                            }}
                          />
                          {rad.fecha_causacion ? (
                            <span style={{ fontSize: "0.85em", color: "#10b981", fontWeight: "500", display: "flex", alignItems: "center", gap: "4px" }}>
                              <i className="fa-solid fa-check-circle"></i> {new Date(rad.fecha_causacion).toLocaleDateString()}
                            </span>
                          ) : (
                            <span style={{ fontSize: "0.85em", color: "#94a3b8" }}>Pendiente</span>
                          )}
                        </div>
                      </td>
                      <td style={{ padding: "16px", borderBottom: "1px solid #e2e8f0" }}>
                        <input
                          type="text"
                          defaultValue={rad.numero_egreso || ""}
                          onBlur={(e) => handleEgresoBlur(e, rad)}
                          onKeyDown={(e) => { if (e.key === "Enter") e.target.blur(); }}
                          disabled={egresoDisabled}
                          placeholder="Ej. EG-1023"
                          style={{ 
                            padding: "8px 12px", width: "120px", border: "1px solid #cbd5e1", 
                            borderRadius: "6px", backgroundColor: egresoDisabled ? "#f1f5f9" : "white",
                            color: "#334155", fontSize: "0.95em", outline: "none", transition: "border-color 0.2s",
                            cursor: egresoDisabled ? "not-allowed" : "text"
                          }}
                        />
                      </td>
                      <td style={{ padding: "16px", borderBottom: "1px solid #e2e8f0" }}>
                        <div style={{ display: "flex", alignItems: "center", gap: "10px" }}>
                          <input
                            type="checkbox"
                            checked={rad.pagado || false}
                            onChange={() => handlePagarChange(rad)}
                            disabled={pagoDisabled}
                            style={{ 
                              cursor: pagoDisabled ? "not-allowed" : "pointer",
                              width: "18px", height: "18px", accentColor: "var(--primary-color)"
                            }}
                          />
                          {rad.fecha_pago ? (
                            <span style={{ fontSize: "0.85em", color: "#10b981", fontWeight: "500", display: "flex", alignItems: "center", gap: "4px" }}>
                              <i className="fa-solid fa-check-circle"></i> {new Date(rad.fecha_pago).toLocaleDateString()}
                            </span>
                          ) : (
                            <span style={{ fontSize: "0.85em", color: "#94a3b8" }}>Pendiente</span>
                          )}
                        </div>
                        <div style={{ marginTop: "8px" }}>
                          {(!rad.comprobantes_subidos || userRole === "Superadministrador") ? (
                            <button 
                              className={`doc-btn ${!esTesoreria ? "doc-btn-secondary" : "doc-btn-primary"}`} 
                              style={{ padding: "4px 8px", fontSize: "0.75em", cursor: !esTesoreria ? "not-allowed" : "pointer", opacity: !esTesoreria ? 0.6 : 1, display: "inline-block" }}
                              title="Subir comprobante(s) de pago"
                              disabled={!esTesoreria}
                              onClick={() => setModalComprobante({ open: true, radicadoId: rad.id })}
                            >
                              <i className="fa-solid fa-upload"></i> Subir Comp.
                            </button>
                          ) : (
                            <button 
                              className="doc-btn doc-btn-secondary" 
                              style={{ padding: "4px 8px", fontSize: "0.75em", display: "inline-block" }}
                              onClick={() => handleVerComprobante(rad)}
                              title="Ver anexos subidos"
                            >
                              <i className="fa-solid fa-eye"></i> Ver Anexos
                            </button>
                          )}
                        </div>
                      </td>
                      <td style={{ padding: "16px", borderBottom: "1px solid #e2e8f0", textAlign: "center" }}>
                        <button 
                          className="doc-btn doc-btn-primary"
                          onClick={() => openSaia(rad, "finanzas")}
                          style={{ padding: "8px 16px", fontSize: "0.9em", display: "inline-flex", alignItems: "center", gap: "6px" }}
                        >
                          <i className="fa-solid fa-eye"></i> Ver
                        </button>
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        )}
      </div>

      <ModalSubirComprobante 
        isOpen={modalComprobante.open} 
        onClose={() => setModalComprobante({ open: false, radicadoId: null })}
        radicadoId={modalComprobante.radicadoId}
        obtenerToken={obtenerToken}
        onUploadSuccess={() => {
          if (handleComprobantesSubidosLocal) {
            handleComprobantesSubidosLocal(modalComprobante.radicadoId);
          }
          setModalComprobante({ open: false, radicadoId: null });
        }}
      />
    </div>
  );
}
