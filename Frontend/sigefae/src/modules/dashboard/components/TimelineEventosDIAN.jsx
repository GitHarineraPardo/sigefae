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

  // Inferencia RADIAN: Si existe un evento posterior (032, 034 o 031), los eventos anteriores
  // obligatoriamente ya ocurrieron (provienen de Inbound o registro previo de la DIAN)
  const has030 = Boolean(eventosMap["030"] || eventosMap["ACKNOWLEDGED"]);
  const has032 = Boolean(eventosMap["032"] || eventosMap["RECEIVEDGOODS"]);
  const has034 = Boolean(eventosMap["034"] || eventosMap["ACCEPTED"]);
  const has031 = Boolean(eventosMap["031"] || eventosMap["REJECTED"]);

  if (!has030 && (has032 || has034 || has031)) {
    const ref = eventosMap["032"] || eventosMap["034"] || eventosMap["031"] || {};
    const synthetic = {
      codigo_dian: "030",
      tipo_evento: "ACKNOWLEDGED",
      fecha_envio: ref.fecha_envio || ref.created_at || new Date().toISOString(),
      usuario_nombre: "Buzón Inbound (Automático)",
      observacion: "Acuse de recibo generado automáticamente por el proveedor tecnológico al recibir la factura",
    };
    eventosMap["030"] = synthetic;
    eventosMap["ACKNOWLEDGED"] = synthetic;
  }

  if (!has032 && (has034 || has031)) {
    const ref = eventosMap["034"] || eventosMap["031"] || {};
    const synthetic = {
      codigo_dian: "032",
      tipo_evento: "RECEIVEDGOODS",
      fecha_envio: ref.fecha_envio || ref.created_at || new Date().toISOString(),
      usuario_nombre: "Buzón Inbound (Automático)",
      observacion: "Recibo de bienes o prestación de servicios registrado previamente ante la DIAN",
    };
    eventosMap["032"] = synthetic;
    eventosMap["RECEIVEDGOODS"] = synthetic;
  }

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
