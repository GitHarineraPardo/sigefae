/* eslint-disable react-hooks/set-state-in-effect */
import { useState, useEffect, useCallback, useRef } from "react";
import { API } from "../constants/api.js";

const fetchNoStore = (url, options = {}) =>
  fetch(url, {
    ...options,
    cache: "no-store",
    headers: {
      ...options.headers,
      "Cache-Control": "no-cache, no-store, must-revalidate",
      Pragma: "no-cache",
    },
  });

export function useGastosProveedor(obtenerToken, activeTab) {
  const [data, setData] = useState(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  // Filtros
  const [fechaDesde, setFechaDesde] = useState("");
  const [fechaHasta, setFechaHasta] = useState("");
  const [proveedoresSeleccionados, setProveedoresSeleccionados] = useState([]);

  // Pestañas internas
  const [activeSubTab, setActiveSubTab] = useState("general"); // "general" | "por_proveedor" | "desglose"
  const [tipoDesglose, setTipoDesglose] = useState("items"); // "items" | "normas"
  const [busquedaItem, setBusquedaItem] = useState("");
  const [busquedaProveedor, setBusquedaProveedor] = useState("");

  const requestSequence = useRef(0);

  const fetchGastos = useCallback(async () => {
    const requestId = ++requestSequence.current;
    setLoading(true);
    setError(null);

    try {
      const params = new URLSearchParams();
      if (fechaDesde) params.append("fecha_inicio", fechaDesde);
      if (fechaHasta) params.append("fecha_fin", fechaHasta);
      params.append("_t", String(Date.now()));

      const res = await fetchNoStore(`${API}/gastos-proveedor?${params.toString()}`, {
        headers: { Authorization: `Bearer ${obtenerToken()}` },
      });

      if (!res.ok) {
        const errJson = await res.json().catch(() => ({}));
        throw new Error(errJson.error || `Error ${res.status}: no se pudieron cargar los gastos`);
      }

      const json = await res.json();
      if (requestId === requestSequence.current) {
        setData(json);
      }
    } catch (err) {
      if (requestId === requestSequence.current) {
        setError(err.message || "Error al cargar los datos de gastos");
      }
    } finally {
      if (requestId === requestSequence.current) {
        setLoading(false);
      }
    }
  }, [fechaDesde, fechaHasta, obtenerToken]);

  useEffect(() => {
    if (activeTab === "gastos-proveedor") {
      void fetchGastos();
    }
  }, [activeTab, fetchGastos]);

  // Manejadores de selección de proveedores
  const agregarProveedor = (id) => {
    const numId = Number(id);
    if (!numId) return;
    setProveedoresSeleccionados((prev) => (prev.includes(numId) ? prev : [...prev, numId]));
  };

  const quitarProveedor = (id) => {
    const numId = Number(id);
    setProveedoresSeleccionados((prev) => prev.filter((p) => p !== numId));
  };

  const seleccionarTodosProveedores = () => {
    if (!data?.proveedores) return;
    setProveedoresSeleccionados(data.proveedores.map((p) => p.id));
  };

  const limpiarProveedores = () => {
    setProveedoresSeleccionados([]);
  };

  const limpiarFiltros = () => {
    setFechaDesde("");
    setFechaHasta("");
    setProveedoresSeleccionados([]);
    setBusquedaItem("");
    setBusquedaProveedor("");
  };

  return {
    data,
    loading,
    error,
    refetch: fetchGastos,
    // Filtros
    fechaDesde,
    setFechaDesde,
    fechaHasta,
    setFechaHasta,
    proveedoresSeleccionados,
    agregarProveedor,
    quitarProveedor,
    seleccionarTodosProveedores,
    limpiarProveedores,
    limpiarFiltros,
    // Vistas y búsquedas
    activeSubTab,
    setActiveSubTab,
    tipoDesglose,
    setTipoDesglose,
    busquedaItem,
    setBusquedaItem,
    busquedaProveedor,
    setBusquedaProveedor,
  };
}
