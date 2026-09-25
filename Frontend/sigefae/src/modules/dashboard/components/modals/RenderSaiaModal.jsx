import { useRef, useCallback, useEffect, useState } from "react";
import { formatCurrency, isFinalState } from "../../helpers/formatters.js";
import RenderFlujoAprobacion from "../RenderFlujoAprobacion.jsx";
import RenderTrazabilidad from "../RenderTrazabilidad.jsx";
import RenderNormasReparto from "../RenderNormasReparto.jsx";
import RenderComentarios from "../RenderComentarios.jsx";

export default function RenderSaiaModal({
  saiaModalOpen, setSaiaModalOpen, saiaRadicado, setSaiaRadicado, saiaActiveTab, setSaiaActiveTab,
  saiaAnexoIdx, setSaiaAnexoIdx, saiaPdfUrl,
  tareasFlujo, historialTrazabilidad, normasRepartoRadicado, comentarios,
  nuevoComentario, setNuevoComentario, enviandoComentario, handleEnviarComentario,
  generandoPdf, handleDescargarExpediente,
  handleVerAnexo, handleDescargarAnexo, handleBorrarAnexo, handleSubirAnexo,
  completandoTarea, handleCompletarTarea,
  solicitarRechazo, marcarCompletado, adminRechazar,
  esAdmin, esUsuario, userId, userRol,
  setPdfEditor, activeTab, setSelectedTareaId, setSelectedRadicadoId,
  setShowDevolverModal, setDevolverForm,
  openNormaModal, handleEliminarNorma, activosFijosRadicado, openActivoModal, handleEliminarActivoFijo, readOnly: externalReadOnly
}) {
  if (!saiaModalOpen || !saiaRadicado) return null;
  const archivosPdf = (saiaRadicado.archivos || []).filter(a => {
    const ext = (a.extension || a.nombre?.split('.').pop() || '').toLowerCase();
    return ['pdf', 'png', 'jpg', 'jpeg', 'webp', 'gif'].includes(ext);
  });
  const anexoActual = archivosPdf[saiaAnexoIdx];
  const readOnly = isFinalState(saiaRadicado.estado_posesion) || externalReadOnly;
  const esContabilidad = userRol === "Contabilidad";
  const puedeGestionarRecurso = (creadoPorId) => {
    if (esAdmin) return true;
    const propietario = Number(creadoPorId || 0);
    return propietario > 0 && propietario === Number(userId);
  };

  // ── Resizable sidebar ──────────────────────────────────────
  const SIDEBAR_MIN = 320;
  const SIDEBAR_MAX_RATIO = 0.75; // 75% del viewport
  const STORAGE_KEY = "saia_sidebar_width";
  const [sidebarWidth, setSidebarWidth] = useState(() => {
    const saved = parseInt(localStorage.getItem(STORAGE_KEY));
    return saved && saved >= SIDEBAR_MIN ? saved : 540;
  });
  const isDragging = useRef(false);
  const dragStartX = useRef(0);
  const dragStartWidth = useRef(0);
  const sidebarRef = useRef(null);

  const onMouseDown = useCallback((e) => {
    e.preventDefault();
    isDragging.current = true;
    dragStartX.current = e.clientX;
    dragStartWidth.current = sidebarWidth;
    document.body.style.cursor = "ew-resize";
    document.body.style.userSelect = "none";
  }, [sidebarWidth]);

  useEffect(() => {
    const onMouseMove = (e) => {
      if (!isDragging.current) return;
      const delta = dragStartX.current - e.clientX; // sidebar crece hacia la izquierda
      const maxW = Math.floor(window.innerWidth * SIDEBAR_MAX_RATIO);
      const newW = Math.max(SIDEBAR_MIN, Math.min(maxW, dragStartWidth.current + delta));
      setSidebarWidth(newW);
    };
    const onMouseUp = () => {
      if (!isDragging.current) return;
      isDragging.current = false;
      document.body.style.cursor = "";
      document.body.style.userSelect = "";
      // Guardar en localStorage
      if (sidebarRef.current) {
        localStorage.setItem(STORAGE_KEY, sidebarRef.current.offsetWidth);
      }
    };
    window.addEventListener("mousemove", onMouseMove);
    window.addEventListener("mouseup", onMouseUp);
    return () => {
      window.removeEventListener("mousemove", onMouseMove);
      window.removeEventListener("mouseup", onMouseUp);
    };
  }, []);
  // ────────────────────────────────────────────────────────────

  return (
    <div className="saia-overlay" onClick={() => setSaiaModalOpen(false)}>
      <div className="saia-container" onClick={e => e.stopPropagation()}>
        <div className="saia-topbar">
          <h3><i className="fa-solid fa-stamp"></i> Radicado #{saiaRadicado.numero_radicado}</h3>
          {archivosPdf.length > 0 && (
            <select className="saia-anexo-select" value={saiaAnexoIdx} onChange={e => setSaiaAnexoIdx(Number(e.target.value))}>
              {archivosPdf.map((a, i) => <option key={a.id} value={i}>{a.nombre}</option>)}
            </select>
          )}
          <button className="doc-btn doc-btn-secondary" onClick={() => setSaiaModalOpen(false)} style={{ padding: '6px 14px', fontSize: '0.85em' }}>
            <i className="fa-solid fa-xmark"></i> Cerrar
          </button>
        </div>
        <div className="saia-body">
          <div className="saia-pdf-area">
            {saiaPdfUrl ? (
              anexoActual?.extension?.toLowerCase() !== 'pdf' && !anexoActual?.nombre?.toLowerCase().endsWith('.pdf') ? (
                <div style={{ width: '100%', height: '100%', display: 'flex', alignItems: 'center', justifyContent: 'center', backgroundColor: '#0f172a' }}>
                  <img src={saiaPdfUrl} alt={anexoActual?.nombre} style={{ maxWidth: '100%', maxHeight: '100%', objectFit: 'contain' }} />
                </div>
              ) : (
                <iframe className="saia-pdf-frame" src={saiaPdfUrl} title="Visor de Documentos" />
              )
            ) : (
              <div className="saia-pdf-empty"><i className="fa-solid fa-file-pdf" style={{ fontSize: '3em' }}></i><p>No hay anexos visualizables disponibles</p></div>
            )}
          </div>
          {/* ── Drag handle ── */}
          <div
            className="saia-resize-handle"
            onMouseDown={onMouseDown}
            title="Arrastrar para redimensionar"
          />
          <div ref={sidebarRef} className="saia-sidebar" style={{ width: sidebarWidth, minWidth: SIDEBAR_MIN, flexShrink: 0 }}>
            <div className="saia-tabs">
              {['info', 'flujo', 'trazabilidad', 'normas', 'comentarios', 'acciones'].map(tab => (
                <button key={tab} className={`saia-tab ${saiaActiveTab === tab ? 'active' : ''}`} onClick={() => setSaiaActiveTab(tab)} title={tab}>
                  {tab === 'info' && <i className="fa-solid fa-circle-info"></i>}
                  {tab === 'flujo' && <i className="fa-solid fa-route"></i>}
                  {tab === 'trazabilidad' && <i className="fa-solid fa-clock-rotate-left"></i>}
                  {tab === 'normas' && <i className="fa-solid fa-chart-pie"></i>}
                  {tab === 'comentarios' && <i className="fa-solid fa-comments"></i>}
                  {tab === 'acciones' && <i className="fa-solid fa-bolt"></i>}
                </button>
              ))}
            </div>
            <div className="saia-tab-content">
              {saiaActiveTab === 'info' && (
                <>
                  <div className="doc-section">
                    <h4><i className="fa-solid fa-circle-info"></i> Información del Radicado</h4>
                    <div className="doc-grid">
                      <div className="doc-field"><label>Número</label><span>{saiaRadicado.numero_radicado}</span></div>
                      <div className="doc-field"><label>Fecha</label><span>{new Date(saiaRadicado.fecha_radicacion).toLocaleString()}</span></div>
                      <div className="doc-field"><label>Tipo</label><span>{saiaRadicado.tipo_radicacion?.nombre || '—'}</span></div>
                      <div className="doc-field"><label>Ruta</label><span>{saiaRadicado.ruta?.nombre || '—'}</span></div>
                      <div className="doc-field"><label>Área</label><span>{typeof saiaRadicado.ruta?.area === 'string' ? saiaRadicado.ruta.area : (saiaRadicado.ruta?.area?.nombre || '—')}</span></div>
                      <div className="doc-field"><label>Método Pago</label><span>{saiaRadicado.metodo_pago?.nombre || '—'}</span></div>
                      <div className="doc-field"><label>Estado</label><span>{saiaRadicado.estado_posesion || '—'}</span></div>
                      <div className="doc-field"><label>Paso Actual</label><span>{saiaRadicado.paso_actual?.nombre || 'Inicio'}</span></div>
                      <div className="doc-field"><label>Responsable</label><span>{saiaRadicado.usuario_actual?.nombre || '—'}</span></div>
                    </div>
                  </div>
                  {saiaRadicado.documento_comercial && (
                    <>
                      <div className="doc-section">
                        <h4><i className="fa-solid fa-file-invoice"></i> Documento Comercial</h4>
                        <div className="doc-grid">
                          <div className="doc-field"><label>Tipo</label><span>{saiaRadicado.documento_comercial.tipo}</span></div>
                          <div className="doc-field"><label>Número</label><span>{saiaRadicado.documento_comercial.numero_documento}</span></div>
                          <div className="doc-field"><label>Fecha Emisión</label><span>{new Date(saiaRadicado.documento_comercial.fecha_documento).toLocaleDateString()}</span></div>
                          <div className="doc-field"><label>Fecha Venc.</label><span>{saiaRadicado.documento_comercial.fecha_vencimiento ? new Date(saiaRadicado.documento_comercial.fecha_vencimiento).toLocaleDateString() : '—'}</span></div>
                          <div className="doc-field"><label>Forma Pago</label><span>{saiaRadicado.documento_comercial.forma_pago || '—'}</span></div>
                          <div className="doc-field"><label>Proveedor</label><span>{saiaRadicado.documento_comercial.proveedor?.razon_social || '—'}</span></div>
                          <div className="doc-field"><label>Receptor</label><span>{saiaRadicado.documento_comercial.receptor?.nombre || '—'}</span></div>
                        </div>
                      </div>
                      <div className="doc-section">
                        <h4><i className="fa-solid fa-calculator"></i> Valores</h4>
                        <div className="doc-totals">
                          <div className="doc-total-row"><span>Subtotal</span><span>{formatCurrency(saiaRadicado.documento_comercial.subtotal)}</span></div>
                          <div className="doc-total-row"><span>IVA</span><span>{formatCurrency(saiaRadicado.documento_comercial.iva)}</span></div>
                          <div className="doc-total-row total-final"><span>Total</span><span>{formatCurrency(saiaRadicado.documento_comercial.total)}</span></div>
                        </div>
                      </div>
                      {saiaRadicado.documento_comercial.detalles?.length > 0 && (
                        <div className="doc-section">
                          <h4><i className="fa-solid fa-list"></i> Ítems ({saiaRadicado.documento_comercial.detalles.length})</h4>
                          <table className="doc-items-table">
                            <thead><tr><th>Desc.</th><th>Cant.</th><th>V.Unit</th><th>Total</th></tr></thead>
                            <tbody>{saiaRadicado.documento_comercial.detalles.map(item => (
                              <tr key={item.id}><td>{item.descripcion}</td><td>{item.cantidad}</td><td>{formatCurrency(item.valor_unitario)}</td><td>{formatCurrency(item.total)}</td></tr>
                            ))}</tbody>
                          </table>
                        </div>
                      )}
                    </>
                  )}
                  {saiaRadicado.qr?.url && (
                    <div className="doc-section" style={{ textAlign: 'center' }}>
                      <h4><i className="fa-solid fa-qrcode"></i> QR Expediente</h4>
                      <img src={`https://api.qrserver.com/v1/create-qr-code/?size=180x180&data=${encodeURIComponent(saiaRadicado.qr.url)}`} alt="QR" style={{ margin: '12px auto', display: 'block', borderRadius: 8, border: '1px solid #e5e7eb' }} />
                    </div>
                  )}
                </>
              )}
              {saiaActiveTab === 'flujo' && <RenderFlujoAprobacion tareasFlujo={tareasFlujo} />}
              {saiaActiveTab === 'trazabilidad' && <RenderTrazabilidad historialTrazabilidad={historialTrazabilidad} />}
              {saiaActiveTab === 'normas' && <RenderNormasReparto normasRepartoRadicado={normasRepartoRadicado} activosFijosRadicado={activosFijosRadicado} radicadoId={saiaRadicado.id} readOnly={readOnly} esAdmin={esAdmin} puedeGestionarRecurso={puedeGestionarRecurso} openNormaModal={openNormaModal} handleEliminarNorma={handleEliminarNorma} openActivoModal={openActivoModal} handleEliminarActivoFijo={handleEliminarActivoFijo} subtotalRadicado={saiaRadicado.documento_comercial?.subtotal || saiaRadicado.documento_comercial?.total || 0} />}
              {saiaActiveTab === 'comentarios' && <div className="saia-comments-wrapper"><RenderComentarios comentarios={comentarios} nuevoComentario={nuevoComentario} setNuevoComentario={setNuevoComentario} enviandoComentario={enviandoComentario} handleEnviarComentario={handleEnviarComentario} radicadoId={saiaRadicado.id} userId={userId} readOnly={readOnly} /></div>}
              {saiaActiveTab === 'acciones' && (
                <div className="doc-section" style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
                  <h4><i className="fa-solid fa-bolt"></i> Acciones</h4>
                  <button className="doc-btn doc-btn-secondary" onClick={() => handleDescargarExpediente(saiaRadicado, tareasFlujo, historialTrazabilidad)} disabled={generandoPdf}>
                    <i className={`fa-solid ${generandoPdf ? 'fa-spinner fa-spin' : 'fa-file-pdf'}`}></i> {generandoPdf ? 'Generando...' : 'Descargar Expediente'}
                  </button>
                  {anexoActual && !readOnly && (
                    <button className="doc-btn doc-btn-primary" onClick={() => setPdfEditor({ open: true, archivoId: anexoActual.id, archivoNombre: anexoActual.nombre, radicadoId: saiaRadicado.id })}>
                      <i className="fa-solid fa-pen-to-square"></i> Abrir Editor PDF
                    </button>
                  )}
                  <div className="doc-section" style={{ margin: 0, padding: 10 }}>
                    <h4 style={{ fontSize: '0.75em', marginBottom: 8 }}><i className="fa-solid fa-paperclip"></i> Anexos ({(saiaRadicado.archivos || []).length})</h4>
                    {(saiaRadicado.archivos || []).length === 0 ? <p style={{ fontSize: '0.8em', color: '#6b7280' }}>No hay anexos.</p> : (
                      <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
                        {saiaRadicado.archivos.map(arch => (
                          <div key={arch.id} style={{ display: 'flex', alignItems: 'center', gap: 8, padding: '6px 8px', background: '#f1f5f9', borderRadius: 6, fontSize: '0.8em' }}>
                            <i className={`fa-solid fa-file${['pdf','png','jpg','jpeg','webp','gif'].includes((arch.extension||'').toLowerCase()) ? (arch.extension?.toLowerCase()==='pdf' ? '-pdf' : '-image') : ''}`} style={{ color: 'var(--pardo-blue)' }}></i>
                            <span style={{ flex: 1, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }} title={arch.nombre}>{arch.nombre}</span>
                            <button className="btn-icon btn-edit" style={{ width: 26, height: 26, flexShrink: 0 }} onClick={() => handleVerAnexo(arch.id, arch.nombre)} title="Ver / Previsualizar"><i className="fa-solid fa-eye" style={{ fontSize: '0.75em' }}></i></button>
                            <button className="btn-icon btn-edit" style={{ width: 26, height: 26, flexShrink: 0 }} onClick={() => handleDescargarAnexo(arch.id, arch.nombre)} title="Descargar"><i className="fa-solid fa-download" style={{ fontSize: '0.75em' }}></i></button>
                            {(esAdmin || puedeGestionarRecurso(arch.creado_por_id || arch.creado_por?.id)) && !readOnly && (
                              <button className="btn-icon btn-toggle" style={{ width: 26, height: 26, flexShrink: 0 }} onClick={() => handleBorrarAnexo(arch, saiaRadicado.id)} title="Eliminar"><i className="fa-solid fa-xmark" style={{ fontSize: '0.75em' }}></i></button>
                            )}
                          </div>
                        ))}
                      </div>
                    )}
                  </div>
                  {!readOnly && (
                    <div className="doc-section" style={{ margin: 0, padding: 10 }}>
                      <h4 style={{ fontSize: '0.75em', marginBottom: 8 }}><i className="fa-solid fa-cloud-arrow-up"></i> Adjuntar archivo</h4>
                      <input type="file" id="anexo-saia-input" onChange={(e) => handleSubirAnexo(e, saiaRadicado.id, setSelectedTareaId, setSelectedRadicadoId, setSaiaRadicado, saiaModalOpen, saiaRadicado)} style={{ display: 'none' }} />
                      <button className="doc-btn doc-btn-secondary" onClick={() => document.getElementById('anexo-saia-input').click()}><i className="fa-solid fa-upload"></i> Seleccionar archivo</button>
                    </div>
                  )}
                  {(() => {
                    const tareaActiva = tareasFlujo.find(t => t.estado?.nombre === "En Proceso");
                    const esResponsable = tareaActiva && (String(tareaActiva.usuario_asignado_id) === String(userId) || String(saiaRadicado.usuario_actual?.id) === String(userId));
                    const botones = [];

                    // ── Botones para el usuario responsable (Aprobador asignado) ──
                    if (esResponsable && !readOnly) {
                      botones.push(<button key="completar" className="doc-btn doc-btn-primary" onClick={() => handleCompletarTarea(tareaActiva.id, saiaRadicado.id)} disabled={completandoTarea}><i className={`fa-solid ${completandoTarea ? 'fa-spinner fa-spin' : 'fa-check'}`}></i> {completandoTarea ? 'Completando...' : 'Marcar como Completado'}</button>);
                      botones.push(<button key="devolver" className="doc-btn doc-btn-secondary" onClick={() => { setDevolverForm({ tarea_destino_id: "", observacion: "", retorno_directo: true }); setShowDevolverModal(true); }} style={{ borderColor: 'var(--pardo-red)', color: 'var(--pardo-red)' }}><i className="fa-solid fa-reply"></i> Devolver</button>);
                    }

                    // ── Solicitar rechazo (cualquier usuario no-admin, no-contabilidad) ──
                    if (esUsuario && !esContabilidad && !readOnly) botones.push(<button key="solicitar" className="doc-btn doc-btn-secondary" onClick={() => solicitarRechazo(saiaRadicado.id)}><i className="fa-solid fa-ban"></i> Solicitar Rechazo</button>);

                    // ── Botones para Contabilidad (devolver y rechazar) ──
                    if (esContabilidad && saiaRadicado.estado_posesion !== "Rechazado") {
                      // Para devolver: usar la última tarea completada como "tarea activa"
                      const tareasCompletadas = tareasFlujo.filter(t => t.estado?.nombre === "Completada");
                      const ultimaTareaCompletada = tareasCompletadas.length > 0 ? tareasCompletadas[tareasCompletadas.length - 1] : null;
                      if (ultimaTareaCompletada) {
                        botones.push(
                          <button key="conta-devolver" className="doc-btn doc-btn-secondary" onClick={() => { setDevolverForm({ tarea_destino_id: "", observacion: "", retorno_directo: false }); setShowDevolverModal(true); }} style={{ borderColor: 'var(--pardo-red)', color: 'var(--pardo-red)' }}>
                            <i className="fa-solid fa-reply"></i> Devolver Documento
                          </button>
                        );
                      }
                      botones.push(
                        <button key="conta-rechazar" className="doc-btn doc-btn-danger" onClick={() => adminRechazar(saiaRadicado.id, () => setSaiaModalOpen(false))}>
                          <i className="fa-solid fa-ban"></i> Rechazar Documento
                        </button>
                      );
                    }

                    // ── Botones admin (completar y rechazar) ──
                    if (!readOnly && esAdmin) {
                      botones.push(<button key="admin-completar" className="doc-btn doc-btn-primary" onClick={() => marcarCompletado(saiaRadicado.id)}><i className="fa-solid fa-check"></i> Marcar Completado (Admin)</button>);
                      botones.push(<button key="admin-rechazar" className="doc-btn doc-btn-danger" onClick={() => adminRechazar(saiaRadicado.id, () => setSaiaModalOpen(false))}><i className="fa-solid fa-ban"></i> Rechazar Definitivo</button>);
                    }
                    return <>{botones}</>;
                  })()}
                </div>
              )}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}