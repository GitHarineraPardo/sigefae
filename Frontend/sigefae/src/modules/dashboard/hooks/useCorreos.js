import { useState, useEffect } from "react";
import { API } from "../constants/api";

export function useCorreos(obtenerToken, activeTab) {
  const [correos, setCorreos] = useState([]);
  const [selectedCorreoId, setSelectedCorreoId] = useState(null);
  const [correoDetail, setCorreoDetail] = useState(null);
  const [loading, setLoading] = useState(false);
  const [searchCorreos, setSearchCorreos] = useState("");
  const [sortCorreos, setSortCorreos] = useState("fecha_desc");

  useEffect(() => {
    if (activeTab !== "correos") return;
    const controller = new AbortController();
    const fetchData = async () => {
      setLoading(true);
      const t = `?_t=${new Date().getTime()}`;
      try {
        const res = await fetch(`${API}/correo${t}`, {
          headers: { Authorization: `Bearer ${obtenerToken()}` },
          signal: controller.signal,
        });
        const data = await res.json();
        if (Array.isArray(data)) setCorreos(data);
      } catch (err) {
        if (!controller.signal.aborted) console.error(err);
      } finally {
        if (!controller.signal.aborted) setLoading(false);
      }
    };
    void fetchData();
    return () => controller.abort();
  }, [activeTab, obtenerToken]);

  useEffect(() => {
    if (!selectedCorreoId) return;
    const controller = new AbortController();
    const fetchData = async () => {
      setCorreoDetail(null);
      const t = `?_t=${new Date().getTime()}`;
      try {
        const res = await fetch(`${API}/correo/${selectedCorreoId}${t}`, {
          headers: { Authorization: `Bearer ${obtenerToken()}` },
          signal: controller.signal,
        });
        const data = await res.json();
        if (data?.id) setCorreoDetail(data);
      } catch (err) {
        if (!controller.signal.aborted) console.error(err);
      }
    };
    void fetchData();
    return () => controller.abort();
  }, [selectedCorreoId, obtenerToken]);

  const handleVerArchivo = (filename) => {
    if (!correoDetail) return;
    window.open(`${API}/storage/mails/${correoDetail.id_mensaje}/${filename}`, "_blank");
  };

  const getFilteredCorreos = () => {
    let result = [...correos];
    if (searchCorreos) {
      const q = searchCorreos.toLowerCase();
      result = result.filter(c => (c.asunto?.toLowerCase().includes(q)) || (c.remitente?.toLowerCase().includes(q)));
    }
    result.sort((a, b) => {
      if (sortCorreos === 'fecha_desc') return new Date(b.fecha_recepcion) - new Date(a.fecha_recepcion);
      if (sortCorreos === 'fecha_asc') return new Date(a.fecha_recepcion) - new Date(b.fecha_recepcion);
      if (sortCorreos === 'estado') return (a.estado_correo?.nombre || '').localeCompare(b.estado_correo?.nombre || '');
      return 0;
    });
    return result;
  };

  return {
    correos, selectedCorreoId, setSelectedCorreoId, correoDetail,
    loading, searchCorreos, setSearchCorreos, sortCorreos, setSortCorreos,
    handleVerArchivo, getFilteredCorreos
  };
}