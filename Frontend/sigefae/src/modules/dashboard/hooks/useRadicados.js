import { useState, useEffect } from "react";
import { API } from "../constants/api";
import { isFinalState } from "../helpers/formatters";

const fetchNoStore = (url, options = {}) => fetch(url, {
  ...options,
  cache: "no-store",
  headers: {
    ...options.headers,
    "Cache-Control": "no-cache, no-store, must-revalidate",
    Pragma: "no-cache",
  },
});

export function useRadicados(obtenerToken, activeTab) {
  const [radicados, setRadicados] = useState([]);
  const [selectedRadicadoId, setSelectedRadicadoId] = useState(null);
  const [radicadoDetail, setRadicadoDetail] = useState(null);
  const [loadingRadicados, setLoadingRadicados] = useState(false);
  const [searchRadicados, setSearchRadicados] = useState("");
  const [sortRadicados, setSortRadicados] = useState("fecha_desc");
  const [filtroZona, setFiltroZona] = useState("");
  const [tareasPorRadicado, setTareasPorRadicado] = useState({});

  useEffect(() => {
    if (activeTab !== "radicados" && activeTab !== "finanzas") return;
    const controller = new AbortController();
    const fetchData = async () => {
      setLoadingRadicados(true);
      const t = `?_t=${new Date().getTime()}`;
      try {
        const res = await fetchNoStore(`${API}/documentoradicado${t}`, {
          headers: { Authorization: `Bearer ${obtenerToken()}` },
          signal: controller.signal,
        });
        const data = await res.json();
        if (!Array.isArray(data)) return;
        setRadicados(data);
        const tareasMap = {};
        await Promise.all(data.map(async (rad) => {
          if (isFinalState(rad.estado_posesion)) return;
          try {
            const tr = `?_t=${new Date().getTime()}`;
            const tRes = await fetchNoStore(`${API}/documentoradicado/${rad.id}/tareas${tr}`, { headers: { Authorization: `Bearer ${obtenerToken()}` } });
            const tareas = await tRes.json();
            const tareaActiva = (Array.isArray(tareas) ? tareas : []).find(t => t.estado?.nombre === "En Proceso");
            if (tareaActiva) tareasMap[rad.id] = tareaActiva;
          } catch { /* ignorar errores de tareas individuales */ }
        }));
        setTareasPorRadicado(tareasMap);
      } catch (err) {
        if (!controller.signal.aborted) console.error(err);
      } finally {
        if (!controller.signal.aborted) setLoadingRadicados(false);
      }
    };
    void fetchData();
    return () => controller.abort();
  }, [activeTab, obtenerToken]);

  useEffect(() => {
    if (!selectedRadicadoId) return;
    const controller = new AbortController();
    const fetchDetail = async () => {
      setRadicadoDetail(null);
      const t = `?_t=${new Date().getTime()}`;
      try {
        const res = await fetchNoStore(`${API}/documentoradicado/${selectedRadicadoId}${t}`, {
          headers: { Authorization: `Bearer ${obtenerToken()}` },
          signal: controller.signal,
        });
        if (!res.ok) throw new Error(`No se pudo cargar el radicado (${res.status})`);
        const data = await res.json();
        if (data?.id) {
          if (data.ruta?.id && !data.ruta?.area) {
            try {
              const rutasRes = await fetchNoStore(`${API}/rutas${t}`, { headers: { Authorization: `Bearer ${obtenerToken()}` } });
              const rutasList = await rutasRes.json();
              const rutaCompleta = (Array.isArray(rutasList) ? rutasList : []).find(r => r.id === data.ruta.id);
              if (rutaCompleta) data.ruta.area = rutaCompleta.area;
            } catch { /* enriquecimiento opcional */ }
          }
          setRadicadoDetail(data);
        }
      } catch (err) {
        if (!controller.signal.aborted) console.error(err);
      }
    };
    void fetchDetail();
    return () => controller.abort();
  }, [selectedRadicadoId, obtenerToken]);

  const getFilteredRadicados = () => {
    let result = [...radicados];
    if (filtroZona) {
      result = result.filter(r => r.ruta?.zona === filtroZona);
    }
    if (searchRadicados) {
      const q = searchRadicados.toLowerCase();
      result = result.filter(r =>
        (r.numero_radicado?.toLowerCase().includes(q)) ||
        (r.documento_comercial?.numero_documento?.toLowerCase().includes(q)) ||
        (r.documento_comercial?.proveedor?.razon_social?.toLowerCase().includes(q))
      );
    }
    result.sort((a, b) => {
      const fechaPrincipal = (radicado) => radicado.documento_comercial?.fecha_recepcion_dian || radicado.fecha_radicacion;
      if (sortRadicados === 'fecha_desc') return new Date(fechaPrincipal(b)) - new Date(fechaPrincipal(a));
      if (sortRadicados === 'fecha_asc') return new Date(fechaPrincipal(a)) - new Date(fechaPrincipal(b));
      if (sortRadicados === 'estado') return (a.estado_posesion || '').localeCompare(b.estado_posesion || '');
      return 0;
    });
    return result;
  };

  return {
    radicados, setRadicados, selectedRadicadoId, setSelectedRadicadoId, radicadoDetail, setRadicadoDetail,
    loadingRadicados, searchRadicados, setSearchRadicados, sortRadicados, setSortRadicados,
    filtroZona, setFiltroZona,
    tareasPorRadicado, getFilteredRadicados
  };
}