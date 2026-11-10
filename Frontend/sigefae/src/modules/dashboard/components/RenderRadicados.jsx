import { useState } from "react";
import { isFinalState } from "../helpers/formatters";

export default function RenderRadicados({
  radicados, selectedRadicadoId, loadingRadicados, searchRadicados, setSearchRadicados,
  sortRadicados, setSortRadicados, filtroZona, setFiltroZona, tareasPorRadicado, getFilteredRadicados, openSaia
}) {
  const [activeSubTab, setActiveSubTab] = useState("en_proceso");
  const [fechaInicio, setFechaInicio] = useState("");
  const [fechaFin, setFechaFin] = useState("");
  const [viewMode, setViewMode] = useState("cards");

  const baseFilteredList = getFilteredRadicados();
  const zonasDisponibles = [...new Set(radicados.map(r => r.ruta?.zona).filter(Boolean))].sort();
  const getFechaCalendario = (fecha) => {
    if (!fecha) return "";
    const date = new Date(fecha);
    if (Number.isNaN(date.getTime())) return "";
    return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, "0")}-${String(date.getDate()).padStart(2, "0")}`;
  };
  const getFechaPrincipal = (radicado) => radicado.documento_comercial?.fecha_recepcion_dian || radicado.fecha_radicacion;
  
  const filteredList = baseFilteredList.filter(rad => {
    if (activeSubTab === "finalizados" && !isFinalState(rad.estado_posesion)) return false;
    if (activeSubTab !== "finalizados" && isFinalState(rad.estado_posesion)) return false;
    const fecha = getFechaCalendario(getFechaPrincipal(rad));
    if (fechaInicio && (!fecha || fecha < fechaInicio)) return false;
    if (fechaFin && (!fecha || fecha > fechaFin)) return false;
    return true;
  });
  const formatearFecha = (fecha) => fecha
    ? new Date(fecha).toLocaleDateString("es-CO", { day: "2-digit", month: "short", year: "numeric" })
    : "—";
  const getEstadoLabel = (estado) => estado === "Completado" ? "Finalizado"
    : estado === "EnProceso" ? "En Proceso"
      : (estado === "Devuelto" || estado === "Rechazado") ? "Rechazado"
        : estado === "Libre" ? "Libre" : "En espera";
  const getStatusClass = (estado) => estado === "Completado" ? "radicado"
    : estado === "EnProceso" ? "doc-pendiente"
      : (estado === "Devuelto" || estado === "Rechazado") ? "doc-rechazado" : "";
  const limpiarFechas = () => {
    setFechaInicio("");
    setFechaFin("");
  };
  return (
    <div className="fullwidth-list-container">
      <div className="fullwidth-list-header" style={{ flexDirection: "column", alignItems: "stretch", gap: "15px" }}>
        <div className="radicados-toolbar">
          <h3><i className="fa-solid fa-stamp"></i> Documentos Radicados ({filteredList.length})</h3>
        <div className="list-controls radicados-controls">
            <input type="text" placeholder="Buscar radicado, doc, proveedor..." value={searchRadicados} onChange={e => setSearchRadicados(e.target.value)} />
            <select value={filtroZona || ""} onChange={e => setFiltroZona(e.target.value)}>
              <option value="">Todas las zonas</option>
              {zonasDisponibles.map(z => (
                <option key={z} value={z}>{z}</option>
              ))}
            </select>
            <select value={sortRadicados} onChange={e => setSortRadicados(e.target.value)}>
              <option value="fecha_desc">Más recientes</option>
              <option value="fecha_asc">Más antiguos</option>
              <option value="estado">Por Estado</option>
            </select>
          </div>
          </div>
        <div className="radicados-subtabs-row">
          <div className="radicados-subtabs">
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
          <div className="date-filter">
            <label htmlFor="radicados-fecha-inicio"><i className="fa-regular fa-calendar"></i> Desde</label>
            <input id="radicados-fecha-inicio" type="date" value={fechaInicio} max={fechaFin || undefined} onChange={e => setFechaInicio(e.target.value)} />
          </div>
          <div className="date-filter">
            <label htmlFor="radicados-fecha-fin"><i className="fa-regular fa-calendar"></i> Hasta</label>
            <input id="radicados-fecha-fin" type="date" min={fechaInicio || undefined} onChange={e => setFechaFin(e.target.value)} value={fechaFin} />
          </div>
          {(fechaInicio || fechaFin) && <button className="clear-date-filter" type="button" onClick={limpiarFechas}><i className="fa-solid fa-rotate-left"></i> Limpiar fechas</button>}
          <div className="view-switcher" aria-label="Modo de visualización">
            <span>Vista</span>
            <button type="button" className={viewMode === "cards" ? "active" : ""} onClick={() => setViewMode("cards")} title="Vista de tarjetas"><i className="fa-solid fa-grip"></i><span>Tarjetas</span></button>
            <button type="button" className={viewMode === "table" ? "active" : ""} onClick={() => setViewMode("table")} title="Vista tabulada"><i className="fa-solid fa-table-list"></i><span>Tabla</span></button>
          </div>
        </div>
      </div>

      <div className="fullwidth-list-body">
        {loadingRadicados ? <p style={{ padding: "15px" }}>Cargando...</p> : filteredList.length === 0 ? (
          <div className="correo-empty"><i className="fa-solid fa-inbox"></i><p>No hay documentos radicados.</p></div>
        ) : (
          viewMode === "cards" ? <div className="card-grid">
            {filteredList.map((rad) => (
              <div key={rad.id} className={`rad-card ${selectedRadicadoId === rad.id ? "active" : ""}`} onClick={() => openSaia(rad, "radicados")}>
                <div className="rad-card-top">
                  <span className="rad-card-tipo">{rad.documento_comercial?.tipo || "—"}</span>
                  <span className={`status-badge ${rad.estado_posesion === "Completado" ? "radicado" : rad.estado_posesion === "EnProceso" ? "doc-pendiente" : (rad.estado_posesion === "Devuelto" || rad.estado_posesion === "Rechazado") ? "doc-rechazado" : ""}`}>
                    {getEstadoLabel(rad.estado_posesion)}
                  </span>
                </div>
                <div className="rad-card-numero"><i className="fa-solid fa-hashtag"></i> {rad.numero_radicado}</div>
                <div className="rad-card-doc">{rad.documento_comercial?.numero_documento || "Sin documento"}</div>
                <div className="rad-card-doc" title={rad.documento_comercial?.proveedor?.razon_social || "Proveedor no disponible"}>
                  <i className="fa-solid fa-building"></i> {rad.documento_comercial?.proveedor?.razon_social || "Proveedor no disponible"}
                </div>
                <div className="rad-card-doc">
                  <i className="fa-solid fa-briefcase"></i> Cargo actual: {rad.cargo_actual?.nombre || "Sin cargo asignado"}
                </div>
                <div className="rad-card-bottom">
                  <span className="rad-card-date"><i className="fa-regular fa-calendar"></i> Recibido DIAN: {formatearFecha(getFechaPrincipal(rad))}</span>
                  <span className="rad-card-date rad-card-date-secondary">Radicado: {formatearFecha(rad.fecha_radicacion)}</span>
                  {(!isFinalState(rad.estado_posesion)) && rad.usuario_actual?.nombre && (
                    <span className="rad-card-user" title={`${rad.usuario_actual.nombre} — ${tareasPorRadicado[rad.id]?.descripcion || rad.paso_actual?.nombre || "Sin paso"}`}>
                      <i className="fa-solid fa-user"></i> {rad.usuario_actual.nombre}
                    </span>
                  )}
                </div>
              </div>
            ))}
          </div> : <div className="radicados-table-wrap">
            <table className="radicados-table">
              <thead><tr>
                <th>Radicado</th><th>Documento</th><th>Proveedor</th><th>Tipo</th>
                <th>Fecha</th><th>Zona</th><th>Cargo actual</th><th>Estado</th>
              </tr></thead>
              <tbody>{filteredList.map((rad) => (
                <tr key={rad.id} className={selectedRadicadoId === rad.id ? "active" : ""} onClick={() => openSaia(rad, "radicados")}>
                  <td className="table-main-cell"><strong>{rad.numero_radicado || "—"}</strong><small>{rad.cargo_actual?.nombre || rad.cargo_actual_nombre || "Sin cargo asignado"}</small></td>
                  <td>{rad.documento_comercial?.numero_documento || "Sin documento"}</td>
                  <td className="table-truncate" title={rad.documento_comercial?.proveedor?.razon_social}>{rad.documento_comercial?.proveedor?.razon_social || "Sin proveedor"}</td>
                  <td>{rad.documento_comercial?.tipo || "—"}</td>
                  <td className="table-date"><strong>{formatearFecha(getFechaPrincipal(rad))}</strong><small>Radicado: {formatearFecha(rad.fecha_radicacion)}</small></td>
                  <td>{rad.ruta?.zona || "—"}</td>
                  <td className="table-truncate">{rad.cargo_actual?.nombre || "Sin cargo"}</td>
                  <td><span className={`status-badge ${getStatusClass(rad.estado_posesion)}`}>{getEstadoLabel(rad.estado_posesion)}</span></td>
                </tr>
              ))}</tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  );
}