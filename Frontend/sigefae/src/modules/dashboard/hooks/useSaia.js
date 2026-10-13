import { useState, useEffect } from "react";
import { API } from "../constants/api";

const fetchNoStore = (url, options = {}) => fetch(url, {
  ...options,
  cache: "no-store",
  headers: {
    ...options.headers,
    "Cache-Control": "no-cache, no-store, must-revalidate",
    Pragma: "no-cache",
  },
});

export function useSaia(obtenerToken) {
  const [saiaModalOpen, setSaiaModalOpen] = useState(false);
  const [saiaRadicado, setSaiaRadicado] = useState(null);
  const [saiaActiveTab, setSaiaActiveTab] = useState("info");
  const [saiaAnexoIdx, setSaiaAnexoIdx] = useState(0);
  const [saiaPdfUrl, setSaiaPdfUrl] = useState(null);

  // PDF blob para visor SAIA
  useEffect(() => {
    let cancelled = false;

    const loadPdf = async () => {
      if (!saiaModalOpen || !saiaRadicado) {
        setSaiaPdfUrl(prev => {
          if (prev) URL.revokeObjectURL(prev);
          return null;
        });
        return;
      }
      const archivos = Array.isArray(saiaRadicado.archivos) ? saiaRadicado.archivos : [];
      const archivosPdf = archivos.filter(a => {
        const ext = (a.extension || a.nombre?.split('.').pop() || '').toLowerCase();
        return ['pdf', 'png', 'jpg', 'jpeg', 'webp', 'gif'].includes(ext);
      });
      const anexoActual = archivosPdf[saiaAnexoIdx];
      if (!anexoActual) {
        setSaiaPdfUrl(prev => {
          if (prev) URL.revokeObjectURL(prev);
          return null;
        });
        return;
      }
      setSaiaPdfUrl(prev => {
        if (prev) URL.revokeObjectURL(prev);
        return null;
      });
      try {
        const res = await fetchNoStore(
          `${API}/archivo/${anexoActual.id}/download?_t=${Date.now()}`,
          { headers: { Authorization: `Bearer ${obtenerToken()}` } }
        );
        if (!res.ok) throw new Error("No se pudo cargar el PDF");
        const blob = await res.blob();
        if (!cancelled) {
          const url = URL.createObjectURL(blob);
          setSaiaPdfUrl(prev => { if (prev) URL.revokeObjectURL(prev); return url; });
        }
      } catch {
        if (!cancelled) {
          setSaiaPdfUrl(prev => {
            if (prev) URL.revokeObjectURL(prev);
            return null;
          });
        }
      }
    };
    void loadPdf();
    return () => { cancelled = true; };
  }, [saiaModalOpen, saiaRadicado, saiaAnexoIdx, obtenerToken]);

  const openSaia = async (radicadoBase, fromTab, setSelectedTareaId, setSelectedRadicadoId, recargarFlujo, recargarTrazabilidad, recargarNormas, recargarComentarios) => {
    setSaiaPdfUrl(prev => {
      if (prev) URL.revokeObjectURL(prev);
      return null;
    });
    try {
      const res = await fetchNoStore(`${API}/documentoradicado/${radicadoBase.id}?_t=${Date.now()}`, { headers: { Authorization: `Bearer ${obtenerToken()}` } });
      if (!res.ok) throw new Error(`No se pudo cargar el radicado (${res.status})`);
      const rad = await res.json();
      if (rad?.id) {
        if (rad.ruta?.id && !rad.ruta?.area) {
          try {
            const rutasRes = await fetchNoStore(`${API}/rutas?_t=${Date.now()}`, { headers: { Authorization: `Bearer ${obtenerToken()}` } });
            const rutasList = await rutasRes.json();
            const rutaCompleta = (Array.isArray(rutasList) ? rutasList : []).find(r => r.id === rad.ruta.id);
            if (rutaCompleta) rad.ruta.area = rutaCompleta.area;
          } catch { /* enriquecimiento de ruta opcional, no crítico */ }
        }
        setSaiaRadicado(rad);
        setSaiaAnexoIdx(0); setSaiaActiveTab("info");
        if (fromTab === "tareas") { setSelectedTareaId(rad.id); setSelectedRadicadoId(null); }
        else { setSelectedRadicadoId(rad.id); setSelectedTareaId(null); }
        await recargarFlujo(rad.id); await recargarTrazabilidad(rad.id); await recargarNormas(rad.id); await recargarComentarios(rad.id);
        setSaiaModalOpen(true);
      }
    } catch (err) { console.error(err); alert("Error cargando detalle del radicado"); }
  };

  return {
    saiaModalOpen, setSaiaModalOpen, saiaRadicado, setSaiaRadicado,
    saiaActiveTab, setSaiaActiveTab, saiaAnexoIdx, setSaiaAnexoIdx, saiaPdfUrl, setSaiaPdfUrl,
    openSaia
  };
}