import {  } from "../helpers/formatters";

export default function RenderTareas({
  misTareas, misTareasCompletadas, selectedTareaId, loadingTareas,
  tareasSubTab, setTareasSubTab, searchTareas, setSearchTareas,
  sortTareas, setSortTareas, getFilteredTareas, openSaia
}) {
  const tareasMostradas = getFilteredTareas(tareasSubTab === "activas" ? misTareas : misTareasCompletadas);
  return (
    <div className="fullwidth-list-container">
      <div className="fullwidth-list-header">
        <h3><i className="fa-solid fa-clipboard-list"></i> Mis Tareas</h3>
        <div className="tareas-subtabs">
          <button className={tareasSubTab === "activas" ? "active" : ""} onClick={() => { setTareasSubTab("activas"); }}>
            <i className="fa-solid fa-spinner"></i> En Proceso ({misTareas.length})
          </button>
          <button className={tareasSubTab === "completadas" ? "active" : ""} onClick={() => { setTareasSubTab("completadas"); }}>
            <i className="fa-solid fa-check-circle"></i> Completadas ({misTareasCompletadas.length})
          </button>
        </div>
        <div className="list-controls">
          <input type="text" placeholder="Buscar radicado, doc, proveedor..." value={searchTareas} onChange={e => setSearchTareas(e.target.value)} />
          <select value={sortTareas} onChange={e => setSortTareas(e.target.value)}>
            <option value="fecha_desc">Más recientes</option>
            <option value="fecha_asc">Más antiguos</option>
            <option value="estado">Por Estado</option>
          </select>
        </div>
      </div>
      <div className="fullwidth-list-body">
        {loadingTareas ? <p style={{ padding: "15px" }}>Cargando...</p> : tareasMostradas.length === 0 ? (
          <div className="correo-empty"><i className="fa-solid fa-inbox"></i><p>{tareasSubTab === "activas" ? "No tienes documentos asignados." : "No hay documentos finalizados."}</p></div>
        ) : (
          <div className="card-grid">
            {tareasMostradas.map((rad) => (
              <div key={rad.id} className={`rad-card ${selectedTareaId === rad.id ? "active" : ""}`} onClick={() => openSaia(rad, "tareas")}>
                <div className="rad-card-top">
                  <span className="rad-card-tipo">{rad.documento_comercial?.tipo || "—"}</span>
                  <span className={`status-badge ${rad.estado_posesion === "Completado" ? "radicado" : rad.estado_posesion === "EnProceso" ? "doc-pendiente" : (rad.estado_posesion === "Devuelto" || rad.estado_posesion === "Rechazado") ? "doc-rechazado" : ""}`}>
                    {rad.estado_posesion === "Completado" ? "Finalizado" : rad.estado_posesion === "EnProceso" ? "En Proceso" : (rad.estado_posesion === "Devuelto" || rad.estado_posesion === "Rechazado") ? "Rechazado" : "Sin estado"}
                  </span>
                </div>
                <div className="rad-card-numero"><i className="fa-solid fa-hashtag"></i> {rad.numero_radicado}</div>
                <div className="rad-card-doc">{rad.documento_comercial?.numero_documento || "Sin documento"}</div>
                <div className="rad-card-bottom">
                  <span className="rad-card-date"><i className="fa-regular fa-calendar"></i> {new Date(rad.fecha_radicacion).toLocaleDateString()}</span>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}