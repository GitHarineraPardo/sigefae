import { formatCurrency } from "../helpers/formatters";

export default function RenderDocumentos({
  selectedDocId, setSelectedDocId, docDetail, loadingDocs,
  searchDocs, setSearchDocs, sortDocs, setSortDocs, getFilteredDocs,
  esAdmin, openCrearDocModal, openRadicarModal
}) {
  const filteredList = getFilteredDocs();
  return (
    <div className="correos-container">
      <div className="correos-list">
        <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", padding: "0 15px" }}>
          <h3>Pendientes de Revisión ({filteredList.length})</h3>
          {esAdmin && (
            <button className="doc-btn doc-btn-primary" onClick={openCrearDocModal}>
              <i className="fa-solid fa-plus"></i> Crear Documento Manual
            </button>
          )}
        </div>
        <div className="list-controls">
          <input type="text" placeholder="Buscar factura o proveedor..." value={searchDocs} onChange={e => setSearchDocs(e.target.value)} />
          <select value={sortDocs} onChange={e => setSortDocs(e.target.value)}>
            <option value="fecha_desc">Más recientes</option>
            <option value="fecha_asc">Más antiguos</option>
            <option value="estado">Por Estado</option>
          </select>
        </div>
        {loadingDocs ? <p style={{ padding: "15px" }}>Cargando documentos...</p> : filteredList.length === 0 ? (
          <p style={{ padding: "15px", color: "#6b7280" }}>No hay documentos pendientes.</p>
        ) : (
          filteredList.map((doc) => (
            <div key={doc.id} className={`correo-item ${selectedDocId === doc.id ? "active" : ""}`} onClick={() => setSelectedDocId(doc.id)}>
              <div className="correo-item-header"><strong>{doc.proveedor?.razon_social || "Sin proveedor"}</strong><span className="correo-date">{new Date(doc.fecha_documento).toLocaleDateString()}</span></div>
              <div className="correo-item-subject">{doc.tipo} - {doc.numero_documento}</div>
              <div className="correo-item-status"><span className="status-badge doc-pendiente">{formatCurrency(doc.total)}</span></div>
            </div>
          ))
        )}
      </div>

      <div className="correo-detail">
        {!selectedDocId ? (
          <div className="correo-empty"><i className="fa-solid fa-file-invoice"></i><p>Selecciona un documento para revisarlo</p></div>
        ) : !docDetail ? <p style={{ padding: "20px" }}>Cargando detalle...</p> : (
          <div className="doc-detail-content">
            <div className="doc-header">
              <div className="doc-header-top"><h2>{docDetail.tipo} #{docDetail.numero_documento}</h2><span className="doc-total">{formatCurrency(docDetail.total)}</span></div>
              {docDetail.cufe && <p className="doc-cufe"><strong>CUFE:</strong> {docDetail.cufe}</p>}
            </div>
            <div className="doc-body">
              <div className="doc-section">
                <h4><i className="fa-solid fa-circle-info"></i> Información General</h4>
                <div className="doc-grid">
                  <div className="doc-field"><label>Tipo</label><span>{docDetail.tipo}</span></div>
                  <div className="doc-field"><label>Número Documento</label><span>{docDetail.numero_documento}</span></div>
                  <div className="doc-field"><label>Fecha Emisión</label><span>{new Date(docDetail.fecha_documento).toLocaleDateString()}</span></div>
                  <div className="doc-field"><label>Fecha Vencimiento</label><span>{docDetail.fecha_vencimiento ? new Date(docDetail.fecha_vencimiento).toLocaleDateString() : "—"}</span></div>
                  <div className="doc-field"><label>Forma Pago</label><span>{docDetail.forma_pago || "—"}</span></div>
                  <div className="doc-field"><label>Moneda</label><span>{docDetail.moneda?.nombre || "—"}</span></div>
                  <div className="doc-field"><label>Asunto</label><span>{docDetail.correo?.asunto || "—"}</span></div>
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
              {docDetail.correo && (
                <div className="doc-section">
                  <h4><i className="fa-solid fa-envelope"></i> Correo de Origen</h4>
                  <p style={{ fontSize: "0.85em", color: "#6b7280" }}><strong>Asunto:</strong> {docDetail.correo.asunto}</p>
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