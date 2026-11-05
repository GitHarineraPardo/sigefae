export default function RenderFlujoAprobacion({ tareasFlujo, esPrimerTurnoAuxOriente }) {
  return (
    <div className="doc-section">
      <h4><i className="fa-solid fa-route"></i> Flujo de Aprobación</h4>
      {esPrimerTurnoAuxOriente && (
        <div style={{ padding: '8px 12px', background: '#eff6ff', border: '1px solid #bfdbfe', borderRadius: '6px', marginBottom: '12px', fontSize: '0.85em', color: '#1e40af', display: 'flex', alignItems: 'center', gap: '8px' }}>
          <i className="fa-solid fa-circle-info"></i>
          <span><strong>Turno de Definición:</strong> Como Auxiliar Admon Oriente, al hacer clic en <em>"Definir Flujo y Continuar"</em> podrás marcar o desmarcar qué pasos formarán parte del flujo oficial.</span>
        </div>
      )}
      {tareasFlujo.length === 0 ? (
        <p style={{ color: "#6b7280", fontSize: "0.9em" }}>No hay pasos definidos para este flujo.</p>
      ) : (
        <table className="doc-items-table">
          <thead><tr><th>Cargo</th><th>Acción a realizar</th><th>Estado</th><th>Fecha Realización</th></tr></thead>
          <tbody>
            {tareasFlujo.map((t) => (
              <tr key={t.id} style={t.estado?.nombre === "En Proceso" ? { background: "rgba(44,82,130,0.05)" } : {}}>
                <td>{typeof t.cargo_asignado === "string" ? t.cargo_asignado || "—" : t.cargo_asignado?.nombre || "—"}</td>
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