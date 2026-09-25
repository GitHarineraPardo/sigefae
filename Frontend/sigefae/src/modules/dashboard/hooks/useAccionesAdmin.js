import { API } from "../constants/api";

export function useAccionesAdmin(obtenerToken, setRadicadoDetail, setSaiaRadicado) {
  const solicitarRechazo = async (documentoId) => {
    if (!confirm("¿Desea solicitar el rechazo de este documento?")) return;
    const motivo = prompt("Escriba un motivo (opcional):", "");
    if (motivo === null) return;
    try {
      const res = await fetch(`${API}/documentoradicado/${documentoId}/solicitar-rechazo`, {
        method: "POST", headers: { "Content-Type": "application/json", Authorization: `Bearer ${obtenerToken()}` },
        body: JSON.stringify({ mensaje: motivo || "Solicitud de rechazo" }),
      });
      if (!res.ok) { const err = await res.json(); throw new Error(err.error || "Error enviando solicitud"); }
      alert("Solicitud de rechazo enviada a los administradores");
    } catch (err) { alert("Error: " + err.message); }
  };

  const marcarCompletado = async (radicadoId) => {
    if (!confirm("¿Marcar este radicado como completado? Esta acción es final.")) return;
    const motivo = prompt("Escriba un mensaje (opcional):", "");
    if (motivo === null) return;
    try {
      const res = await fetch(`${API}/documentoradicado/${radicadoId}/completar`, {
        method: "POST", headers: { "Content-Type": "application/json", Authorization: `Bearer ${obtenerToken()}` },
        body: JSON.stringify({ mensaje: motivo || "Marcado como completado por admin" }),
      });
      if (!res.ok) { const err = await res.json(); throw new Error(err.error || "Error al completar"); }
      alert("Radicado marcado como completado");
      const nuevo = await fetch(`${API}/documentoradicado/${radicadoId}`, { headers: { Authorization: `Bearer ${obtenerToken()}` } });
      const data = await nuevo.json();
      setRadicadoDetail(data);
      if (setSaiaRadicado) setSaiaRadicado(data);
    } catch (err) { alert("Error: " + err.message); }
  };

  const adminRechazar = async (radicadoId, onSuccess) => {
    if (!confirm("¿Confirmar rechazo definitivo de este radicado?")) return;
    const motivo = prompt("Motivo (opcional):", "");
    if (motivo === null) return;
    try {
      const res = await fetch(`${API}/documentoradicado/${radicadoId}/rechazar`, {
        method: "POST", headers: { "Content-Type": "application/json", Authorization: `Bearer ${obtenerToken()}` },
        body: JSON.stringify({ mensaje: motivo || "Rechazado" }),
      });
      if (!res.ok) { const err = await res.json(); throw new Error(err.error || "Error al rechazar"); }
      alert("Radicado marcado como rechazado");
      const nuevo = await fetch(`${API}/documentoradicado/${radicadoId}`, { headers: { Authorization: `Bearer ${obtenerToken()}` } });
      const data = await nuevo.json();
      setRadicadoDetail(data);
      if (setSaiaRadicado) setSaiaRadicado(data);
      if (onSuccess) onSuccess();
    } catch (err) { alert("Error: " + err.message); }
  };

  return { solicitarRechazo, marcarCompletado, adminRechazar };
}