export default function RenderFlujoAprobacion({ tareasFlujo }) {
  return (
    <div className="doc-section">
      <h4><i className="fa-solid fa-route"></i> Flujo de Aprobación</h4>
      {tareasFlujo.length === 0 ? (
        <p style={{ color: "#6b7280", fontSize: "0.9em" }}>No hay pasos definidos para este flujo.</p>
      ) : (
        <table className="doc-items-table">
          <thead><tr><th>Usuario</th><th>Correo</th><th>Acción a realizar</th><th>Estado</th><th>Fecha Realización</th></tr></thead>
          <tbody>
            {tareasFlujo.map((t) => (
              <tr key={t.id} style={t.estado?.nombre === "En Proceso" ? { background: "rgba(44,82,130,0.05)" } : {}}>
                <td>{t.usuario_asignado?.nombre || (t.cargo_asignado?.nombre ? `Cargo: ${t.cargo_asignado.nombre}` : typeof t.cargo_asignado === "string" && t.cargo_asignado ? `Cargo: ${t.cargo_asignado}` : "—")}</td>
                <td>{t.usuario_asignado?.email || "—"}</td>
                <td>{t.descripcion}</td>
                <td><span className={`status-badge ${t.estado?.nombre === "Completada" ? "radicado" : t.estado?.nombre === "En Proceso" ? "doc-pendiente" : t.estado?.nombre === "Devuelta" ? "doc-rechazado" : ""}`}>{t.estado?.nombre || "Pendiente"}</span></td>
                <td>{t.fecha_finalizacion ? new Date(t.fecha_finalizacion).toLocaleString() : "—"}</td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </div>
  );
}