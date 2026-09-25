import { useState } from "react";
import { generarExpedientePDF } from "../../../utils/expedientePdf.js";

export function usePdfExpediente() {
  const [generandoPdf, setGenerandoPdf] = useState(false);

  const handleDescargarExpediente = async (radicado, tareasFlujo, historialTrazabilidad, API) => {
    if (!radicado) return;
    try {
      setGenerandoPdf(true);
      let radicadoActual = radicado;
      try {
        const res = await fetch(`${API}/documentoradicado/${radicado.id}?_t=${Date.now()}`, {
          cache: "no-store",
          headers: { Authorization: `Bearer ${localStorage.getItem("token") || ""}` },
        });
        if (res.ok) {
          const actualizado = await res.json();
          if (actualizado?.id) radicadoActual = actualizado;
        }
      } catch (err) {
        console.warn("No se pudo actualizar el radicado antes de generar el expediente", err);
      }

      const anexosUrls = (radicadoActual.archivos || [])
        .filter(a => a.extension?.toLowerCase() === 'pdf' || a.nombre?.toLowerCase().endsWith('.pdf'))
        .map(a => `${API}/archivo/${a.id}/download?download=1&_t=${Date.now()}`);
      await generarExpedientePDF(radicadoActual, tareasFlujo, historialTrazabilidad, anexosUrls);
    } catch (err) {
      console.error("Error al generar expediente", err);
      alert("Hubo un error al generar el expediente PDF.");
    } finally { setGenerandoPdf(false); }
  };

  return { generandoPdf, handleDescargarExpediente };
}