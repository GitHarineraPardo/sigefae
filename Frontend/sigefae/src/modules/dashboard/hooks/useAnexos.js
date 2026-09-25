import { useState } from "react";
import { API } from "../constants/api";

export function useAnexos(obtenerToken, puedeGestionarRecurso, esAdmin, activeTab, setSelectedTareaId, setSelectedRadicadoId, saiaModalOpen, saiaRadicado, setSaiaRadicado, saiaAnexoIdx, setSaiaAnexoIdx, saiaPdfUrl, setSaiaPdfUrl) {
  const [generandoPdf, setGenerandoPdf] = useState(false);

  const handleVerAnexo = async (archivoId) => {
    const win = window.open("", "_blank");
    try {
      const res = await fetch(`${API}/archivo/${archivoId}/download`, { headers: { Authorization: `Bearer ${obtenerToken()}` } });
      if (!res.ok) { if (win) win.close(); throw new Error("Error obteniendo archivo"); }
      const blob = await res.blob(); const url = window.URL.createObjectURL(blob);
      if (win) win.location.href = url; else window.open(url, "_blank");
      setTimeout(() => window.URL.revokeObjectURL(url), 60000);
    } catch (err) { if (win) win.close(); alert("Error: " + err.message); }
  };

  const handleDescargarAnexo = async (archivoId, nombre) => {
    try {
      const res = await fetch(`${API}/archivo/${archivoId}/download`, { headers: { Authorization: `Bearer ${obtenerToken()}` } });
      if (!res.ok) throw new Error("Error descargando archivo");
      const blob = await res.blob(); const url = window.URL.createObjectURL(blob);
      const a = document.createElement("a"); a.href = url; a.download = nombre;
      document.body.appendChild(a); a.click(); document.body.removeChild(a);
      setTimeout(() => window.URL.revokeObjectURL(url), 60000);
    } catch (err) { alert("Error: " + err.message); }
  };

  const handleBorrarAnexo = async (archivo, radicadoId) => {
    const archivoId = archivo?.id || archivo;
    const creadoPor = Number(archivo?.creado_por_id || archivo?.creado_por?.id || 0);
    if (!esAdmin && !puedeGestionarRecurso(creadoPor)) { alert("Solo el usuario que subió el anexo o el administrador pueden eliminarlo."); return; }
    if (!confirm("¿Está seguro de eliminar este archivo?")) return;
    try {
      const res = await fetch(`${API}/archivo/${archivoId}`, { method: "DELETE", headers: { Authorization: `Bearer ${obtenerToken()}` } });
      if (!res.ok) throw new Error("Error eliminando archivo");
      alert("Archivo eliminado");
      if (activeTab === "tareas") { setSelectedTareaId(null); setTimeout(() => setSelectedTareaId(radicadoId), 10); }
      else if (activeTab === "radicados") { setSelectedRadicadoId(null); setTimeout(() => setSelectedRadicadoId(radicadoId), 10); }
      if (saiaModalOpen && saiaRadicado?.id === radicadoId) {
        setSaiaRadicado(prev => ({ ...prev, archivos: (prev.archivos || []).filter(a => a.id !== archivoId) }));
        const pdfList = (saiaRadicado.archivos || []).filter(a => a.extension?.toLowerCase() === 'pdf' || a.nombre?.toLowerCase().endsWith('.pdf'));
        if (pdfList[saiaAnexoIdx]?.id === archivoId) { setSaiaAnexoIdx(0); if (saiaPdfUrl) { URL.revokeObjectURL(saiaPdfUrl); setSaiaPdfUrl(null); } }
      }
    } catch (err) { alert("Error: " + err.message); }
  };

  const handleSubirAnexo = async (e, radicadoId, setSelectedTareaId, setSelectedRadicadoId, setSaiaRadicado, saiaModalOpen, saiaRadicado) => {
    const file = e.target.files[0]; if (!file) return;
    const formData = new FormData(); formData.append("file", file);
    try {
      const res = await fetch(`${API}/documentoradicado/${radicadoId}/anexos`, {
        method: "POST", headers: { Authorization: `Bearer ${obtenerToken()}` }, body: formData,
      });
      if (!res.ok) { const err = await res.json(); throw new Error(err.error || "Error subiendo archivo"); }
      const uploaded = await res.json();
      // 1. Actualizar estado y refetch de inmediato
      if (activeTab === "tareas") { setSelectedTareaId(null); setTimeout(() => setSelectedTareaId(radicadoId), 10); }
      else if (activeTab === "radicados") { setSelectedRadicadoId(null); setTimeout(() => setSelectedRadicadoId(radicadoId), 10); }

      if (saiaModalOpen && saiaRadicado?.id === radicadoId) {
        if (uploaded?.id) {
          setSaiaRadicado((prev) => {
            if (!prev) return prev;
            const archivos = prev.archivos || [];
            if (archivos.some((archivo) => archivo.id === uploaded.id)) return prev;
            return { ...prev, archivos: [...archivos, uploaded] };
          });
        }

        try {
          const radRes = await fetch(`${API}/documentoradicado/${radicadoId}?_t=${Date.now()}`, { headers: { Authorization: `Bearer ${obtenerToken()}` } });
          const updated = await radRes.json();
          if (updated?.id) {
            const archivosActualizados = [...(updated.archivos || [])];
            if (uploaded?.id && !archivosActualizados.some((archivo) => archivo.id === uploaded.id)) {
              archivosActualizados.push(uploaded);
            }
            setSaiaRadicado((prev) => ({ ...updated, archivos: archivosActualizados.length ? archivosActualizados : (prev?.archivos || []) }));
            const viewables = archivosActualizados.filter(a => {
              const ext = (a.extension || a.nombre?.split('.').pop() || '').toLowerCase();
              return ['pdf', 'png', 'jpg', 'jpeg', 'webp', 'gif'].includes(ext);
            });
            if (viewables.length > 0) {
              const idx = viewables.findIndex(a => a.nombre === file.name);
              if (idx !== -1 && setSaiaAnexoIdx) setSaiaAnexoIdx(idx);
            }
          }
        } catch (e) { console.error(e); }
      }

      // 2. Mostrar alerta sin bloquear el ciclo de renderizado
      setTimeout(() => {
        alert("Archivo subido correctamente");
      }, 50);
    } catch (err) { alert("Error: " + err.message); }
    e.target.value = "";
  };

  return { generandoPdf, setGenerandoPdf, handleVerAnexo, handleDescargarAnexo, handleBorrarAnexo, handleSubirAnexo };
}