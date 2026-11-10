export default function RenderComentarios({
  comentarios, nuevoComentario, setNuevoComentario, enviandoComentario,
  handleEnviarComentario, radicadoId, userId, readOnly = false
}) {
  return (
    <div className="doc-section">
      <h4>
        <i className="fa-solid fa-comment-dots"></i>{" "}
        Comentarios ({comentarios.length})
      </h4>

      <div
        style={{
          maxHeight: 240,        
          overflowY: "auto",
          marginBottom: 12,
          paddingRight: 4
        }}
      >
        {comentarios.length === 0 ? (
          <p style={{ color: "#6b7280", fontSize: "0.9em" }}>
            No hay comentarios.
          </p>
        ) : (
          comentarios.map((c) => (
            <div
              key={c.id}
              style={{
                background: "#f3f4f6",
                borderRadius: 8,
                padding: "10px 12px",
                marginBottom: 8
              }}
            >
              <div
                style={{
                  display: "flex",
                  justifyContent: "space-between",
                  alignItems: "center",
                  marginBottom: 4
                }}
              >
                <strong style={{ fontSize: "0.85em", color: "#1f2937" }}>
                  {c.usuario_nombre || `Usuario #${c.usuario_id}`}
                </strong>

                <span style={{ fontSize: "0.75em", color: "#6b7280" }}>
                  {new Date(c.fecha).toLocaleString()}
                </span>
              </div>

              <p
                style={{
                  margin: 0,
                  fontSize: "0.9em",
                  color: "#374151",
                  whiteSpace: "pre-wrap"
                }}
              >
                {c.descripcion}
              </p>
            </div>
          ))
        )}
      </div>

      {readOnly ? (
        <div
          style={{
            padding: "10px 12px",
            borderRadius: 6,
            background: "#f3f4f6",
            color: "#6b7280",
            fontSize: "0.9em"
          }}
        >
          <i className="fa-solid fa-lock"></i>{" "}
          El documento está finalizado. No se pueden agregar más comentarios.
        </div>
      ) : (
        <div style={{ display: "flex", gap: 8, alignItems: "flex-start" }}>
          <textarea
            value={nuevoComentario}
            onChange={(e) => setNuevoComentario(e.target.value)}
            placeholder="Escribe un comentario..."
            rows={2}
            style={{
              flex: 1,
              padding: 8,
              borderRadius: 6,
              border: "1px solid #d1d5db",
              resize: "vertical",
              fontFamily: "inherit",
              fontSize: "0.9em",
              minHeight: 70 
            }}
          />

          <button
            className="doc-btn doc-btn-primary"
            onClick={() => handleEnviarComentario(radicadoId, userId)}
            disabled={enviandoComentario || !nuevoComentario.trim()}
            style={{ marginTop: 2 }}
          >
            <i className="fa-solid fa-paper-plane"></i>{" "}
            {enviandoComentario ? "Enviando..." : "Enviar"}
          </button>
        </div>
      )}
    </div>
  );
}
