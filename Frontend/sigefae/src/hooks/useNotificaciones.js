// useNotificaciones.js
import { useState, useEffect, useCallback, useRef } from "react";
import { obtenerToken } from "../modules/auth/token.js";
import {
  playNotificacionSound,
  mostrarNotificacionNativa,
  pedirPermisoNotificaciones,
  inicializarAudio,
} from "../utils/notificacion.js";

import { API } from "../modules/dashboard/constants/api";

export function useNotificaciones() {
  const [notificaciones, setNotificaciones] = useState([]);
  const [noLeidas, setNoLeidas] = useState(0);
  const [loading, setLoading] = useState(false);
  const idsVistosRef = useRef(new Set());
  const primeraCargaRef = useRef(true);

  const cargar = useCallback(async () => {
    setLoading(true);
    try {
      const res = await fetch(`${API}/notificacion/mias?_t=${new Date().getTime()}`, {
        headers: { Authorization: `Bearer ${obtenerToken()}` },
        cache: 'no-store' // Evitar caché del navegador
      });
      const data = await res.json();
      if (Array.isArray(data)) {
        const pendientes = data.filter((n) => n.estado === "Pendiente");

        if (!primeraCargaRef.current) {
          const nuevas = pendientes.filter((n) => !idsVistosRef.current.has(n.id));
          if (nuevas.length > 0) {
            playNotificacionSound();
            const ultima = nuevas[0];
            mostrarNotificacionNativa("SIGEFAE", ultima.mensaje, () => {
              window.dispatchEvent(
                new CustomEvent("navegar-a-radicado", {
                  detail: ultima.documento_radicado_id,
                })
              );
            });
            // Emitir evento in-app para toasts (Admin/visual)
            try {
              window.dispatchEvent(new CustomEvent('notificacion-nueva', { detail: ultima }));
            } catch (e) { /* ignore */ }
          }
        }

        idsVistosRef.current = new Set(pendientes.map((n) => n.id));
        primeraCargaRef.current = false;
        setNotificaciones(data);
        setNoLeidas(pendientes.length);
      }
    } catch (err) {
      console.error("Error cargando notificaciones:", err);
    } finally {
      setLoading(false);
    }
  }, []);

	const marcarLeida = async (id) => {
		setNotificaciones((prev) =>
			prev.map((n) => (n.id === id ? { ...n, estado: "Leida" } : n))
		);
		setNoLeidas((prev) => Math.max(0, prev - 1));
		try {
			const res = await fetch(`${API}/notificacion/${id}/leida`, {
				method: "PATCH",
				headers: { Authorization: `Bearer ${obtenerToken()}` },
			});
			if (!res.ok) cargar();
		} catch (err) {
			console.error(err);
			cargar();
		}
	};

	const borrarNotificacion = async (id) => {
		idsVistosRef.current.delete(id);
		setNotificaciones((prev) => {
			const target = prev.find((n) => n.id === id);
			if (target && target.estado === "Pendiente") {
				setNoLeidas((nl) => Math.max(0, nl - 1));
			}
			return prev.filter((n) => n.id !== id);
		});
		try {
			const res = await fetch(`${API}/notificacion/${id}`, {
				method: "DELETE",
				headers: { Authorization: `Bearer ${obtenerToken()}` },
			});
			if (!res.ok) cargar();
		} catch (err) {
			console.error(err);
			cargar();
		}
	};

  useEffect(() => {
    pedirPermisoNotificaciones();
    inicializarAudio();
  }, []);

  useEffect(() => {
    cargar();
    const interval = setInterval(cargar, 30000);
    return () => clearInterval(interval);
  }, [cargar]);

	return { notificaciones, noLeidas, loading, marcarLeida, borrarNotificacion, recargar: cargar };
}