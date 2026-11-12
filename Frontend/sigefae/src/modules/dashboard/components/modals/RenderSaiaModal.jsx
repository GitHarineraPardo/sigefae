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
  esAdmin, esUsuario, userId, userRol, userCargo, userCargoId,
  setPdfEditor,  setSelectedTareaId, setSelectedRadicadoId,
  setShowDevolverModal, setDevolverForm,
  openNormaModal, handleEliminarNorma, activosFijosRadicado, openActivoModal, handleEliminarActivoFijo, readOnly: externalReadOnly,
  eventosDian = [], cargandoEventos = false, emitiendoEvento = false, handleEmitirEvento
}) {
  const [showModalDefinirFlujo, setShowModalDefinirFlujo] = useState(false);
  const [pasosSeleccionados, setPasosSeleccionados] = useState([]);

  // Estado para emisión de evento DIAN
  const [modalEvento, setModalEvento] = useState(null);
  const [formEvento, setFormEvento] = useState({ motivo: "", obs: "", comentario: "", codigoMotivo: "01" });
  const [errorEmision, setErrorEmision] = useState(null);

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

  const toggleSection = (event) => {
    const heading = event.target.closest("h4");
    if (!heading || !heading.parentElement?.classList.contains("doc-section")) return;
    if (event.target.closest("button, a, input, select, textarea")) return;
    event.stopPropagation();
    const section = heading.parentElement;
    const collapsed = section.classList.toggle("collapsed");
    heading.setAttribute("aria-expanded", String(!collapsed));
  };

  const handleSectionKeyDown = (event) => {
    if (event.key !== "Enter" && event.key !== " ") return;
    if (event.target.tagName !== "H4") return;
    event.preventDefault();
    toggleSection(event);
  };

  if (!saiaModalOpen || !saiaRadicado) return null;

  const archivos = Array.isArray(saiaRadicado.archivos) ? saiaRadicado.archivos : [];
  const archivosPdf = archivos.filter(a => {
    const ext = (a.extension || a.nombre?.split('.').pop() || '').toLowerCase();
    return ['pdf', 'png', 'jpg', 'jpeg', 'webp', 'gif'].includes(ext);
  });
  const anexoActual = archivosPdf[saiaAnexoIdx];
  const readOnly = isFinalState(saiaRadicado.estado_posesion) || externalReadOnly;
  const cargoResponsable = typeof saiaRadicado.cargo_actual === "string"
    ? saiaRadicado.cargo_actual
    : saiaRadicado.cargo_actual?.nombre || saiaRadicado.cargo_actual_nombre || "—";
  const esContabilidad = userRol === "Contabilidad";
  const puedeGestionarRecurso = (creadoPorId) => {
    if (esAdmin) return true;
    const propietario = Number(creadoPorId || 0);
    return propietario > 0 && propietario === Number(userId);
  };

  const esAuxOriente = (userCargo && userCargo.toLowerCase().includes("auxiliar admon oriente")) || esAdmin;
  const tareaActiva = (Array.isArray(tareasFlujo) ? tareasFlujo : []).find(t => t.estado?.nombre === "En Proceso");
  const tareasCompletadas = (Array.isArray(tareasFlujo) ? tareasFlujo : []).filter(t => t.estado?.nombre === "Completada");
  const esPrimerTurnoAuxOriente = 
    Boolean(esAuxOriente) &&
    tareasFlujo.length > 0 &&
    Boolean(tareaActiva) &&
    tareasFlujo[0].id === tareaActiva.id &&
    tareasCompletadas.length === 0;

  // ── Lógica de Eventos DIAN exclusiva para Auxiliar Admon Oriente ──
  const tieneCargoAuxOriente = Boolean(userCargo && userCargo.toLowerCase().includes("auxiliar admon oriente"));
  const esRutaFinalizada = isFinalState(saiaRadicado.estado_posesion) || saiaRadicado.estado_posesion === 'Completado';
  const tieneLDF = Boolean(saiaRadicado.documento_comercial?.ldf);
  const puedeEmitirEventos = tieneCargoAuxOriente && esRutaFinalizada && tieneLDF;

  const has030 = eventosDian.some(e => e.codigo_dian === "030" || e.tipo_evento === "ACKNOWLEDGED");
  const has032 = eventosDian.some(e => e.codigo_dian === "032" || e.tipo_evento === "RECEIVEDGOODS");
  const has034 = eventosDian.some(e => e.codigo_dian === "034" || e.tipo_evento === "ACCEPTED");
  const has031 = eventosDian.some(e => e.codigo_dian === "031" || e.tipo_evento === "REJECTED");

  // Inferencia RADIAN: Si existe un evento superior o la factura ya vino con un estado DIAN,
  // los eventos anteriores ya ocurrieron (provienen de Inbound o registro previo ante la DIAN)
  const docEvento = (saiaRadicado?.documento_comercial?.evento_dian || "").toUpperCase().trim();
  const effectiveHas030 = has030 || has032 || has034 || has031 || ["ACKNOWLEDGED", "RECEIVEDGOODS", "ACCEPTED", "REJECTED"].includes(docEvento);
  const effectiveHas032 = has032 || has034 || has031 || ["RECEIVEDGOODS", "ACCEPTED", "REJECTED"].includes(docEvento);
  const effectiveHas034 = has034 || docEvento === "ACCEPTED";
  const effectiveHas031 = has031 || docEvento === "REJECTED";

  const botonesEventos = [];
  if (!effectiveHas030) {
    botonesEventos.push({
      codigo: "030",
      tipo: "ACKNOWLEDGED",
      label: "Acuse (AC)",
      icono: "fa-envelope-open-text",
      clase: "acuse",
      motivoDefault: "Acuse",
      obsDefault: "Factura electrónica recibida",
    });
  }
  if (effectiveHas030 && !effectiveHas032) {
    botonesEventos.push({
      codigo: "032",
      tipo: "RECEIVEDGOODS",
      label: "Recibo Bienes (RB)",
      icono: "fa-box",
      clase: "recibo",
      motivoDefault: "Otro",
      obsDefault: "Bienes o servicios recibidos a satisfacción",
    });
  }
  if (effectiveHas032 && !effectiveHas034 && !effectiveHas031) {
    botonesEventos.push({
      codigo: "034",
      tipo: "ACCEPTED",
      label: "Aceptación (AE)",
      icono: "fa-circle-check",
      clase: "aceptar",
      motivoDefault: "Aceptación",
      obsDefault: "Factura aprobada y aceptada para pago",
    });
    botonesEventos.push({
      codigo: "031",
      tipo: "REJECTED",
      label: "Reclamo (RC)",
      icono: "fa-circle-xmark",
      clase: "rechazar",
      motivoDefault: "Servicio no prestado",
      obsDefault: "Reclamo registrado por disconformidad",
    });
  }

  const abrirModalEmision = (btn) => {
    // Validar pre-requisitos de eventos DIAN
    if (btn.codigo === "032" && !effectiveHas030) {
      alert("Para emitir el Recibo de Bienes o Servicios (032), primero debe registrarse el Acuse de Recibo (030).");
      return;
    }
    if ((btn.codigo === "034" || btn.codigo === "031") && !effectiveHas032) {
      alert("Para emitir la Aceptación (034) o Reclamo (031), primero debe registrarse el Recibo de Bienes o Servicios (032).");
      return;
    }

    setErrorEmision(null);
    setModalEvento(btn);
    setFormEvento({
      motivo: btn.motivoDefault,
      obs: btn.obsDefault,
      comentario: btn.obsDefault,
      codigoMotivo: "01"
    });
  };

  const submitEmitirEvento = async () => {
    if (!modalEvento || !handleEmitirEvento) return;
    setErrorEmision(null);
    try {
      await handleEmitirEvento(saiaRadicado.id, {
        tipo: modalEvento.codigo,
        motivo: formEvento.motivo,
        observacion: formEvento.obs,
        comentario: formEvento.comentario || formEvento.obs,
        codigo_motivo: modalEvento.codigo === "031" ? formEvento.codigoMotivo : undefined,
      });
      setModalEvento(null);
    } catch (err) {
      setErrorEmision(err.message || "Error al emitir el evento ante la DIAN");
    }
  };

  return (
    <>
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
          <div className="saia-pdf-area" style={{ position: "relative" }}>
            {/* ── Barra Flotante de Eventos DIAN para Auxiliar Admon Oriente ── */}
            {puedeEmitirEventos && botonesEventos.length > 0 && (
              <div className="saia-eventos-floating-bar">
                <div className="floating-bar-label">
                  <i className="fa-solid fa-stamp" style={{ color: "var(--pardo-blue)" }}></i>
                  <span>Eventos DIAN:</span>
                </div>
                <div className="floating-bar-buttons">
                  {botonesEventos.map(btn => (
                    <button
                      key={btn.codigo}
                      className="floating-btn"
                      onClick={() => abrirModalEmision(btn)}
                      disabled={emitiendoEvento}
                      title={`Emitir ${btn.label} ante la DIAN`}
                    >
                      <i className={`fa-solid ${btn.icono}`}></i>
                      <span>{btn.label}</span>
                    </button>
                  ))}
                </div>
              </div>
            )}

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
                  {tab === 'comentarios' && <i className="fa-solid fa-comment-dots"></i>}
                  {tab === 'acciones' && <i className="fa-solid fa-bolt"></i>}
                </button>
              ))}
            </div>
            <div className="saia-tab-content" onClick={toggleSection} onKeyDown={handleSectionKeyDown}>
              {saiaActiveTab === 'info' && (
                <>
                  <div className="doc-section">
                    <h4><i className="fa-solid fa-circle-info"></i> Información del Radicado</h4>
                    <div className="doc-grid">
                      <div className="doc-field"><label>Número</label><span>{saiaRadicado.numero_radicado}</span></div>
                      <div className="doc-field"><label>Fecha recibido DIAN</label><span>{saiaRadicado.documento_comercial?.fecha_recepcion_dian ? new Date(saiaRadicado.documento_comercial.fecha_recepcion_dian).toLocaleString() : '—'}</span></div>
                      <div className="doc-field"><label>Fecha radicación interna</label><span>{saiaRadicado.fecha_radicacion ? new Date(saiaRadicado.fecha_radicacion).toLocaleString() : '—'}</span></div>
                      <div className="doc-field"><label>Tipo</label><span>{saiaRadicado.tipo_radicacion?.nombre || '—'}</span></div>
                      <div className="doc-field"><label>Ruta</label><span>{saiaRadicado.ruta?.nombre || '—'}</span></div>
                      <div className="doc-field"><label>Área</label><span>{typeof saiaRadicado.ruta?.area === 'string' ? saiaRadicado.ruta.area : (saiaRadicado.ruta?.area?.nombre || '—')}</span></div>
                      <div className="doc-field"><label>Método Pago</label><span>{saiaRadicado.metodo_pago?.nombre || '—'}</span></div>
                      <div className="doc-field"><label>Estado</label><span>{saiaRadicado.estado_posesion || '—'}</span></div>
                      {!isFinalState(saiaRadicado.estado_posesion) && <div className="doc-field"><label>Paso Actual</label><span>{saiaRadicado.paso_actual?.nombre || 'Inicio'}</span></div>}
                      <div className="doc-field"><label>Cargo responsable</label><span>{cargoResponsable}</span></div>
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
                </>
              )}
              {saiaActiveTab === 'flujo' && <RenderFlujoAprobacion tareasFlujo={tareasFlujo} esPrimerTurnoAuxOriente={esPrimerTurnoAuxOriente} />}
              {saiaActiveTab === 'trazabilidad' && (
                <RenderTrazabilidad
                  historialTrazabilidad={historialTrazabilidad}
                  eventosDian={eventosDian}
                  cargandoEventos={cargandoEventos}
                />
              )}
              {saiaActiveTab === 'normas' && <RenderNormasReparto normasRepartoRadicado={normasRepartoRadicado} activosFijosRadicado={activosFijosRadicado} radicadoId={saiaRadicado.id} readOnly={readOnly} esAdmin={esAdmin} puedeGestionarRecurso={puedeGestionarRecurso} openNormaModal={openNormaModal} handleEliminarNorma={handleEliminarNorma} openActivoModal={openActivoModal} handleEliminarActivoFijo={handleEliminarActivoFijo} subtotalRadicado={saiaRadicado.documento_comercial?.subtotal || saiaRadicado.documento_comercial?.total || 0} />}
              {saiaActiveTab === 'comentarios' && <div className="saia-comments-wrapper"><RenderComentarios comentarios={comentarios} nuevoComentario={nuevoComentario} setNuevoComentario={setNuevoComentario} enviandoComentario={enviandoComentario} handleEnviarComentario={handleEnviarComentario} radicadoId={saiaRadicado.id} userId={userId} readOnly={readOnly} /></div>}
              {saiaActiveTab === 'acciones' && (
                <div className="doc-section" style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
                  <h4><i className="fa-solid fa-bolt"></i> Acciones</h4>
                  <button className="doc-btn doc-btn-secondary" onClick={() => handleDescargarExpediente(saiaRadicado, tareasFlujo, historialTrazabilidad, eventosDian, activosFijosRadicado)} disabled={generandoPdf}>
                    <i className={`fa-solid ${generandoPdf ? 'fa-spinner fa-spin' : 'fa-file-pdf'}`}></i> {generandoPdf ? 'Generando...' : 'Descargar Expediente'}
                  </button>
                  {anexoActual && !readOnly && (
                    <button className="doc-btn doc-btn-primary" onClick={() => setPdfEditor({ open: true, archivoId: anexoActual.id, archivoNombre: anexoActual.nombre, radicadoId: saiaRadicado.id })}>
                      <i className="fa-solid fa-pen-to-square"></i> Abrir Editor PDF
                    </button>
                  )}
                  <div className="doc-section" style={{ margin: 0, padding: 10 }}>
                    <h4 style={{ fontSize: '0.75em', marginBottom: 8 }}><i className="fa-solid fa-paperclip"></i> Anexos ({archivos.length})</h4>
                    {archivos.length === 0 ? <p style={{ fontSize: '0.8em', color: '#6b7280' }}>No hay anexos.</p> : (
                      <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
                        {archivos.map(arch => (
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
                    // Compare assignments by their stable database IDs.
                    const esResponsable = tareaActiva && (
                      Number(tareaActiva.usuario_asignado_id) === Number(userId) ||
                      Number(saiaRadicado.usuario_actual?.id) === Number(userId) ||
                      (Number(tareaActiva.cargo_asignado_id) > 0 && Number(tareaActiva.cargo_asignado_id) === Number(userCargoId)) ||
                      (Number(saiaRadicado.cargo_actual_id) > 0 && Number(saiaRadicado.cargo_actual_id) === Number(userCargoId))
                    );
                    const botones = [];

                    // ── Botones para el usuario responsable (Aprobador asignado) ──
                    if (esResponsable && !readOnly) {
                      botones.push(
                        <button
                          key="completar"
                          className="doc-btn doc-btn-primary"
                          onClick={() => {
                            if (esPrimerTurnoAuxOriente && tareasFlujo.length > 2) {
                              setPasosSeleccionados(tareasFlujo.slice(1, -1).map(t => t.id));
                              setShowModalDefinirFlujo(true);
                            } else {
                              handleCompletarTarea(tareaActiva.id, saiaRadicado.id);
                            }
                          }}
                          disabled={completandoTarea}
                        >
                          <i className={`fa-solid ${completandoTarea ? 'fa-spinner fa-spin' : (esPrimerTurnoAuxOriente && tareasFlujo.length > 2 ? 'fa-list-check' : 'fa-check')}`}></i>{' '}
                          {completandoTarea ? 'Completando...' : (esPrimerTurnoAuxOriente && tareasFlujo.length > 2 ? 'Definir Flujo y Continuar' : 'Marcar como Completado')}
                        </button>
                      );
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

    {showModalDefinirFlujo && (
      <div
        className="modal-overlay"
        style={{
          position: 'fixed',
          top: 0,
          left: 0,
          right: 0,
          bottom: 0,
          backgroundColor: 'rgba(0, 0, 0, 0.65)',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          zIndex: 99999
        }}
        onClick={() => setShowModalDefinirFlujo(false)}
      >
        <div
          className="modal-content"
          style={{
            maxWidth: '560px',
            width: '92%',
            backgroundColor: '#ffffff',
            borderRadius: '14px',
            padding: '24px',
            boxShadow: '0 20px 25px -5px rgba(0, 0, 0, 0.2), 0 10px 10px -5px rgba(0, 0, 0, 0.1)',
            maxHeight: '90vh',
            display: 'flex',
            flexDirection: 'column'
          }}
          onClick={(e) => e.stopPropagation()}
        >
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '16px' }}>
            <h3 style={{ margin: 0, fontSize: '1.2rem', display: 'flex', alignItems: 'center', gap: '8px', color: '#111827' }}>
              <i className="fa-solid fa-route" style={{ color: 'var(--pardo-red, #dc2626)' }}></i> Definir Flujo Oficial
            </h3>
            <button
              onClick={() => setShowModalDefinirFlujo(false)}
              style={{ background: 'transparent', border: 'none', cursor: 'pointer', fontSize: '1.2rem', color: '#9ca3af', padding: '4px' }}
              title="Cerrar"
            >
              <i className="fa-solid fa-xmark"></i>
            </button>
          </div>

          <div style={{ padding: '12px 14px', background: '#eff6ff', border: '1px solid #bfdbfe', borderRadius: '8px', marginBottom: '16px', fontSize: '0.88em', color: '#1e40af', lineHeight: 1.4 }}>
            <i className="fa-solid fa-circle-info" style={{ marginRight: '6px' }}></i>
            Como <strong>Auxiliar Administrativo Oriente</strong>, seleccione las revisiones por las que debe pasar oficialmente este radicado. Los pasos desmarcados serán omitidos.
          </div>

          <div style={{ flex: 1, overflowY: 'auto', display: 'flex', flexDirection: 'column', gap: '8px', marginBottom: '20px', paddingRight: '4px' }}>
            {/* Paso 1: Actual (Obligatorio) */}
            <div style={{ display: 'flex', alignItems: 'center', gap: '12px', padding: '10px 14px', borderRadius: '8px', background: '#f3f4f6', border: '1px solid #e5e7eb' }}>
              <i className="fa-solid fa-lock" style={{ color: '#9ca3af', fontSize: '0.9em' }}></i>
              <input type="checkbox" checked disabled style={{ width: '18px', height: '18px' }} />
              <div style={{ flex: 1 }}>
                <div style={{ fontWeight: 600, color: '#374151', fontSize: '0.95em' }}>Revisión Auxiliar Admon Oriente</div>
                <div style={{ fontSize: '0.8em', color: '#6b7280' }}>Paso actual — Obligatorio</div>
              </div>
              <span style={{ fontSize: '0.75em', padding: '3px 8px', borderRadius: '10px', background: '#e5e7eb', color: '#4b5563', fontWeight: 600 }}>Paso Actual</span>
            </div>

            {/* Pasos Intermedios a marcar/desmarcar */}
            {tareasFlujo.slice(1, -1).map((t) => {
              const isSelected = pasosSeleccionados.includes(t.id);
              const cargoNombre = typeof t.cargo_asignado === "string" ? t.cargo_asignado : t.cargo_asignado?.nombre || "Sin cargo asignado";
              return (
                <label
                  key={t.id}
                  style={{
                    display: 'flex',
                    alignItems: 'center',
                    gap: '12px',
                    padding: '10px 14px',
                    borderRadius: '8px',
                    border: isSelected ? '1px solid #86efac' : '1px solid #e5e7eb',
                    background: isSelected ? '#f0fdf4' : '#ffffff',
                    cursor: 'pointer',
                    transition: 'all 0.15s ease'
                  }}
                >
                  <input
                    type="checkbox"
                    checked={isSelected}
                    onChange={(e) => {
                      if (e.target.checked) {
                        setPasosSeleccionados((prev) => [...prev, t.id]);
                      } else {
                        setPasosSeleccionados((prev) => prev.filter((id) => id !== t.id));
                      }
                    }}
                    style={{ width: '18px', height: '18px', accentColor: 'var(--pardo-red, #dc2626)', cursor: 'pointer' }}
                  />
                  <div style={{ flex: 1 }}>
                    <div style={{ fontWeight: 600, color: isSelected ? '#15803d' : '#374151', fontSize: '0.95em' }}>
                      {cargoNombre}
                    </div>
                    <div style={{ fontSize: '0.82em', color: '#6b7280' }}>
                      {t.descripcion}
                    </div>
                  </div>
                  <span
                    style={{
                      fontSize: '0.75em',
                      padding: '3px 10px',
                      borderRadius: '12px',
                      background: isSelected ? '#dcfce7' : '#fee2e2',
                      color: isSelected ? '#15803d' : '#991b1b',
                      fontWeight: 600
                    }}
                  >
                    {isSelected ? 'Incluido' : 'Omitido'}
                  </span>
                </label>
              );
            })}

            {/* Paso Final: Cierre (Obligatorio) */}
            <div style={{ display: 'flex', alignItems: 'center', gap: '12px', padding: '10px 14px', borderRadius: '8px', background: '#f3f4f6', border: '1px solid #e5e7eb' }}>
              <i className="fa-solid fa-lock" style={{ color: '#9ca3af', fontSize: '0.9em' }}></i>
              <input type="checkbox" checked disabled style={{ width: '18px', height: '18px' }} />
              <div style={{ flex: 1 }}>
                <div style={{ fontWeight: 600, color: '#374151', fontSize: '0.95em' }}>Revisión Final Auxiliar Admon Oriente</div>
                <div style={{ fontSize: '0.8em', color: '#6b7280' }}>Cierre documental antes de Contabilidad — Obligatorio</div>
              </div>
              <span style={{ fontSize: '0.75em', padding: '3px 8px', borderRadius: '10px', background: '#e5e7eb', color: '#4b5563', fontWeight: 600 }}>Paso Final</span>
            </div>
          </div>

          <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '10px', paddingTop: '10px', borderTop: '1px solid #f3f4f6' }}>
            <button
              type="button"
              className="doc-btn doc-btn-secondary"
              onClick={() => setShowModalDefinirFlujo(false)}
              disabled={completandoTarea}
            >
              Cancelar
            </button>
            <button
              type="button"
              className="doc-btn doc-btn-primary"
              disabled={completandoTarea}
              onClick={async () => {
                setShowModalDefinirFlujo(false);
                await handleCompletarTarea(tareaActiva.id, saiaRadicado.id, pasosSeleccionados);
              }}
            >
              <i className={`fa-solid ${completandoTarea ? 'fa-spinner fa-spin' : 'fa-check-double'}`}></i> Confirmar y Continuar Flujo
            </button>
          </div>
        </div>
      </div>
    )}

      {/* ── Modal de Confirmación de Emisión de Evento DIAN ── */}
      {modalEvento && (
        <div className="evento-confirm-overlay" onClick={() => !emitiendoEvento && setModalEvento(null)}>
          <div className="evento-confirm-modal" onClick={e => e.stopPropagation()}>
            <div className="evento-confirm-header">
              <h3>
                <i className={`fa-solid ${modalEvento.icono}`} style={{ color: 'var(--pardo-navy)' }}></i>
                Emitir Evento DIAN: {modalEvento.label}
              </h3>
              <button
                className="modal-close-btn"
                onClick={() => !emitiendoEvento && setModalEvento(null)}
                disabled={emitiendoEvento}
                style={{ background: 'none', border: 'none', cursor: 'pointer', fontSize: '1.1em', color: '#64748b' }}
              >
                <i className="fa-solid fa-xmark"></i>
              </button>
            </div>

            <div className="evento-confirm-body">
              <div style={{ padding: '10px 14px', borderRadius: '8px', background: '#eff6ff', border: '1px solid #bfdbfe', fontSize: '12px', color: '#1e40af', display: 'flex', alignItems: 'center', gap: '8px' }}>
                <i className="fa-solid fa-circle-info" style={{ fontSize: '14px' }}></i>
                <span>Este evento será transmitido electrónicamente a la DIAN y quedará registrado de forma oficial e irrevocable para la factura electrónica.</span>
              </div>

              {errorEmision && (
                <div style={{ padding: '10px 14px', borderRadius: '8px', background: '#fef2f2', border: '1px solid #fecaca', fontSize: '12px', color: '#b91c1c', display: 'flex', alignItems: 'center', gap: '8px' }}>
                  <i className="fa-solid fa-triangle-exclamation"></i>
                  <span>{errorEmision}</span>
                </div>
              )}

              {modalEvento.codigo === '031' && (
                <div style={{ display: 'flex', flexDirection: 'column', gap: '5px' }}>
                  <label style={{ fontSize: '12px', fontWeight: 600, color: '#334155' }}>Código de Motivo de Reclamo:</label>
                  <select
                    value={formEvento.codigoMotivo || '01'}
                    onChange={e => setFormEvento({ ...formEvento, codigoMotivo: e.target.value })}
                    style={{ padding: '8px 10px', borderRadius: '6px', border: '1px solid #cbd5e1', fontSize: '12.5px' }}
                    disabled={emitiendoEvento}
                  >
                    <option value="01">01 - Documento con inconsistencias</option>
                    <option value="02">02 - Mercancía no entregada totalmente</option>
                    <option value="03">03 - Mercancía no entregada parcialmente</option>
                    <option value="04">04 - Servicio no prestado</option>
                  </select>
                </div>
              )}

              <div style={{ display: 'flex', flexDirection: 'column', gap: '5px' }}>
                <label style={{ fontSize: '12px', fontWeight: 600, color: '#334155' }}>Motivo:</label>
                <input
                  type="text"
                  value={formEvento.motivo}
                  onChange={e => setFormEvento({ ...formEvento, motivo: e.target.value })}
                  placeholder="Motivo del evento"
                  style={{ padding: '8px 10px', borderRadius: '6px', border: '1px solid #cbd5e1', fontSize: '12.5px' }}
                  disabled={emitiendoEvento}
                />
              </div>

              <div style={{ display: 'flex', flexDirection: 'column', gap: '5px' }}>
                <label style={{ fontSize: '12px', fontWeight: 600, color: '#334155' }}>Observación / Comentario:</label>
                <textarea
                  value={formEvento.obs}
                  onChange={e => setFormEvento({ ...formEvento, obs: e.target.value, comentario: e.target.value })}
                  placeholder="Observación sobre este evento..."
                  rows={3}
                  style={{ padding: '8px 10px', borderRadius: '6px', border: '1px solid #cbd5e1', fontSize: '12.5px', resize: 'vertical' }}
                  disabled={emitiendoEvento}
                />
              </div>
            </div>

            <div className="evento-confirm-footer">
              <button
                type="button"
                className="doc-btn doc-btn-secondary"
                onClick={() => setModalEvento(null)}
                disabled={emitiendoEvento}
              >
                Cancelar
              </button>
              <button
                type="button"
                className="doc-btn doc-btn-primary"
                onClick={submitEmitirEvento}
                disabled={emitiendoEvento || !formEvento.motivo?.trim()}
              >
                <i className={`fa-solid ${emitiendoEvento ? 'fa-spinner fa-spin' : 'fa-paper-plane'}`}></i>
                {emitiendoEvento ? 'Transmitiendo a DIAN...' : 'Confirmar y Emitir'}
              </button>
            </div>
          </div>
        </div>
      )}
    </>
  );
}