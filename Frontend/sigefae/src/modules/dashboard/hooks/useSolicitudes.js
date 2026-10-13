import { useState, useEffect } from "react";
import { API } from "../constants/api";

export function useSolicitudes(obtenerToken, activeTab, esAdmin) {
  const [solicitudes, setSolicitudes] = useState([]);
  const [loadingSolicitudes, setLoadingSolicitudes] = useState(false);

  useEffect(() => {
    if (activeTab !== "solicitudes") return;
    const controller = new AbortController();
    const fetchData = async () => {
      setLoadingSolicitudes(true);
      const endpoint = esAdmin ? `${API}/solicitud-rechazo` : `${API}/solicitud-rechazo/mias`;
      try {
        const r = await fetch(endpoint, {
          headers: { Authorization: `Bearer ${obtenerToken()}` },
          signal: controller.signal,
        });
        const data = await r.json();
        setSolicitudes(Array.isArray(data) ? data : []);
      } catch (err) {
        if (!controller.signal.aborted) {
          console.error(err);
          setSolicitudes([]);
        }
      } finally {
        if (!controller.signal.aborted) setLoadingSolicitudes(false);
      }
    };
    void fetchData();
    return () => controller.abort();
  }, [activeTab, esAdmin, obtenerToken]);

  const handleDecidir = async (id, accept) => {
    if (!confirm(accept ? 'Aceptar solicitud y marcar como rechazado?' : 'Rechazar solicitud (no marcar documento)?')) return;
    try {
      const res = await fetch(`${API}/solicitud-rechazo/${id}/decidir`, {
        method: 'POST', headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${obtenerToken()}` },
        body: JSON.stringify({ accept, mensaje: '' })
      });
      if (!res.ok) throw new Error((await res.json()).error || 'Error');
      alert(accept ? 'Solicitud aceptada' : 'Solicitud rechazada');
      setSolicitudes(prev => prev.filter(x => x.id !== id));
    } catch (err) { alert('Error: ' + err.message); }
  };

  return { solicitudes, loadingSolicitudes, handleDecidir };
}