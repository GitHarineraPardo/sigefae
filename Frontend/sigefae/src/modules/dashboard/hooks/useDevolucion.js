import { useState } from "react";
import { API } from "../constants/api";
import { isFinalState } from "../helpers/formatters";

export function useDevolucion(obtenerToken, activeTab, setSelectedTareaId, setSelectedRadicadoId, setTareasFlujo, setHistorialTrazabilidad, setComentarios, setMisTareas, setMisTareasCompletadas, setRadicados, userId) {
  const [showDevolverModal, setShowDevolverModal] = useState(false);
  const [devolverForm, setDevolverForm] = useState({ tarea_destino_id: "", observacion: "", retorno_directo: true });
  const [devolviendo, setDevolviendo] = useState(false);

  const handleDevolverTarea = async (tareaId) => {
    if (!devolverForm.tarea_destino_id || !devolverForm.observacion.trim()) {
      alert("Debe seleccionar un paso de destino y escribir un motivo de devolución."); return;
    }
    setDevolviendo(true);
    try {
      const res = await fetch(`${API}/tarea/${tareaId}/devolver`, {
        method: "POST", headers: { "Content-Type": "application/json", Authorization: `Bearer ${obtenerToken()}` },
        body: JSON.stringify({ tarea_destino_id: parseInt(devolverForm.tarea_destino_id), observacion: devolverForm.observacion.trim(), retorno_directo: devolverForm.retorno_directo })
      });
      if (!res.ok) { const errData = await res.json(); throw new Error(errData.error || "Error al devolver la tarea"); }
      setShowDevolverModal(false); setDevolverForm({ tarea_destino_id: "", observacion: "", retorno_directo: true });
      if (activeTab === "tareas") { setSelectedTareaId(null); }
      else { setSelectedRadicadoId(null); }
      setTareasFlujo([]); setHistorialTrazabilidad([]); setComentarios([]);
      const listaRes = await fetch(`${API}/documentoradicado`, { headers: { Authorization: `Bearer ${obtenerToken()}` } });
      const listaData = await listaRes.json();
      if (Array.isArray(listaData)) {
        if (activeTab === "tareas") {
          setMisTareas(listaData.filter(r => r.usuario_actual_id === userId && !isFinalState(r.estado_posesion)));
          setMisTareasCompletadas(listaData.filter(r => isFinalState(r.estado_posesion)));
        } else if (activeTab === "radicados" || activeTab === "finanzas") { setRadicados(listaData); }
      }
      alert("Tarea devuelta correctamente");
    } catch (err) { alert("Error: " + err.message); }
    finally { setDevolviendo(false); }
  };

  return { showDevolverModal, setShowDevolverModal, devolverForm, setDevolverForm, devolviendo, handleDevolverTarea };
}