import { useState } from "react";
import { generarExpedientePDF } from "../../../utils/expedientePdf.js";

export function usePdfExpediente() {
  const [generandoPdf, setGenerandoPdf] = useState(false);

  const handleDescargarExpediente = async (
    radicado,
    tareasFlujo,
    historialTrazabilidad,
    API,
    eventosDianProp,
    activosFijosProp
  ) => {
    if (!radicado) return;
    try {
      setGenerandoPdf(true);
      let radicadoActual = radicado;
      const token = localStorage.getItem("token") || "";

      // 1. Refrescar datos del radicado
      try {
        const res = await fetch(`${API}/documentoradicado/${radicado.id}?_t=${Date.now()}`, {
          cache: "no-store",
          headers: { Authorization: `Bearer ${token}` },
        });
        if (res.ok) {
          const actualizado = await res.json();
          if (actualizado?.id) radicadoActual = actualizado;
        }
      } catch (err) {
        console.warn("No se pudo actualizar el radicado antes de generar el expediente", err);
      }

      // 2. Obtener activos fijos
      let activosFijos = activosFijosProp;
      if (!activosFijos || activosFijos.length === 0) {
        activosFijos = radicadoActual.activos_fijos || [];
      }
      if (!activosFijos || activosFijos.length === 0) {
        try {
          const resAf = await fetch(`${API}/documentoradicado/${radicadoActual.id}/activos-fijos?_t=${Date.now()}`, {
            headers: { Authorization: `Bearer ${token}` },
          });
          if (resAf.ok) {
            const dataAf = await resAf.json();
            if (Array.isArray(dataAf)) activosFijos = dataAf;
          }
        } catch (err) {
          console.warn("No se pudieron cargar activos fijos para el expediente", err);
        }
      }

      // 3. Obtener eventos DIAN
      let eventosDianList = eventosDianProp;
      if (!eventosDianList || eventosDianList.length === 0) {
        try {
          const resEv = await fetch(`${API}/documentoradicado/${radicadoActual.id}/eventos-dian?_t=${Date.now()}`, {
            headers: { Authorization: `Bearer ${token}` },
          });
          if (resEv.ok) {
            const dataEv = await resEv.json();
            if (Array.isArray(dataEv)) eventosDianList = dataEv;
          }
        } catch (err) {
          console.warn("No se pudieron cargar eventos DIAN para el expediente", err);
        }
      }

      // 4. Anexos: Factura y anexos PDF
      // Priorizar el PDF de la factura al principio, seguido por los demás anexos
      const docNum = (radicadoActual.documento_comercial?.numero_documento || "").toLowerCase().trim();
      const archivosPdf = (radicadoActual.archivos || [])
        .filter((a) => a.extension?.toLowerCase() === "pdf" || a.nombre?.toLowerCase().endsWith(".pdf"))
        .sort((a, b) => {
          const aName = (a.nombre || "").toLowerCase();
          const bName = (b.nombre || "").toLowerCase();
          const aIsFactura = aName.includes("factura") || (docNum && aName.includes(docNum));
          const bIsFactura = bName.includes("factura") || (docNum && bName.includes(docNum));
          if (aIsFactura && !bIsFactura) return -1;
          if (!aIsFactura && bIsFactura) return 1;
          return (a.id || 0) - (b.id || 0);
        });

      const anexosUrls = archivosPdf.map(
        (a) => `${API}/archivo/${a.id}/download?download=1&_t=${Date.now()}`
      );

      // 5. Generar y descargar el expediente completo
      await generarExpedientePDF({
        radicado: radicadoActual,
        flujo: tareasFlujo,
        trazabilidad: historialTrazabilidad,
        anexosUrls,
        activosFijos,
        eventosDian: eventosDianList,
      });
    } catch (err) {
      console.error("Error al generar expediente", err);
      alert("Hubo un error al generar el expediente PDF.");
    } finally {
      setGenerandoPdf(false);
    }
  };

  return { generandoPdf, handleDescargarExpediente };
}