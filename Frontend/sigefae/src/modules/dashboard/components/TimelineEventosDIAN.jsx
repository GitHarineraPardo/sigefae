import React from "react";

const EVENTOS_CONFIG = [
  {
    codigo: "030",
    tipo: "ACKNOWLEDGED",
    inicial: "AC",
    nombre: "Acuse de Recibo (030)",
    color: "#2563eb",
    side: "left",
  },
  {
    codigo: "032",
    tipo: "RECEIVEDGOODS",
    inicial: "RB",
    nombre: "Recibo Bienes / Servicios (032)",
    color: "#d97706",
    side: "right",
  },
  {
    codigo: "034",
    tipo: "ACCEPTED",
    inicial: "AE",
    nombre: "Aceptación Expresa (034)",
    color: "#16a34a",
    side: "left",
  },
  {
    codigo: "031",
    tipo: "REJECTED",
    inicial: "RC",
    nombre: "Reclamo / Rechazo (031)",
    color: "#dc2626",
    side: "right",
  },
];

export default function TimelineEventosDIAN({ eventosDian = [], cargando = false }) {
  // Mapear eventos por código o tipo
  const eventosMap = {};
  (eventosDian || []).forEach(e => {
    if (e.codigo_dian) eventosMap[e.codigo_dian] = e;
    if (e.tipo_evento) eventosMap[e.tipo_evento.toUpperCase()] = e;
  });

  return (
    <div className="dian-timeline-container">
      <div className="dian-timeline-title">
        <i className="fa-solid fa-stamp" style={{ color: "var(--pardo-blue)" }}></i>
        <span>Eventos Electrónicos DIAN (RADIAN)</span>
        {cargando && <i className="fa-solid fa-circle-notch fa-spin" style={{ fontSize: "11px", marginLeft: "auto", color: "#94a3b8" }}></i>}
      </div>

      <div className="dian-timeline-track">
        <div className="dian-timeline-spine"></div>

        {EVENTOS_CONFIG.map((cfg) => {
          const evento = eventosMap[cfg.codigo] || eventosMap[cfg.tipo];
          const isActivo = Boolean(evento);
          const isLeft = cfg.side === "left";

          return (
            <div key={cfg.codigo} className="dian-timeline-row">
              {/* Lado Izquierdo */}
              <div className={`dian-timeline-card-wrapper left`}>
                {isActivo && isLeft ? (
                  <div className="dian-timeline-card">
                    <div className="dian-timeline-card-title">
                      <span>{cfg.nombre}</span>
                    </div>
                    <div className="dian-timeline-card-meta">
                      <i className="fa-regular fa-clock" style={{ marginRight: 3 }}></i>
                      {new Date(evento.fecha_envio || evento.created_at).toLocaleString([], { dateStyle: 'short', timeStyle: 'short' })}
                      {evento.usuario_nombre && ` · ${evento.usuario_nombre}`}
                    </div>
                    {evento.observacion && (
                      <div className="dian-timeline-card-comment">
                        "{evento.observacion}"
                      </div>
                    )}
                    <div className="dian-timeline-arrow"></div>
                  </div>
                ) : !isActivo && isLeft ? (
                  <span className="dian-timeline-pending-label">{cfg.nombre}</span>
                ) : null}
              </div>

              {/* Bolita central */}
              <div
                className={`dian-timeline-circle ${isActivo ? "active" : "pending"}`}
                style={isActivo ? { backgroundColor: cfg.color } : {}}
                title={`${cfg.nombre} - ${isActivo ? "Emitido" : "Pendiente"}`}
              >
                {cfg.inicial}
              </div>

              {/* Lado Derecho */}
              <div className={`dian-timeline-card-wrapper right`}>
                {isActivo && !isLeft ? (
                  <div className="dian-timeline-card">
                    <div className="dian-timeline-card-title">
                      <span>{cfg.nombre}</span>
                    </div>
                    <div className="dian-timeline-card-meta">
                      <i className="fa-regular fa-clock" style={{ marginRight: 3 }}></i>
                      {new Date(evento.fecha_envio || evento.created_at).toLocaleString([], { dateStyle: 'short', timeStyle: 'short' })}
                      {evento.usuario_nombre && ` · ${evento.usuario_nombre}`}
                    </div>
                    {evento.observacion && (
                      <div className="dian-timeline-card-comment">
                        "{evento.observacion}"
                      </div>
                    )}
                    <div className="dian-timeline-arrow"></div>
                  </div>
                ) : !isActivo && !isLeft ? (
                  <span className="dian-timeline-pending-label">{cfg.nombre}</span>
                ) : null}
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
}
