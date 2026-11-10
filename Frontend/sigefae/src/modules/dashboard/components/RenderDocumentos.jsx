import { useState } from "react";
import { formatCurrency } from "../helpers/formatters";
import { API } from "../constants/api";

const renderEventoBadge = (evento, compact = false) => {
  if (!evento) return null;
  const ev = String(evento).toUpperCase();
  let label = evento;
  let bg = "#f8fafc";
  let color = "#475569";
  let border = "#e2e8f0";
  let icon = "fa-clock";

  if (ev === "ACKNOWLEDGED" || ev === "030") {
    label = compact ? "Acuse (030)" : "Acuse de Recibo (030)";
    bg = "#f1f5f9";
    color = "#334155";
    border = "#cbd5e1";
    icon = "fa-check";
  } else if (ev === "RECEIVEDGOODS" || ev === "032") {
    label = compact ? "Recibo Bienes (032)" : "Recibo Bienes / Servicios (032)";
    bg = "#f0f9ff";
    color = "#0369a1";
    border = "#bae6fd";
    icon = "fa-box";
  } else if (ev === "ACCEPTED" || ev === "034" || ev === "033") {
    label = compact ? "Aceptada (034)" : "Aceptación Expresa (034)";
    bg = "#f0fdf4";
    color = "#15803d";
    border = "#bbf7d0";
    icon = "fa-circle-check";
  } else if (ev === "REJECTED" || ev === "031") {
    label = compact ? "Reclamada (031)" : "Reclamo / Rechazo (031)";
    bg = "#fef2f2";
    color = "#991b1b";
    border = "#fecaca";
    icon = "fa-circle-xmark";
  }

  return (
    <span
      style={{
        display: "inline-flex",
        alignItems: "center",
        gap: "4px",
        fontSize: compact ? "11px" : "11.5px",
        fontWeight: 500,
        padding: compact ? "1px 6px" : "3px 8px",
        borderRadius: "4px",
        backgroundColor: bg,
        color: color,
        border: `1px solid ${border}`,
        lineHeight: "1.35",
        whiteSpace: "nowrap",
      }}
    >
      <i className={`fa-solid ${icon}`} style={{ fontSize: "10px", opacity: 0.85 }}></i>
      <span>{label}</span>
    </span>
  );
};

export default function RenderDocumentos({
  selectedDocId, setSelectedDocId, docDetail, loadingDocs,
  searchDocs, setSearchDocs, sortDocs, setSortDocs, getFilteredDocs,
  esAdmin, openCrearDocModal, openRadicarModal,
  syncingInbound, syncResult, handleSyncInbound
}) {
  const [fechaInicio, setFechaInicio] = useState("");
  const [fechaFin, setFechaFin] = useState("");
  const getFechaPrincipal = (documento) => documento.fecha_recepcion_dian || documento.created_at;
  const getFechaCalendario = (fecha) => {
    if (!fecha) return "";
    const date = new Date(fecha);
    if (Number.isNaN(date.getTime())) return "";
    return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, "0")}-${String(date.getDate()).padStart(2, "0")}`;
  };
  const formatearFecha = (fecha) => fecha
    ? new Date(fecha).toLocaleDateString("es-CO", { day: "2-digit", month: "short", year: "numeric" })
    : "—";
  const filteredList = getFilteredDocs().filter((documento) => {
    const fecha = getFechaCalendario(getFechaPrincipal(documento));
    if (fechaInicio && (!fecha || fecha < fechaInicio)) return false;
    if (fechaFin && (!fecha || fecha > fechaFin)) return false;
    return true;
  });
  return (
    <div className="correos-container">
      <div className="correos-list">
        <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", padding: "0 15px", flexWrap: "wrap", gap: "8px" }}>
          <h3>Facturas Inbound ({filteredList.length})</h3>
          <div style={{ display: "flex", gap: "8px" }}>
            {handleSyncInbound && (
              <button
                className="doc-btn"
                onClick={() => handleSyncInbound(30)}
                disabled={syncingInbound}
                title="Consultar y descargar facturas electrónicas desde el Proveedor Tecnológico DIAN"
                style={{
                  backgroundColor: syncingInbound ? "#9ca3af" : "#0284c7",
                  color: "#ffffff",
                  display: "inline-flex",
                  alignItems: "center",
                  gap: "6px",
                  cursor: syncingInbound ? "not-allowed" : "pointer"
                }}
              >
                <i className={`fa-solid fa-arrows-rotate ${syncingInbound ? "fa-spin" : ""}`}></i>
                {syncingInbound ? "Sincronizando..." : "Sincronizar DIAN"}
              </button>
            )}
            {esAdmin && (
              <button className="doc-btn doc-btn-primary" onClick={openCrearDocModal}>
                <i className="fa-solid fa-plus"></i> Crear Manual
              </button>
            )}
          </div>
        </div>

        {syncResult && (
          <div style={{
            margin: "10px 15px",
            padding: "8px 12px",
            backgroundColor: "#ecfdf5",
            border: "1px solid #a7f3d0",
            borderRadius: "6px",
            fontSize: "12px",
            color: "#065f46",
            display: "flex",
            alignItems: "center",
            gap: "8px"
          }}>
            <i className="fa-solid fa-circle-check"></i>
            <span>
              Sincronización exitosa: <strong>{syncResult.total_encontrados}</strong> en DIAN, <strong>{syncResult.nuevos_procesados}</strong> nuevas descargadas, <strong>{syncResult.actualizados}</strong> actualizadas.
            </span>
          </div>
        )}

        <div className="list-controls">
          <input type="text" placeholder="Buscar factura o proveedor..." value={searchDocs} onChange={e => setSearchDocs(e.target.value)} />
          <select value={sortDocs} onChange={e => setSortDocs(e.target.value)}>
            <option value="fecha_desc">Más recientes</option>
            <option value="fecha_asc">Más antiguos</option>
            <option value="estado">Por Estado</option>
          </select>
        </div>
        <div className="inbound-date-filters">
          <div className="date-filter">
            <label htmlFor="inbound-fecha-inicio"><i className="fa-regular fa-calendar"></i> Recibido desde</label>
            <input id="inbound-fecha-inicio" type="date" value={fechaInicio} max={fechaFin || undefined} onChange={e => setFechaInicio(e.target.value)} />
          </div>
          <div className="date-filter">
            <label htmlFor="inbound-fecha-fin"><i className="fa-regular fa-calendar"></i> Recibido hasta</label>
            <input id="inbound-fecha-fin" type="date" value={fechaFin} min={fechaInicio || undefined} onChange={e => setFechaFin(e.target.value)} />
          </div>
          {(fechaInicio || fechaFin) && (
            <button className="clear-date-filter" type="button" onClick={() => { setFechaInicio(""); setFechaFin(""); }}>
              <i className="fa-solid fa-rotate-left"></i> Limpiar fechas
            </button>
          )}
        </div>
        {loadingDocs ? <p style={{ padding: "15px" }}>Cargando documentos...</p> : filteredList.length === 0 ? (
          <p style={{ padding: "15px", color: "#6b7280" }}>No hay documentos pendientes.</p>
        ) : (
          filteredList.map((doc) => {
            const tipoLabel = doc.tipo === "FACTURA_ELECTRONICA" ? "Factura" : doc.tipo;
            return (
              <div key={doc.id} className={`correo-item ${selectedDocId === doc.id ? "active" : ""}`} onClick={() => setSelectedDocId(doc.id)}>
                <div className="correo-item-header">
                  <strong>{doc.proveedor?.razon_social || "Sin proveedor"}</strong>
                  <span className="correo-date">{formatearFecha(getFechaPrincipal(doc))}</span>
                </div>
                <div className="correo-item-subject" style={{ color: "#64748b", fontSize: "12px", marginBottom: "6px" }}>
                  {tipoLabel} · #{doc.numero_documento}
                </div>
                <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", gap: "8px", marginTop: "4px" }}>
                  {renderEventoBadge(doc.evento_dian, true)}
                  <span style={{ fontWeight: 600, fontSize: "12.5px", color: "#1e293b", marginLeft: "auto" }}>
                    {formatCurrency(doc.total)}
                  </span>
                </div>
              </div>
            );
          })
        )}
      </div>

      <div className="correo-detail">
        {!selectedDocId ? (
          <div className="correo-empty"><i className="fa-solid fa-file-invoice"></i><p>Selecciona un documento para revisarlo</p></div>
        ) : !docDetail ? <p style={{ padding: "20px" }}>Cargando detalle...</p> : (
          <div className="doc-detail-content">
            <div className="doc-header">
              <div className="doc-header-top">
                <h2>{docDetail.tipo === "FACTURA_ELECTRONICA" ? "Factura Electrónica" : docDetail.tipo} #{docDetail.numero_documento}</h2>
                <span className="doc-total">{formatCurrency(docDetail.total)}</span>
              </div>
              {docDetail.evento_dian && (
                <div style={{ marginTop: "8px", display: "flex", alignItems: "center", gap: "8px", flexWrap: "wrap" }}>
                  <span style={{ fontSize: "12px", color: "#64748b" }}>Estado DIAN:</span>
                  {renderEventoBadge(docDetail.evento_dian, false)}
                  {docDetail.fecha_recepcion_dian && (
                    <span style={{ fontSize: "12px", color: "#94a3b8" }}>
                      · Recibido DIAN: {new Date(docDetail.fecha_recepcion_dian).toLocaleDateString()}
                    </span>
                  )}
                </div>
              )}
              {docDetail.cufe && <p className="doc-cufe" style={{ marginTop: "6px" }}><strong>CUFE:</strong> {docDetail.cufe}</p>}
              {docDetail.ldf && <p className="doc-cufe" style={{ marginTop: "2px", fontSize: "11px", color: "#6b7280" }}><strong>LDF:</strong> {docDetail.ldf}</p>}
            </div>

            <div className="doc-body">
              {(docDetail.ruta_pdf || docDetail.ruta_xml) && (
                <div className="doc-section">
                  <h4><i className="fa-solid fa-paperclip"></i> Archivos DIAN</h4>
                  <div style={{ display: "flex", gap: "10px", flexWrap: "wrap", padding: "6px 0" }}>
                    {docDetail.ruta_pdf && (
                      <a
                        href={`${API}/${docDetail.ruta_pdf}`}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="btn-doc-archivo"
                      >
                        <i className="fa-solid fa-file-pdf" style={{ color: "#dc2626", fontSize: "14px" }}></i>
                        <span>Ver Factura (PDF)</span>
                      </a>
                    )}
                    {docDetail.ruta_xml && (
                      <a
                        href={`${API}/${docDetail.ruta_xml}`}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="btn-doc-archivo"
                      >
                        <i className="fa-solid fa-file-code" style={{ color: "#0284c7", fontSize: "14px" }}></i>
                        <span>Ver XML (UBL)</span>
                      </a>
                    )}
                  </div>
                </div>
              )}

              <div className="doc-section">
                <h4><i className="fa-solid fa-circle-info"></i> Información General</h4>
                <div className="doc-grid">
                  <div className="doc-field"><label>Tipo</label><span>{docDetail.tipo}</span></div>
                  <div className="doc-field"><label>Número Documento</label><span>{docDetail.numero_documento}</span></div>
                  <div className="doc-field"><label>Fecha Emisión</label><span>{new Date(docDetail.fecha_documento).toLocaleDateString()}</span></div>
                  <div className="doc-field"><label>Fecha Vencimiento</label><span>{docDetail.fecha_vencimiento ? new Date(docDetail.fecha_vencimiento).toLocaleDateString() : "—"}</span></div>
                  <div className="doc-field"><label>Forma Pago</label><span>{docDetail.forma_pago || "—"}</span></div>
                  <div className="doc-field"><label>Moneda</label><span>{docDetail.moneda?.nombre || "—"}</span></div>
                  <div className="doc-field"><label>Asunto</label><span>{docDetail.asunto || "—"}</span></div>
                </div>
              </div>

              <div className="doc-section">
                <h4><i className="fa-solid fa-building"></i> Proveedor</h4>
                <div className="doc-grid">
                  <div className="doc-field"><label>Razón Social</label><span>{docDetail.proveedor?.razon_social || "—"}</span></div>
                  <div className="doc-field"><label>NIT</label><span>{docDetail.proveedor?.numero_documento || "—"}</span></div>
                </div>
              </div>

              <div className="doc-section">
                <h4><i className="fa-solid fa-user-tie"></i> Receptor</h4>
                <div className="doc-grid">
                  <div className="doc-field"><label>Nombre</label><span>{docDetail.receptor?.nombre || "—"}</span></div>
                  <div className="doc-field"><label>NIT</label><span>{docDetail.receptor?.numero_documento || "—"}</span></div>
                </div>
              </div>

              <div className="doc-section">
                <h4><i className="fa-solid fa-calculator"></i> Resumen Financiero</h4>
                <div className="doc-totals">
                  <div className="doc-total-row"><span>Subtotal</span><span>{formatCurrency(docDetail.subtotal)}</span></div>
                  <div className="doc-total-row"><span>IVA</span><span>{formatCurrency(docDetail.iva)}</span></div>
                  <div className="doc-total-row total-final"><span>Total</span><span>{formatCurrency(docDetail.total)}</span></div>
                </div>
              </div>

              {docDetail.detalles && docDetail.detalles.length > 0 && (
                <div className="doc-section">
                  <h4><i className="fa-solid fa-list"></i> Detalle de Ítems ({docDetail.detalles.length})</h4>
                  <table className="doc-items-table">
                    <thead><tr><th>Descripción</th><th>Cantidad</th><th>Valor Unit.</th><th>Total</th></tr></thead>
                    <tbody>{docDetail.detalles.map((item) => (
                      <tr key={item.id}><td>{item.descripcion}</td><td>{item.cantidad}</td><td>{formatCurrency(item.valor_unitario)}</td><td>{formatCurrency(item.total)}</td></tr>
                    ))}</tbody>
                  </table>
                </div>
              )}

              {docDetail.ldf && (
                <div className="doc-section">
                  <h4><i className="fa-solid fa-cloud-arrow-down"></i> Origen del Documento</h4>
                  <p style={{ fontSize: "0.85em", color: "#6b7280" }}>
                    <strong>Proveedor Tecnológico DIAN:</strong> Integración Inbound directa
                  </p>
                </div>
              )}
            </div>

            <div className="doc-actions">
              <button className="doc-btn doc-btn-primary" onClick={() => openRadicarModal(docDetail.id, docDetail.subtotal)}>
                <i className="fa-solid fa-stamp"></i> Aprobar para Radicación
              </button>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}