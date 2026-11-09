import TimelineEventosDIAN from "./TimelineEventosDIAN.jsx";

export default function RenderTrazabilidad({ historialTrazabilidad = [], eventosDian = [], cargandoEventos = false }) {
  return (
    <div className="doc-section" style={{ display: "flex", flexDirection: "column", flex: 1, minHeight: 0 }}>
      {/* ── Eventos DIAN (RADIAN) ── */}
      <TimelineEventosDIAN eventosDian={eventosDian} cargando={cargandoEventos} />

      <h4>
        <i className="fa-solid fa-clock-rotate-left"></i>{" "}
        Historial de Trazabilidad Interna ({historialTrazabilidad.length})
      </h4>
      <div
        style={{
          flex: 1,
          minHeight: 250,
          maxHeight: "550px",
          overflowY: "auto",
          marginBottom: 12,
          paddingRight: 4
        }}
      >
        {historialTrazabilidad.length === 0 ? (
          <p style={{ color: "#6b7280", fontSize: "0.9em" }}>
            No hay historial disponible.
          </p>
        ) : (
          <div style={{ display: "flex", flexDirection: "column", gap: 8 }}>
            {historialTrazabilidad.map((t, idx) => (
              <div
                key={idx}
                style={{
                  padding: "10px",
                  background: "#f9fafb",
                  borderLeft: "3px solid var(--pardo-red)",
                  borderRadius: 4
                }}
              >
                <div style={{ display: "flex", justifyContent: "space-between", marginBottom: 4 }}>
                  <strong style={{ color: "var(--pardo-navy)", fontSize: "0.9em" }}>{t.accion}</strong>
                  <span style={{ fontSize: "0.8em", color: "#6b7280" }}>
                    {new Date(t.fecha).toLocaleString()}
                  </span>
                </div>
                <div style={{ fontSize: "0.85em", color: "#4b5563", marginBottom: 4 }}>
                  {t.descripcion}
                </div>
                <div style={{ fontSize: "0.8em", color: "#9ca3af", textAlign: "right" }}>
                  <i className="fa-solid fa-user" style={{ marginRight: 4 }}></i>
                  {t.usuario_nombre || "Sistema"}
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
