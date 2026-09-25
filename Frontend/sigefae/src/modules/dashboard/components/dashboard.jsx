import { useState, useEffect } from "react";
import "../dashboard.css";
import "@fortawesome/fontawesome-free/css/all.min.css";
import logo from "/logo.png";
import { obtenerToken } from "../../auth/token.js";
import PdfEditor from "../../../components/PdfEditor.jsx";
import NotificacionesDropdown from "../../../components/NotificacionesDropdown.jsx";
import AdminToast from "../../../components/AdminToast.jsx";
import { API } from "../constants/api.js";
import { isFinalState } from "../helpers/formatters.js";
import { useTrazabilidadPorArea } from "../hooks/useTrazabilidadPorArea.js";
import RenderTrazabilidadPorArea from "./RenderTrazabilidadPorArea.jsx";

// Hooks
import { useAuth } from "../hooks/useAuth.js";
import { useCorreos } from "../hooks/useCorreos.js";
import { useDocumentos } from "../hooks/useDocumentos.js";
import { useRadicados } from "../hooks/useRadicados.js";
import { useTareas } from "../hooks/useTareas.js";
import { useCatalogos } from "../hooks/useCatalogos.js";
import { useSolicitudes } from "../hooks/useSolicitudes.js";
import { useFlujoYTrazabilidad } from "../hooks/useFlujoYTrazabilidad.js";
import { useSaia } from "../hooks/useSaia.js";
import { useAnexos } from "../hooks/useAnexos.js";
import { useDevolucion } from "../hooks/useDevolucion.js";
import { useNormasReparto } from "../hooks/useNormasReparto.js";
import { useRadicacion } from "../hooks/useRadicacion.js";
import { useCompletarTarea } from "../hooks/useCompletarTarea.js";
import { useAccionesAdmin } from "../hooks/useAccionesAdmin.js";
import { usePdfExpediente } from "../hooks/usePdfExpediente.js";
import { useFinanzas } from "../hooks/useFinanzas.js";
import { useNotificaciones } from "../../../hooks/useNotificaciones.js";

// Components
import Sidebar from "./Sidebar.jsx";
import RenderWelcome from "./RenderWelcome.jsx";
import RenderCorreos from "./RenderCorreos.jsx";
import RenderDocumentos from "./RenderDocumentos.jsx";
import RenderRadicados from "./RenderRadicados.jsx";
import RenderTareas from "./RenderTarea.jsx";
import RenderCatalogos from "./RenderCatalogos.jsx";
import RenderSolicitudes from "./RenderSolicitudes.jsx";
import RenderSaiaModal from "./modals/RenderSaiaModal.jsx";
import ModalRadicar from "./modals/ModalRadicar.jsx";
import ModalCrearDocumento from "./modals/ModalCrearDocumento.jsx";
import ModalCrearProveedor from "./modals/ModalCrearProveedor.jsx";
import ModalDevolver from "./modals/ModalDevolver.jsx";
import ModalNormaReparto from "./modals/ModalNormaReparto.jsx";
import ModalActivoFijo from "./modals/ModalActivoFijo.jsx";
import RenderFinanzas from "./RenderFinanzas.jsx";

export default function ProcesosLogistica() {
  const [isSidebarOpen, setIsSidebarOpen] = useState(true);
  const { userRol, userId, esAdmin, esUsuario, showDebug, puedeGestionarRecurso, obtenerToken } = useAuth();
  
  const initialTab = esAdmin ? "welcome" : (userRol === "Contabilidad" || userRol === "Tesorería") ? "finanzas" : "tareas";
  const [activeTab, setActiveTab] = useState(initialTab);

  const [fontSize, setFontSize] = useState(() => localStorage.getItem("sigefae_font_size") || "16px");

  useEffect(() => {
    document.documentElement.style.fontSize = fontSize;
    localStorage.setItem("sigefae_font_size", fontSize);
  }, [fontSize]);

  // PDF Editor
  const [pdfEditor, setPdfEditor] = useState({ open: false, archivoId: null, archivoNombre: null, radicadoId: null });

  // Hooks de datos
  const correosHook = useCorreos(obtenerToken, activeTab);
  const documentosHook = useDocumentos(obtenerToken, activeTab);
  const radicadosHook = useRadicados(obtenerToken, activeTab, userId);
  const tareasHook = useTareas(obtenerToken, activeTab, userId);
  const catalogosHook = useCatalogos(obtenerToken, activeTab);
  const solicitudesHook = useSolicitudes(obtenerToken, activeTab, esAdmin);

  // Flujo, trazabilidad, comentarios, normas (compartido entre radicados/tareas/SAIA)
  const flujoHook = useFlujoYTrazabilidad(obtenerToken, radicadosHook.selectedRadicadoId, tareasHook.selectedTareaId);

  // SAIA
  const saiaHook = useSaia(obtenerToken);

  // Anexos
  const anexosHook = useAnexos(obtenerToken, puedeGestionarRecurso, esAdmin, activeTab, tareasHook.setSelectedTareaId, radicadosHook.setSelectedRadicadoId, saiaHook.saiaModalOpen, saiaHook.saiaRadicado, saiaHook.setSaiaRadicado, saiaHook.saiaAnexoIdx, saiaHook.setSaiaAnexoIdx, saiaHook.saiaPdfUrl, saiaHook.setSaiaPdfUrl);

  // Devolución
  const devolucionHook = useDevolucion(obtenerToken, activeTab, tareasHook.setSelectedTareaId, radicadosHook.setSelectedRadicadoId, flujoHook.setTareasFlujo, flujoHook.setHistorialTrazabilidad, flujoHook.setComentarios, tareasHook.setMisTareas, tareasHook.setMisTareasCompletadas, radicadosHook.setRadicados, userId);

  // Normas reparto
  const normasHook = useNormasReparto(obtenerToken, esAdmin, puedeGestionarRecurso, flujoHook.normasRepartoRadicado, flujoHook.setNormasRepartoRadicado, flujoHook.activosFijosRadicado, flujoHook.setActivosFijosRadicado);

  // Radicación
  const radicacionHook = useRadicacion(obtenerToken, userId, documentosHook.setDocumentos, setActiveTab, documentosHook.setSelectedDocId, documentosHook.setDocDetail);

  // Completar tarea
  const completarHook = useCompletarTarea(obtenerToken, userId, activeTab, flujoHook.setTareasFlujo, tareasHook.setTareaDetail, radicadosHook.setRadicadoDetail, tareasHook.setMisTareas, tareasHook.setMisTareasCompletadas, radicadosHook.setRadicados, saiaHook.setSaiaRadicado);

  // Acciones admin
  const accionesAdminHook = useAccionesAdmin(obtenerToken, radicadosHook.setRadicadoDetail, saiaHook.setSaiaRadicado);

  // PDF Expediente
  const pdfExpedienteHook = usePdfExpediente();

  // Finanzas
  const finanzasHook = useFinanzas(obtenerToken, radicadosHook);

  // Notificaciones (única instancia — se pasan como props al dropdown)
  const notifHook = useNotificaciones();

  const trazabilidadAreaHook = useTrazabilidadPorArea(obtenerToken);

  // Colapsar secciones
  useEffect(() => {
    const handler = (e) => {
      const h = e.target.closest && e.target.closest('.doc-section > h4');
      if (!h) return;
      if (h.closest('.saia-comments-wrapper')) return;
      const parent = h.parentElement;
      if (!parent) return;
      parent.classList.toggle('collapsed');
    };
    document.addEventListener('click', handler);
    return () => document.removeEventListener('click', handler);
  }, []);

  // Colapsar por defecto al abrir detalle
  useEffect(() => {
    if (radicadosHook.radicadoDetail || tareasHook.tareaDetail) {
      setTimeout(() => {
        const parent = document.querySelector('.doc-detail-content');
        if (!parent) return;
        parent.querySelectorAll('.doc-section').forEach((el) => el.classList.add('collapsed'));
      }, 30);
    }
  }, [radicadosHook.radicadoDetail, tareasHook.tareaDetail]);

  const getTabInfo = () => {
    switch (activeTab) {
      case "correos": return { icon: "fa-solid fa-envelope", title: "Recepción de Correos", subtitle: "Gestiona las facturas electrónicas recibidas" };
      case "documentos": return { icon: "fa-solid fa-file-invoice", title: "Documentos Pendientes", subtitle: "Revisa y aprueba documentos para radicación" };
      case "radicados": return { icon: "fa-solid fa-stamp", title: "Documentos Radicados", subtitle: "Consulta el estado de los documentos radicados" };
      case "catalogos": return { icon: "fa-solid fa-sliders", title: "Catálogos del Sistema", subtitle: "Gestiona tipos de radicación, pagos y métodos" };
      case "tareas": return { icon: "fa-solid fa-clipboard-list", title: "Mis Tareas", subtitle: "Documentos radicados asignados a ti" };
      case "solicitudes": return { icon: "fa-solid fa-circle-exclamation", title: "Gestión de Solicitudes", subtitle: esAdmin ? "Solicitudes pendientes" : "Tus solicitudes realizadas" };
      case "finanzas": return { icon: "fa-solid fa-file-invoice-dollar", title: "Control Financiero", subtitle: "Revisión de causación y egresos" };
      case "trazabilidad-area": return { icon: "fa-solid fa-route", title: "Trazabilidad por Área", subtitle: "Consulta y descarga trazabilidad filtrada por área y fechas" };
      default: return { icon: "fa-solid fa-house", title: "Procesos administrativos", subtitle: "Selecciona un formato del menú lateral" };
    }
  };

  const tabInfo = getTabInfo();

  const renderContent = () => {
    switch (activeTab) {
      case "correos": return <RenderCorreos {...correosHook} />;
      case "documentos": return <RenderDocumentos {...documentosHook} esAdmin={esAdmin} openRadicarModal={radicacionHook.openRadicarModal} />;
      case "radicados": return <RenderRadicados {...radicadosHook} openSaia={(rad, fromTab) => saiaHook.openSaia(rad, fromTab, tareasHook.setSelectedTareaId, radicadosHook.setSelectedRadicadoId, flujoHook.recargarFlujo, flujoHook.recargarTrazabilidad, flujoHook.recargarNormas, flujoHook.recargarComentarios)} />;
      case "tareas": return <RenderTareas {...tareasHook} openSaia={(rad, fromTab) => saiaHook.openSaia(rad, fromTab, tareasHook.setSelectedTareaId, radicadosHook.setSelectedRadicadoId, flujoHook.recargarFlujo, flujoHook.recargarTrazabilidad, flujoHook.recargarNormas, flujoHook.recargarComentarios)} />;
      case "catalogos": return <RenderCatalogos {...catalogosHook} />;
      case "solicitudes": return <RenderSolicitudes {...solicitudesHook} esAdmin={esAdmin} />;
      case "finanzas": return <RenderFinanzas 
        radicados={radicadosHook.radicados} 
        loadingRadicados={radicadosHook.loadingRadicados} 
        {...finanzasHook} 
        userRole={userRol} 
        handleSubirAnexo={anexosHook.handleSubirAnexo}
        obtenerToken={obtenerToken}
        openSaia={(rad, fromTab) => saiaHook.openSaia(rad, fromTab, tareasHook.setSelectedTareaId, radicadosHook.setSelectedRadicadoId, flujoHook.recargarFlujo, flujoHook.recargarTrazabilidad, flujoHook.recargarNormas, flujoHook.recargarComentarios)} 
      />;
      case "trazabilidad-area": return <RenderTrazabilidadPorArea {...trazabilidadAreaHook} obtenerToken={obtenerToken} />;
      default: return <RenderWelcome />;
    }
  };

  const detalleRadicadoActual = activeTab === "tareas" ? tareasHook.tareaDetail : radicadosHook.radicadoDetail;
  const editorPermitido = detalleRadicadoActual?.estado_posesion && !isFinalState(detalleRadicadoActual.estado_posesion);

  return (
    <>
      <header className="top-header">
        <div className="header-left">
          <button className="menu-toggle" title="Menú" onClick={() => setIsSidebarOpen(!isSidebarOpen)}><i className="fas fa-bars" /></button>
          <img src={logo} alt="Logo" className="logo" />
          <div className="header-text"><h1>SIGEFAE</h1><p>Sistema de Gestion de Facturas Electronicas</p></div>
        </div>
        <div className="header-right">
          <a
            href="https://app.harinerapardo.co/index.php"
            className="btn-volver-panel"
            title="Volver al Panel Principal"
            style={{
              display: "inline-flex",
              alignItems: "center",
              gap: "6px",
              padding: "8px 14px",
              backgroundColor: "rgba(255, 255, 255, 0.15)",
              color: "#ffffff",
              borderRadius: "6px",
              textDecoration: "none",
              fontWeight: "600",
              fontSize: "13px",
              marginRight: "12px",
              border: "1px solid rgba(255, 255, 255, 0.25)",
              transition: "all 0.2s"
            }}
          >
            <i className="fa-solid fa-arrow-left"></i> Volver al Panel
          </a>
          <div className="font-size-selector" title="Tamaño de Letra" style={{ display: "inline-flex", alignItems: "center", gap: "2px", background: "rgba(255, 255, 255, 0.15)", borderRadius: "6px", padding: "2px", marginRight: "10px", border: "1px solid rgba(255, 255, 255, 0.25)" }}>
            <button title="Fuente Pequeña" onClick={() => setFontSize("14px")} style={{ background: fontSize === "14px" ? "rgba(255,255,255,0.35)" : "transparent", border: "none", color: "white", padding: "3px 7px", borderRadius: "4px", cursor: "pointer", fontSize: "11px", fontWeight: "bold" }}>A-</button>
            <button title="Fuente Normal" onClick={() => setFontSize("16px")} style={{ background: fontSize === "16px" ? "rgba(255,255,255,0.35)" : "transparent", border: "none", color: "white", padding: "3px 7px", borderRadius: "4px", cursor: "pointer", fontSize: "13px", fontWeight: "bold" }}>A</button>
            <button title="Fuente Grande" onClick={() => setFontSize("18px")} style={{ background: fontSize === "18px" ? "rgba(255,255,255,0.35)" : "transparent", border: "none", color: "white", padding: "3px 7px", borderRadius: "4px", cursor: "pointer", fontSize: "15px", fontWeight: "bold" }}>A+</button>
          </div>
          <NotificacionesDropdown
            notificaciones={notifHook.notificaciones}
            noLeidas={notifHook.noLeidas}
            marcarLeida={notifHook.marcarLeida}
            borrarNotificacion={notifHook.borrarNotificacion}
            recargar={notifHook.recargar}
            onNavigate={(id) => { if (esAdmin) { setActiveTab("radicados"); radicadosHook.setSelectedRadicadoId(id); tareasHook.setSelectedTareaId(null); } else { setActiveTab("tareas"); tareasHook.setSelectedTareaId(id); radicadosHook.setSelectedRadicadoId(null); } }}
          />
          <AdminToast esAdmin={esAdmin} />
        </div>
      </header>

      <div className="main-container">
        <Sidebar isSidebarOpen={isSidebarOpen} setIsSidebarOpen={setIsSidebarOpen} activeTab={activeTab} setActiveTab={setActiveTab} esAdmin={esAdmin} esUsuario={esUsuario} userRole={userRol} />
        <main className="content-area">
          <div className="content-header">
            <div className="content-title">
              <div className="content-icon"><i className={tabInfo.icon}></i></div>
              <div><h2>{tabInfo.title}</h2><p>{tabInfo.subtitle}</p></div>
            </div>
          </div>
          {renderContent()}
        </main>
      </div>

      {/* Modales */}
      <ModalRadicar {...radicacionHook} />
      <ModalCrearDocumento {...documentosHook} openCrearProveedorModal={documentosHook.openCrearProveedorModal} />
      <ModalCrearProveedor showCrearProveedorModal={documentosHook.showCrearProveedorModal} setShowCrearProveedorModal={documentosHook.setShowCrearProveedorModal} proveedorForm={documentosHook.proveedorForm} creandoProveedor={documentosHook.creandoProveedor} tiposDocumentoCatalogo={documentosHook.tiposDocumentoCatalogo} handleProveedorFormChange={documentosHook.handleProveedorFormChange} handleCrearProveedorSubmit={documentosHook.handleCrearProveedorSubmit} />
      <ModalDevolver {...devolucionHook} tareasFlujo={flujoHook.tareasFlujo} selectedRadicadoId={radicadosHook.selectedRadicadoId} selectedTareaId={tareasHook.selectedTareaId} />
      <ModalNormaReparto {...normasHook} subtotalRadicado={detalleRadicadoActual?.documento_comercial?.subtotal || 0} />
      <ModalActivoFijo {...normasHook} subtotalRadicado={detalleRadicadoActual?.documento_comercial?.subtotal || 0} />

      {/* SAIA */}
      <RenderSaiaModal
        saiaModalOpen={saiaHook.saiaModalOpen}
        setSaiaModalOpen={saiaHook.setSaiaModalOpen}
        saiaRadicado={saiaHook.saiaRadicado}
        setSaiaRadicado={saiaHook.setSaiaRadicado}
        saiaActiveTab={saiaHook.saiaActiveTab}
        setSaiaActiveTab={saiaHook.setSaiaActiveTab}
        saiaAnexoIdx={saiaHook.saiaAnexoIdx}
        setSaiaAnexoIdx={saiaHook.setSaiaAnexoIdx}
        saiaPdfUrl={saiaHook.saiaPdfUrl}
        tareasFlujo={flujoHook.tareasFlujo}
        historialTrazabilidad={flujoHook.historialTrazabilidad}
        normasRepartoRadicado={flujoHook.normasRepartoRadicado}
        comentarios={flujoHook.comentarios}
        nuevoComentario={flujoHook.nuevoComentario}
        setNuevoComentario={flujoHook.setNuevoComentario}
        enviandoComentario={flujoHook.enviandoComentario}
        handleEnviarComentario={flujoHook.handleEnviarComentario}
        generandoPdf={pdfExpedienteHook.generandoPdf}
        handleDescargarExpediente={(rad) => pdfExpedienteHook.handleDescargarExpediente(rad, flujoHook.tareasFlujo, flujoHook.historialTrazabilidad, API)}
        handleVerAnexo={anexosHook.handleVerAnexo}
        handleDescargarAnexo={anexosHook.handleDescargarAnexo}
        handleBorrarAnexo={anexosHook.handleBorrarAnexo}
        handleSubirAnexo={anexosHook.handleSubirAnexo}
        completandoTarea={completarHook.completandoTarea}
        handleCompletarTarea={completarHook.handleCompletarTarea}
        solicitarRechazo={accionesAdminHook.solicitarRechazo}
        marcarCompletado={accionesAdminHook.marcarCompletado}
        adminRechazar={accionesAdminHook.adminRechazar}
        esAdmin={esAdmin}
        esUsuario={esUsuario}
        userId={userId}
        userRol={userRol}
        setPdfEditor={setPdfEditor}
        activeTab={activeTab}
        setSelectedTareaId={tareasHook.setSelectedTareaId}
        setSelectedRadicadoId={radicadosHook.setSelectedRadicadoId}
        setShowDevolverModal={devolucionHook.setShowDevolverModal}
        setDevolverForm={devolucionHook.setDevolverForm}
        openNormaModal={normasHook.openNormaModal}
        handleEliminarNorma={normasHook.handleEliminarNorma}
        activosFijosRadicado={flujoHook.activosFijosRadicado}
        openActivoModal={normasHook.openActivoModal}
        handleEliminarActivoFijo={normasHook.handleEliminarActivoFijo}
        readOnly={activeTab === "finanzas"}
      />

      {/* PDF Editor */}
      {pdfEditor.open && editorPermitido && (
        <PdfEditor
          archivoId={pdfEditor.archivoId}
          archivoNombre={pdfEditor.archivoNombre}
          radicadoId={pdfEditor.radicadoId}
          onClose={() => setPdfEditor({ open: false, archivoId: null, archivoNombre: null, radicadoId: null })}
          onSaved={() => {
            const radicadoId = pdfEditor.radicadoId;
            setPdfEditor({ open: false, archivoId: null, archivoNombre: null, radicadoId: null });
            if (activeTab === "tareas") { tareasHook.setSelectedTareaId(null); setTimeout(() => tareasHook.setSelectedTareaId(radicadoId), 10); }
            else { radicadosHook.setSelectedRadicadoId(null); setTimeout(() => radicadosHook.setSelectedRadicadoId(radicadoId), 10); }
            
            if (saiaHook.saiaModalOpen && saiaHook.saiaRadicado) {
              saiaHook.openSaia(saiaHook.saiaRadicado, activeTab, tareasHook.setSelectedTareaId, radicadosHook.setSelectedRadicadoId, flujoHook.recargarFlujo, flujoHook.recargarTrazabilidad, flujoHook.recargarNormas, flujoHook.recargarComentarios);
            }
          }}
        />
      )}
    </>
  );
}