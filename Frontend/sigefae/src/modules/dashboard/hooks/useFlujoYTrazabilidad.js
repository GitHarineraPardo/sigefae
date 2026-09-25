import { useState, useEffect } from "react";
import { API } from "../constants/api";

export function useFlujoYTrazabilidad(obtenerToken, selectedRadicadoId, selectedTareaId) {
  const [tareasFlujo, setTareasFlujo] = useState([]);
  const [historialTrazabilidad, setHistorialTrazabilidad] = useState([]);
  const [normasRepartoRadicado, setNormasRepartoRadicado] = useState([]);
  const [activosFijosRadicado, setActivosFijosRadicado] = useState([]);
  const [comentarios, setComentarios] = useState([]);
  const [nuevoComentario, setNuevoComentario] = useState("");
  const [enviandoComentario, setEnviandoComentario] = useState(false);

  const radicadoId = selectedRadicadoId || selectedTareaId;

  useEffect(() => {
    if (!radicadoId) {
      setTareasFlujo([]); setHistorialTrazabilidad([]); setNormasRepartoRadicado([]); setActivosFijosRadicado([]); setComentarios([]); return;
    }
    const headers = { Authorization: `Bearer ${obtenerToken()}` };
    const t = new Date().getTime();
    fetch(`${API}/documentoradicado/${radicadoId}/tareas?_t=${t}`, { headers })
      .then(r => r.json()).then(data => setTareasFlujo(Array.isArray(data) ? data : [])).catch(err => console.error(err));
    fetch(`${API}/trazabilidad?documento_radicado_id=${radicadoId}&_t=${t}`, { headers })
      .then(r => r.json()).then(data => setHistorialTrazabilidad(Array.isArray(data) ? data : [])).catch(err => console.error(err));
    fetch(`${API}/documentoradicado/${radicadoId}/normas-reparto?_t=${t}`, { headers })
      .then(r => r.json()).then(data => setNormasRepartoRadicado(Array.isArray(data) ? data : [])).catch(() => setNormasRepartoRadicado([]));
    fetch(`${API}/documentoradicado/${radicadoId}/activos-fijos?_t=${t}`, { headers })
      .then(r => r.json()).then(data => setActivosFijosRadicado(Array.isArray(data) ? data : [])).catch(() => setActivosFijosRadicado([]));

    let cancelled = false;
    setComentarios([]);
    fetch(`${API}/comentario?documento_radicado_id=${radicadoId}&_t=${t}`, { headers })
      .then(r => r.json())
      .then(data => { if (!cancelled) setComentarios(Array.isArray(data) ? data : []); })
      .catch(err => { if (!cancelled) { console.error(err); setComentarios([]); } });
    return () => { cancelled = true; };
  }, [radicadoId, obtenerToken]);

  const handleEnviarComentario = async (radicadoId, userId) => {
    if (!nuevoComentario.trim()) return;
    setEnviandoComentario(true);
    try {
      const res = await fetch(`${API}/comentario`, {
        method: "POST", headers: { "Content-Type": "application/json", Authorization: `Bearer ${obtenerToken()}` },
        body: JSON.stringify({ documento_radicado_id: radicadoId, usuario_id: userId, descripcion: nuevoComentario.trim() })
      });
      if (!res.ok) { const errData = await res.json(); throw new Error(errData.error || "Error enviando comentario"); }
      setNuevoComentario("");
      const listRes = await fetch(`${API}/comentario?documento_radicado_id=${radicadoId}&_t=${new Date().getTime()}`, { headers: { Authorization: `Bearer ${obtenerToken()}` } });
      const listData = await listRes.json();
      setComentarios(Array.isArray(listData) ? listData : []);
    } catch (err) { alert("Error: " + err.message); }
    finally { setEnviandoComentario(false); }
  };

  const recargarComentarios = async (radicadoId) => {
    const listRes = await fetch(`${API}/comentario?documento_radicado_id=${radicadoId}&_t=${new Date().getTime()}`, { headers: { Authorization: `Bearer ${obtenerToken()}` } });
    const listData = await listRes.json();
    setComentarios(Array.isArray(listData) ? listData : []);
  };

  const recargarNormas = async (radicadoId) => {
    const nrRes = await fetch(`${API}/documentoradicado/${radicadoId}/normas-reparto?_t=${new Date().getTime()}`, { headers: { Authorization: `Bearer ${obtenerToken()}` } });
    const nrData = await nrRes.json();
    setNormasRepartoRadicado(Array.isArray(nrData) ? nrData : []);
  };

  const recargarActivosFijos = async (radicadoId) => {
    const res = await fetch(`${API}/documentoradicado/${radicadoId}/activos-fijos?_t=${new Date().getTime()}`, { headers: { Authorization: `Bearer ${obtenerToken()}` } });
    const data = await res.json();
    setActivosFijosRadicado(Array.isArray(data) ? data : []);
  };

  const recargarFlujo = async (radicadoId) => {
    const flujoRes = await fetch(`${API}/documentoradicado/${radicadoId}/tareas?_t=${new Date().getTime()}`, { headers: { Authorization: `Bearer ${obtenerToken()}` } });
    const flujoData = await flujoRes.json();
    setTareasFlujo(Array.isArray(flujoData) ? flujoData : []);
  };

  const recargarTrazabilidad = async (radicadoId) => {
    const res = await fetch(`${API}/trazabilidad?documento_radicado_id=${radicadoId}&_t=${new Date().getTime()}`, { headers: { Authorization: `Bearer ${obtenerToken()}` } });
    const data = await res.json();
    setHistorialTrazabilidad(Array.isArray(data) ? data : []);
  };

  return {
    tareasFlujo, setTareasFlujo, historialTrazabilidad, setHistorialTrazabilidad,
    normasRepartoRadicado, setNormasRepartoRadicado,
    activosFijosRadicado, setActivosFijosRadicado,
    comentarios, setComentarios, nuevoComentario, setNuevoComentario, enviandoComentario,
    handleEnviarComentario, recargarComentarios, recargarNormas, recargarActivosFijos, recargarFlujo, recargarTrazabilidad
  };
}