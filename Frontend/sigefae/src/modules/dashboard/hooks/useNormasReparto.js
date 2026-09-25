import { useState } from "react";
import { API } from "../constants/api";

export function useNormasReparto(
  obtenerToken, esAdmin, puedeGestionarRecurso,
  normasRepartoRadicado, setNormasRepartoRadicado,
  externalActivosFijos, setExternalActivosFijos
) {
  const [showNormaModal, setShowNormaModal] = useState(false);
  const [normaEditandoId, setNormaEditandoId] = useState(null);
  const [normaFormDetalle, setNormaFormDetalle] = useState({ norma_reparto_id: "", porcentaje: "", proyecto: "", descripcion: "" });
  const [normaModalRadicadoId, setNormaModalRadicadoId] = useState(null);
  const [normaFiltroSede, setNormaFiltroSede] = useState("");
  const [normaFiltroArea, setNormaFiltroArea] = useState("");
  const [normasRepartoCatalogo, setNormasRepartoCatalogo] = useState([]);
  const [proyectosDisponibles, setProyectosDisponibles] = useState([]);

  // Activos Fijos
  const [showActivoModal, setShowActivoModal] = useState(false);
  const [activoEditandoId, setActivoEditandoId] = useState(null);
  const [activoFormDetalle, setActivoFormDetalle] = useState({ nombre: "", sucursal: "", proyecto: "", porcentaje: "", descripcion: "" });
  const [activoModalRadicadoId, setActivoModalRadicadoId] = useState(null);
  const [internalActivosFijos, setInternalActivosFijos] = useState([]);
  const activosFijosRadicado = externalActivosFijos !== undefined ? externalActivosFijos : internalActivosFijos;
  const setActivosFijosRadicado = setExternalActivosFijos || setInternalActivosFijos;

  const sedesDisponibles = ["BUCARAMANGA", "MALAMBO", "CUCUTA", "CB", "CIENAGA DE ORO", "GENERAL"];
  const areasDisponibles = ["ADMON", "VENTAS", "PRODUCCION"];

  const fetchProyectos = () => {
    if (proyectosDisponibles.length === 0) {
      fetch(`${API}/proyectos?activo=true&_t=${new Date().getTime()}`, { headers: { Authorization: `Bearer ${obtenerToken()}` } })
        .then(r => r.json()).then(d => setProyectosDisponibles(Array.isArray(d) ? d : []));
    }
  };

  const openNormaModal = (radicadoId, normaExistente = null) => {
    setNormaModalRadicadoId(radicadoId);
    if (normaExistente) {
      setNormaEditandoId(normaExistente.id);
      setNormaFormDetalle({
        norma_reparto_id: String(normaExistente.norma_reparto_id || normaExistente.norma_reparto?.id || ""),
        porcentaje: String(normaExistente.porcentaje || ""),
        proyecto: normaExistente.proyecto || "",
        descripcion: normaExistente.descripcion || ""
      });
      const norma = normasRepartoCatalogo.find(n => n.id === (normaExistente.norma_reparto_id || normaExistente.norma_reparto?.id));
      if (norma) { setNormaFiltroSede(norma.sucursal || ""); setNormaFiltroArea(norma.departamento || ""); }
    } else {
      setNormaEditandoId(null); setNormaFormDetalle({ norma_reparto_id: "", porcentaje: "", proyecto: "", descripcion: "" });
      setNormaFiltroSede(""); setNormaFiltroArea("");
    }
    if (normasRepartoCatalogo.length === 0) {
      fetch(`${API}/normas-reparto?activo=true&_t=${new Date().getTime()}`, { headers: { Authorization: `Bearer ${obtenerToken()}` } })
        .then(r => r.json()).then(d => setNormasRepartoCatalogo(Array.isArray(d) ? d : []));
    }
    fetchProyectos();
    setShowNormaModal(true);
  };

  const openActivoModal = (radicadoId, activoExistente = null) => {
    setActivoModalRadicadoId(radicadoId);
    if (activoExistente) {
      setActivoEditandoId(activoExistente.id);
      setActivoFormDetalle({
        nombre: activoExistente.nombre || "",
        sucursal: activoExistente.sucursal || "",
        proyecto: activoExistente.proyecto || "",
        porcentaje: String(activoExistente.porcentaje || ""),
        descripcion: activoExistente.descripcion || ""
      });
    } else {
      setActivoEditandoId(null);
      setActivoFormDetalle({ nombre: "", sucursal: "", proyecto: "", porcentaje: "", descripcion: "" });
    }
    fetchProyectos();
    setShowActivoModal(true);
  };

  const recargarActivosFijos = async (radicadoId) => {
    const res = await fetch(`${API}/documentoradicado/${radicadoId}/activos-fijos?_t=${new Date().getTime()}`, { headers: { Authorization: `Bearer ${obtenerToken()}` } });
    const data = await res.json();
    setActivosFijosRadicado(Array.isArray(data) ? data : []);
  };

  const handleGuardarActivoFijo = async (radicadoId) => {
    if (!activoFormDetalle.nombre?.trim()) { alert("El nombre del activo fijo es obligatorio"); return; }
    if (!activoFormDetalle.sucursal) { alert("Seleccione una sucursal"); return; }
    if (!activoFormDetalle.porcentaje || parseFloat(activoFormDetalle.porcentaje) <= 0) { alert("Ingrese un porcentaje válido"); return; }
    if (activoFormDetalle.proyecto?.trim()) {
      const proyectoValido = proyectosDisponibles.some(p => p.nombre.toLowerCase() === activoFormDetalle.proyecto.trim().toLowerCase());
      if (!proyectoValido) { alert("El proyecto ingresado no es válido. Debe seleccionar uno del catálogo."); return; }
    }
    try {
      const payload = {
        id: activoEditandoId || undefined,
        nombre: activoFormDetalle.nombre.trim(),
        sucursal: activoFormDetalle.sucursal,
        proyecto: activoFormDetalle.proyecto?.trim() || "",
        porcentaje: parseFloat(activoFormDetalle.porcentaje),
        descripcion: activoFormDetalle.descripcion?.trim() || ""
      };

      // Construir lista final
      let nuevosActivos;
      if (activoEditandoId) {
        nuevosActivos = activosFijosRadicado.map(a => a.id === activoEditandoId ? payload : {
          id: a.id, nombre: a.nombre, sucursal: a.sucursal, proyecto: a.proyecto || "",
          porcentaje: parseFloat(a.porcentaje), descripcion: a.descripcion || ""
        });
        if (!activosFijosRadicado.some(a => a.id === activoEditandoId)) nuevosActivos.push(payload);
      } else {
        nuevosActivos = [
          ...activosFijosRadicado.map(a => ({
            id: a.id, nombre: a.nombre, sucursal: a.sucursal, proyecto: a.proyecto || "",
            porcentaje: parseFloat(a.porcentaje), descripcion: a.descripcion || ""
          })),
          payload
        ];
      }

      // Validar que total normas + total activos no supere el 100%
      const totalNormas = (normasRepartoRadicado || []).reduce((sum, n) => sum + (parseFloat(n.porcentaje) || 0), 0);
      const totalActivos = nuevosActivos.reduce((sum, a) => sum + (parseFloat(a.porcentaje) || 0), 0);
      const granTotal = totalNormas + totalActivos;
      if (granTotal > 100.01) {
        alert(`La suma total de normas de reparto (${totalNormas.toFixed(2)}%) y activos fijos (${totalActivos.toFixed(2)}%) no puede superar el 100% total. Sumaría ${granTotal.toFixed(2)}%.`);
        return;
      }

      const res = await fetch(`${API}/documentoradicado/${radicadoId}/activos-fijos`, {
        method: "POST", headers: { "Content-Type": "application/json", Authorization: `Bearer ${obtenerToken()}` },
        body: JSON.stringify({ activos_fijos: nuevosActivos })
      });
      if (!res.ok) { const err = await res.json(); throw new Error(err.error || "Error guardando activo fijo"); }
      alert(activoEditandoId ? "Activo fijo actualizado" : "Activo fijo agregado");
      setShowActivoModal(false); setActivoEditandoId(null);
      await recargarActivosFijos(radicadoId);
    } catch (err) { alert("Error: " + err.message); }
  };

  const handleEliminarActivoFijo = async (activo, radicadoId) => {
    const creadorId = Number(activo.creado_por_id || activo.creado_por?.id || 0);
    if (!esAdmin && !puedeGestionarRecurso(creadorId)) { alert("No tienes permisos para eliminar este activo fijo."); return; }
    if (!confirm("¿Está seguro de eliminar este activo fijo?")) return;
    try {
      const restantes = activosFijosRadicado.filter(a => a.id !== activo.id).map(a => ({
        id: a.id, nombre: a.nombre, sucursal: a.sucursal, proyecto: a.proyecto || "",
        porcentaje: parseFloat(a.porcentaje), descripcion: a.descripcion || ""
      }));
      const res = await fetch(`${API}/documentoradicado/${radicadoId}/activos-fijos`, {
        method: "POST", headers: { "Content-Type": "application/json", Authorization: `Bearer ${obtenerToken()}` },
        body: JSON.stringify({ activos_fijos: restantes })
      });
      if (!res.ok) { const err = await res.json(); throw new Error(err.error || "Error eliminando activo fijo"); }
      alert("Activo fijo eliminado");
      await recargarActivosFijos(radicadoId);
    } catch (err) { alert("Error: " + err.message); }
  };

  const handleGuardarNormaDetalle = async (radicadoId) => {
    if (!normaFormDetalle.norma_reparto_id || normaFormDetalle.porcentaje === "") { alert("Seleccione una norma y el porcentaje"); return; }
    if (!normaFormDetalle.proyecto || !normaFormDetalle.proyecto.trim()) { alert("El campo Proyecto es obligatorio"); return; }
    
    // Validar que el proyecto esté en el catálogo
    const proyectoValido = proyectosDisponibles.some(p => p.nombre.toLowerCase() === normaFormDetalle.proyecto.trim().toLowerCase());
    if (!proyectoValido) {
      alert("El proyecto ingresado no es válido. Debe seleccionar uno de la lista del catálogo.");
      return;
    }

    try {
      // Norma editada / nueva
      const payloadNorma = {
        norma_reparto_id: parseInt(normaFormDetalle.norma_reparto_id),
        porcentaje: parseFloat(normaFormDetalle.porcentaje),
        proyecto: normaFormDetalle.proyecto.trim(),
        descripcion: normaFormDetalle.descripcion?.trim() || ""
      };

      // Construir lista final preservando datos de las normas no editadas
      let nuevasNormas;
      if (normaEditandoId) {
        nuevasNormas = normasRepartoRadicado.map(n => {
          const nId = n.id;
          if (nId === normaEditandoId) {
            // Reemplazar con datos del form
            return payloadNorma;
          }
          // Preservar datos existentes exactamente
          return {
            norma_reparto_id: n.norma_reparto_id || n.norma_reparto?.id,
            porcentaje: parseFloat(n.porcentaje),
            proyecto: n.proyecto || "",
            descripcion: n.descripcion || ""
          };
        });
        // Si no encontró por id, agregar al final (fallback)
        const encontrado = normasRepartoRadicado.some(n => n.id === normaEditandoId);
        if (!encontrado) nuevasNormas.push(payloadNorma);
      } else {
        // Nueva norma: verificar duplicado
        const yaExiste = normasRepartoRadicado.find(n =>
          (n.norma_reparto_id || n.norma_reparto?.id) === payloadNorma.norma_reparto_id
        );
        if (yaExiste) { alert("Esta norma ya fue agregada"); return; }
        nuevasNormas = [
          ...normasRepartoRadicado.map(n => ({
            norma_reparto_id: n.norma_reparto_id || n.norma_reparto?.id,
            porcentaje: parseFloat(n.porcentaje),
            proyecto: n.proyecto || "",
            descripcion: n.descripcion || ""
          })),
          payloadNorma
        ];
      }

      // Validar que total normas + total activos no supere el 100%
      const totalNormas = nuevasNormas.reduce((sum, n) => sum + (parseFloat(n.porcentaje) || 0), 0);
      const totalActivos = (activosFijosRadicado || []).reduce((sum, a) => sum + (parseFloat(a.porcentaje) || 0), 0);
      const granTotal = totalNormas + totalActivos;
      if (granTotal > 100.01) {
        alert(`La suma total de normas de reparto (${totalNormas.toFixed(2)}%) y activos fijos (${totalActivos.toFixed(2)}%) no puede superar el 100% total. Sumaría ${granTotal.toFixed(2)}%.`);
        return;
      }

      const res = await fetch(`${API}/documentoradicado/${radicadoId}/normas-reparto`, {
        method: "POST", headers: { "Content-Type": "application/json", Authorization: `Bearer ${obtenerToken()}` },
        body: JSON.stringify({ normas: nuevasNormas })
      });
      if (!res.ok) { const err = await res.json(); throw new Error(err.error || "Error guardando norma"); }
      alert(normaEditandoId ? "Norma actualizada correctamente" : "Norma agregada correctamente");
      setShowNormaModal(false); setNormaEditandoId(null);
      const nrRes = await fetch(`${API}/documentoradicado/${radicadoId}/normas-reparto?_t=${new Date().getTime()}`, { headers: { Authorization: `Bearer ${obtenerToken()}` } });
      const nrData = await nrRes.json();
      setNormasRepartoRadicado(Array.isArray(nrData) ? nrData : []);
    } catch (err) { alert("Error: " + err.message); }
  };

  const handleEliminarNorma = async (asignacion, radicadoId) => {
    const norma = asignacion || {};
    const creadorId = Number(norma.creado_por_id || norma.creado_por?.id || 0);
    if (!esAdmin && !puedeGestionarRecurso(creadorId)) { alert("No tienes permisos para eliminar esta norma de reparto."); return; }
    if (!confirm("¿Está seguro de eliminar esta norma de reparto?")) return;
    try {
      const normasRestantes = normasRepartoRadicado.filter(n => n.id !== norma.id).map(n => ({
        norma_reparto_id: n.norma_reparto_id || n.norma_reparto?.id,
        porcentaje: parseFloat(n.porcentaje),
        proyecto: n.proyecto || "",
        descripcion: n.descripcion || ""
      }));
      const res = await fetch(`${API}/documentoradicado/${radicadoId}/normas-reparto`, {
        method: "POST", headers: { "Content-Type": "application/json", Authorization: `Bearer ${obtenerToken()}` },
        body: JSON.stringify({ normas: normasRestantes })
      });
      if (!res.ok) { const err = await res.json(); throw new Error(err.error || "Error eliminando norma"); }
      alert("Norma eliminada");
      const nrRes = await fetch(`${API}/documentoradicado/${radicadoId}/normas-reparto?_t=${new Date().getTime()}`, { headers: { Authorization: `Bearer ${obtenerToken()}` } });
      const nrData = await nrRes.json();
      setNormasRepartoRadicado(Array.isArray(nrData) ? nrData : []);
    } catch (err) { alert("Error: " + err.message); }
  };

  return {
    showNormaModal, setShowNormaModal, normaEditandoId, setNormaEditandoId,
    normaFormDetalle, setNormaFormDetalle, normaModalRadicadoId,
    normaFiltroSede, setNormaFiltroSede, normaFiltroArea, setNormaFiltroArea,
    normasRepartoCatalogo, setNormasRepartoCatalogo, proyectosDisponibles,
    sedesDisponibles, areasDisponibles,
    openNormaModal, handleGuardarNormaDetalle, handleEliminarNorma,
    // Activos Fijos
    showActivoModal, setShowActivoModal, activoEditandoId, activoFormDetalle, setActivoFormDetalle,
    activoModalRadicadoId, activosFijosRadicado, setActivosFijosRadicado,
    openActivoModal, handleGuardarActivoFijo, handleEliminarActivoFijo, recargarActivosFijos
  };
}