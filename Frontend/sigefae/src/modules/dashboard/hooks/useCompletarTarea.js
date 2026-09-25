import { useState } from "react";
import { API } from "../constants/api";
import { isFinalState } from "../helpers/formatters";

export function useCompletarTarea(obtenerToken, userId, activeTab, setTareasFlujo, setTareaDetail, setRadicadoDetail, setMisTareas, setMisTareasCompletadas, setRadicados, setSaiaRadicado) {
  const [completandoTarea, setCompletandoTarea] = useState(false);

  const handleCompletarTarea = async (tareaId, radicadoId) => {
    if (completandoTarea) return;
    setCompletandoTarea(true);
    try {
      const res = await fetch(`${API}/tarea/${tareaId}/completar`, { method: "PATCH", headers: { Authorization: `Bearer ${obtenerToken()}` } });
      let errMsg = null;
      if (!res.ok) {
        try { const errData = await res.json(); errMsg = errData.error || ""; } catch (e) { errMsg = "Error completando tarea"; }
      }
      if (errMsg && !errMsg.toLowerCase().includes("ya está completada")) throw new Error(errMsg);
      if (!errMsg) alert("Tarea completada correctamente");

      const flujoRes = await fetch(`${API}/documentoradicado/${radicadoId}/tareas`, { headers: { Authorization: `Bearer ${obtenerToken()}` } });
      const flujoData = await flujoRes.json();
      setTareasFlujo(Array.isArray(flujoData) ? flujoData : []);

      const detalleRes = await fetch(`${API}/documentoradicado/${radicadoId}`, { headers: { Authorization: `Bearer ${obtenerToken()}` } });
      const detalleData = await detalleRes.json();
      if (detalleData?.id) {
        if (activeTab === "tareas") setTareaDetail(detalleData);
        if (activeTab === "radicados") setRadicadoDetail(detalleData);
        if (setSaiaRadicado) setSaiaRadicado(detalleData);
      }

      const listaRes = await fetch(`${API}/documentoradicado`, { headers: { Authorization: `Bearer ${obtenerToken()}` } });
      const listaData = await listaRes.json();
      if (Array.isArray(listaData)) {
        if (activeTab === "tareas") {
          setMisTareas(listaData.filter(r => r.usuario_actual_id === userId && !isFinalState(r.estado_posesion)));
          setMisTareasCompletadas(listaData.filter(r => isFinalState(r.estado_posesion)));
        } else if (activeTab === "radicados") setRadicados(listaData);
      }

      if (!errMsg) {
        const siguiente = (Array.isArray(flujoData) ? flujoData : []).find(t => t.estado?.nombre === "En Proceso");
        if (siguiente?.usuario_asignado?.id && siguiente.usuario_asignado.id !== userId) {
          await fetch(`${API}/notificacion`, {
            method: "POST", headers: { "Content-Type": "application/json", Authorization: `Bearer ${obtenerToken()}` },
            body: JSON.stringify({ usuario_id: siguiente.usuario_asignado.id, documento_radicado_id: radicadoId, mensaje: `Te asignaron el radicado #${detalleData?.numero_radicado || radicadoId} — Paso: ${siguiente.descripcion || 'revisar'}`, estado: "Pendiente", tipo: "Asignacion", fecha_creacion: new Date().toISOString() })
          });
        }
      }
    } catch (err) { alert("Error: " + err.message); }
    finally { setCompletandoTarea(false); }
  };

  return { completandoTarea, handleCompletarTarea };
}