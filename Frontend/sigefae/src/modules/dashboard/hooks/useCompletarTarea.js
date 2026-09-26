import { useState } from "react";
import { API } from "../constants/api";

export function useCompletarTarea(obtenerToken, userId, activeTab, setTareasFlujo, setTareaDetail, setRadicadoDetail, setRadicados, setSaiaRadicado, recargarTareas, sincronizarRadicado) {
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

      const flujoRes = await fetch(`${API}/documentoradicado/${radicadoId}/tareas?_t=${Date.now()}`, { cache: "no-store", headers: { Authorization: `Bearer ${obtenerToken()}` } });
      const flujoData = await flujoRes.json();
      setTareasFlujo(Array.isArray(flujoData) ? flujoData : []);

      const detalleRes = await fetch(`${API}/documentoradicado/${radicadoId}?_t=${Date.now()}`, { cache: "no-store", headers: { Authorization: `Bearer ${obtenerToken()}` } });
      const detalleData = await detalleRes.json();
      if (detalleData?.id) {
        if (activeTab === "tareas") setTareaDetail(detalleData);
        if (activeTab === "radicados") setRadicadoDetail(detalleData);
        if (setSaiaRadicado) setSaiaRadicado(detalleData);
        sincronizarRadicado(detalleData);
      }

      const listaData = await recargarTareas();
      if (activeTab === "radicados" && Array.isArray(listaData)) setRadicados(listaData);

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