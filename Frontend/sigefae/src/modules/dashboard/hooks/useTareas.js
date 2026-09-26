import { useState, useEffect, useRef } from "react";
import { API } from "../constants/api";
import { isFinalState } from "../helpers/formatters";

export function useTareas(obtenerToken, activeTab, userId, userCargoId) {
  const [misTareas, setMisTareas] = useState([]);
  const [misTareasCompletadas, setMisTareasCompletadas] = useState([]);
  const [selectedTareaId, setSelectedTareaId] = useState(null);
  const [tareaDetail, setTareaDetail] = useState(null);
  const [loadingTareas, setLoadingTareas] = useState(false);
  const [tareasSubTab, setTareasSubTab] = useState("activas");
  const [searchTareas, setSearchTareas] = useState("");
  const [sortTareas, setSortTareas] = useState("fecha_desc");
  const requestSequence = useRef(0);

  const recargarTareas = async () => {
    const requestId = ++requestSequence.current;
    setLoadingTareas(true);
    try {
      const response = await fetch(`${API}/documentoradicado?_t=${Date.now()}`, {
        cache: "no-store",
        headers: { Authorization: `Bearer ${obtenerToken()}` }
      });
      if (!response.ok) throw new Error("No se pudieron actualizar las tareas");
      const data = await response.json();
      if (requestId !== requestSequence.current || !Array.isArray(data)) return data;

      setMisTareas(data.filter(r => (Number(r.usuario_actual_id) === Number(userId) || (Number(r.cargo_actual_id) > 0 && Number(r.cargo_actual_id) === Number(userCargoId))) && !isFinalState(r.estado_posesion)));
      setMisTareasCompletadas(data.filter(r => isFinalState(r.estado_posesion)));
      return data;
    } catch (error) {
      console.error(error);
      return null;
    } finally {
      if (requestId === requestSequence.current) setLoadingTareas(false);
    }
  };

  const sincronizarRadicado = (radicado) => {
    if (!radicado?.id) return;
    const asignado = Number(radicado.usuario_actual_id) === Number(userId) ||
      (Number(radicado.cargo_actual_id) > 0 && Number(radicado.cargo_actual_id) === Number(userCargoId));
    const finalizado = isFinalState(radicado.estado_posesion);

    setMisTareas(current => {
      const restantes = current.filter(item => item.id !== radicado.id);
      return !finalizado && asignado ? [radicado, ...restantes] : restantes;
    });
    setMisTareasCompletadas(current => {
      const restantes = current.filter(item => item.id !== radicado.id);
      return finalizado ? [radicado, ...restantes] : restantes;
    });
  };

  useEffect(() => {
    if (activeTab !== "tareas") return;
    void recargarTareas();
    return () => { requestSequence.current += 1; };
  }, [activeTab, userId, userCargoId, obtenerToken]);

  useEffect(() => {
    if (!selectedTareaId) return;
    setTareaDetail(null);
    const t = `?_t=${new Date().getTime()}`;
    fetch(`${API}/documentoradicado/${selectedTareaId}${t}`, { headers: { Authorization: `Bearer ${obtenerToken()}` } })
      .then((res) => res.json())
      .then(async (data) => {
        if (data?.id) {
          if (data.ruta?.id && !data.ruta?.area) {
            try {
              const rutasRes = await fetch(`${API}/rutas`, { headers: { Authorization: `Bearer ${obtenerToken()}` } });
              const rutasList = await rutasRes.json();
              const rutaCompleta = (Array.isArray(rutasList) ? rutasList : []).find(r => r.id === data.ruta.id);
              if (rutaCompleta) data.ruta.area = rutaCompleta.area;
            } catch (e) {}
          }
          setTareaDetail(data);
        }
      })
      .catch((err) => console.error(err));
  }, [selectedTareaId, obtenerToken]);

  const getFilteredTareas = (baseList) => {
    let result = [...baseList];
    if (searchTareas) {
      const q = searchTareas.toLowerCase();
      result = result.filter(t => (t.numero_radicado?.toLowerCase().includes(q)) || (t.documento_comercial?.numero_documento?.toLowerCase().includes(q)));
    }
    result.sort((a, b) => {
      if (sortTareas === 'fecha_desc') return new Date(b.fecha_creacion) - new Date(a.fecha_creacion);
      if (sortTareas === 'fecha_asc') return new Date(a.fecha_creacion) - new Date(b.fecha_creacion);
      if (sortTareas === 'estado') return (a.estado_posesion || '').localeCompare(b.estado_posesion || '');
      return 0;
    });
    return result;
  };

  return {
    misTareas, setMisTareas, misTareasCompletadas, setMisTareasCompletadas, recargarTareas, sincronizarRadicado, selectedTareaId, setSelectedTareaId,
    tareaDetail, setTareaDetail, loadingTareas, tareasSubTab, setTareasSubTab,
    searchTareas, setSearchTareas, sortTareas, setSortTareas, getFilteredTareas
  };
}