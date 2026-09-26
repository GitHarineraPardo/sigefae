import { useState, useEffect } from "react";
import { API } from "../constants/api";

export function useRadicacion(obtenerToken, userId, setDocumentos, setActiveTab, setSelectedDocId, setDocDetail) {
  const [showRadicarModal, setShowRadicarModal] = useState(false);
  const [radicarDocId, setRadicarDocId] = useState(null);
  const [radicarForm, setRadicarForm] = useState({
    tipo_radicacion_id: "",
    ruta_id: "",
    tipo_pago_id: "",
    metodo_pago_id: "",
    numero_radicado: "",
    es_malambo: false,
    normas_reparto: [],
    activos_fijos: []
  });
  const [radicando, setRadicando] = useState(false);
  const [normasRepartoAutoMsg, setNormasRepartoAutoMsg] = useState("");

  const [tiposRadicacion, setTiposRadicacion] = useState([]);
  const [rutas, setRutas] = useState([]);
  const [tiposPago, setTiposPago] = useState([]);
  const [metodosPago, setMetodosPago] = useState([]);
  const [normasRepartoCatalogo, setNormasRepartoCatalogo] = useState([]);
  const [proyectosCatalogo, setProyectosCatalogo] = useState([]);

  // Inputs para Normas de Reparto
  const [normaFiltroSede, setNormaFiltroSede] = useState("");
  const [normaFiltroArea, setNormaFiltroArea] = useState("");
  const [normaSeleccionadaId, setNormaSeleccionadaId] = useState("");
  const [normaPorcentajeInput, setNormaPorcentajeInput] = useState("");
  const [normaValorInput, setNormaValorInput] = useState("");
  const [normaProyectoInput, setNormaProyectoInput] = useState("");
  const [normaDescripcionInput, setNormaDescripcionInput] = useState("");

  // Inputs para Activos Fijos
  const [activoNombreInput, setActivoNombreInput] = useState("");
  const [activoSucursalInput, setActivoSucursalInput] = useState("");
  const [activoProyectoInput, setActivoProyectoInput] = useState("");
  const [activoPorcentajeInput, setActivoPorcentajeInput] = useState("");
  const [activoValorInput, setActivoValorInput] = useState("");
  const [activoDescripcionInput, setActivoDescripcionInput] = useState("");

  const [subtotalDoc, setSubtotalDoc] = useState(0);

  // normas predeterminadas del proveedor-ruta
  const [proveedorIdActual, setProveedorIdActual] = useState(null);
  const [normasPredeterminadas, setNormasPredeterminadas] = useState([]);
  const [usarNormasPredeterminadas, setUsarNormasPredeterminadas] = useState(null);

  const sedesDisponibles = ["BUCARAMANGA", "MALAMBO", "CUCUTA", "CB", "CIENAGA DE ORO", "GENERAL"];
  const areasDisponibles = ["ADMON", "VENTAS", "PRODUCCION"];

  useEffect(() => {
    if (!showRadicarModal) return;
    const headers = { Authorization: `Bearer ${obtenerToken()}` };
    const t = `_t=${new Date().getTime()}`;
    Promise.all([
      fetch(`${API}/tipo-radicacion?${t}`, { headers }).then(r => r.json()),
      fetch(`${API}/rutas?${t}`, { headers }).then(r => r.json()),
      fetch(`${API}/tipos-pago?${t}`, { headers }).then(r => r.json()),
      fetch(`${API}/metodos-pago?${t}`, { headers }).then(r => r.json()),
      fetch(`${API}/normas-reparto?activo=true&${t}`, { headers }).then(r => r.json()),
      fetch(`${API}/proyectos?activo=true&${t}`, { headers }).then(r => r.json()),
    ]).then(([tr, r, tp, mp, nr, proys]) => {
      setTiposRadicacion(Array.isArray(tr) ? tr : []);
      setRutas(Array.isArray(r) ? r : []);
      setTiposPago(Array.isArray(tp) ? tp : []);
      setMetodosPago(Array.isArray(mp) ? mp : []);
      setNormasRepartoCatalogo(Array.isArray(nr) ? nr : []);
      setProyectosCatalogo(Array.isArray(proys) ? proys : []);
    }).catch(err => console.error("Error cargando catálogos:", err));
  }, [showRadicarModal, obtenerToken]);

  // normas predeterminadas por proveedor y ruta
  useEffect(() => {
    if (!showRadicarModal || !proveedorIdActual || !radicarForm.ruta_id) {
      setNormasPredeterminadas([]);
      setUsarNormasPredeterminadas(null);
      return;
    }
    const cargar = async () => {
      try {
        const res = await fetch(`${API}/proveedor/${proveedorIdActual}/normas-reparto?ruta_id=${radicarForm.ruta_id}`, {
          headers: { Authorization: `Bearer ${obtenerToken()}` }
        });
        const data = await res.json();
        if (Array.isArray(data) && data.length > 0) {
          setNormasPredeterminadas(data);
          setUsarNormasPredeterminadas(null);
        } else {
          setNormasPredeterminadas([]);
          setUsarNormasPredeterminadas(null);
        }
      } catch (e) {
        console.error("Error cargando normas predeterminadas:", e);
        setNormasPredeterminadas([]);
      }
    };
    cargar();
  }, [radicarForm.ruta_id, proveedorIdActual, showRadicarModal, obtenerToken]);

  const openRadicarModal = async (docId, subtotal = 0) => {
    setRadicarDocId(docId);
    setSubtotalDoc(parseFloat(subtotal) || 0);
    setRadicarForm({
      tipo_radicacion_id: "", ruta_id: "", tipo_pago_id: "", metodo_pago_id: "",
      numero_radicado: "", es_malambo: false, normas_reparto: [], activos_fijos: []
    });
    setNormasRepartoAutoMsg("");
    setNormasPredeterminadas([]);
    setUsarNormasPredeterminadas(null);
    setNormaFiltroSede(""); setNormaFiltroArea("");
    setNormaSeleccionadaId(""); setNormaPorcentajeInput(""); setNormaValorInput("");
    setNormaProyectoInput(""); setNormaDescripcionInput("");
    setActivoNombreInput(""); setActivoSucursalInput(""); setActivoProyectoInput("");
    setActivoPorcentajeInput(""); setActivoValorInput(""); setActivoDescripcionInput("");

    try {
      const res = await fetch(`${API}/documentocomercial/${docId}`, {
        headers: { Authorization: `Bearer ${obtenerToken()}` }
      });
      const data = await res.json();
      if (data?.proveedor_id) {
        setProveedorIdActual(data.proveedor_id);
      } else if (data?.proveedor?.id) {
        setProveedorIdActual(data.proveedor.id);
      } else {
        setProveedorIdActual(null);
      }
    } catch (e) {
      console.error("Error obteniendo proveedor del documento:", e);
      setProveedorIdActual(null);
    }

    setShowRadicarModal(true);
  };

  const aceptarNormasPredeterminadas = () => {
    const nuevasNormas = normasPredeterminadas.map(n => ({
      norma_reparto_id: String(n.norma_reparto_id),
      porcentaje: n.porcentaje.toFixed(2),
      valor: subtotalDoc > 0 ? ((n.porcentaje / 100) * subtotalDoc).toFixed(2) : "0",
      proyecto: "",
      descripcion: ""
    }));
    setRadicarForm(prev => ({ ...prev, normas_reparto: nuevasNormas }));
    setUsarNormasPredeterminadas(true);
    setNormasRepartoAutoMsg("");
  };

  const rechazarNormasPredeterminadas = () => {
    setRadicarForm(prev => ({ ...prev, normas_reparto: [] }));
    setUsarNormasPredeterminadas(false);
    setNormasRepartoAutoMsg("");
  };

  const handleRadicarChange = (e) => {
    const { name, value } = e.target;
    setRadicarForm(prev => ({ ...prev, [name]: value }));
  };

  const totalPorcentajeNormas = (radicarForm.normas_reparto || []).reduce((sum, n) => sum + (parseFloat(n.porcentaje) || 0), 0);
  const totalPorcentajeActivos = (radicarForm.activos_fijos || []).reduce((sum, a) => sum + (parseFloat(a.porcentaje) || 0), 0);
  const totalPorcentajeCombinado = totalPorcentajeNormas + totalPorcentajeActivos;

  const handleAgregarNormaModal = () => {
    if (!normaSeleccionadaId) { alert("Selecciona una norma"); return; }
    if (!normaPorcentajeInput && !normaValorInput) { alert("Ingresa un porcentaje o un valor"); return; }
    if (!normaProyectoInput || !normaProyectoInput.trim()) { alert("El campo Proyecto es obligatorio"); return; }

    let pct = parseFloat(normaPorcentajeInput);
    let val = parseFloat(normaValorInput);

    if (subtotalDoc > 0) {
      if (normaValorInput && !normaPorcentajeInput) {
        pct = (val / subtotalDoc) * 100;
      } else if (normaPorcentajeInput && !normaValorInput) {
        val = (pct / 100) * subtotalDoc;
      }
    }

    if (isNaN(pct) || pct <= 0) { alert("Porcentaje inválido"); return; }
    if (subtotalDoc > 0 && val > subtotalDoc) { alert("El valor no puede superar el subtotal"); return; }

    // VALIDACIÓN ESTRICTA DE MÁXIMO 100% COMBINADO
    if (totalPorcentajeCombinado + pct > 100.01) {
      alert(`No se puede superar el 100% total. Actualmente la suma es ${totalPorcentajeCombinado.toFixed(2)}% y estás intentando agregar ${pct.toFixed(2)}%.`);
      return;
    }

    const yaExiste = radicarForm.normas_reparto.find(n => n.norma_reparto_id === normaSeleccionadaId);
    if (yaExiste) { alert("Esta norma ya fue agregada"); return; }

    setRadicarForm(prev => ({
      ...prev,
      normas_reparto: [
        ...prev.normas_reparto,
        {
          norma_reparto_id: normaSeleccionadaId,
          porcentaje: pct.toFixed(2),
          valor: val.toFixed(2),
          proyecto: normaProyectoInput.trim(),
          descripcion: normaDescripcionInput.trim()
        }
      ]
    }));

    setNormaSeleccionadaId(""); setNormaPorcentajeInput(""); setNormaValorInput("");
    setNormaProyectoInput(""); setNormaDescripcionInput("");
  };

  const handleNormaRepartoChange = (index, field, value) => {
    setRadicarForm(prev => {
      const updated = [...prev.normas_reparto];
      updated[index] = { ...updated[index], [field]: value };
      return { ...prev, normas_reparto: updated };
    });
  };

  const handleRemoveNormaReparto = (index) => {
    setRadicarForm(prev => ({ ...prev, normas_reparto: prev.normas_reparto.filter((_, i) => i !== index) }));
  };

  // ── MANEJO DE ACTIVOS FIJOS EN RADICACIÓN ──
  const handleAgregarActivoModal = () => {
    if (!activoNombreInput?.trim()) { alert("Ingresa el nombre del activo fijo"); return; }
    if (!activoSucursalInput) { alert("Selecciona la sucursal del activo"); return; }
    if (!activoPorcentajeInput && !activoValorInput) { alert("Ingresa un porcentaje o un valor"); return; }

    let pct = parseFloat(activoPorcentajeInput);
    let val = parseFloat(activoValorInput);

    if (subtotalDoc > 0) {
      if (activoValorInput && !activoPorcentajeInput) {
        pct = (val / subtotalDoc) * 100;
      } else if (activoPorcentajeInput && !activoValorInput) {
        val = (pct / 100) * subtotalDoc;
      }
    }

    if (isNaN(pct) || pct <= 0) { alert("Porcentaje de activo fijo inválido"); return; }

    // VALIDACIÓN ESTRICTA DE MÁXIMO 100% COMBINADO
    if (totalPorcentajeCombinado + pct > 100.01) {
      alert(`No se puede superar el 100% total. Actualmente la suma es ${totalPorcentajeCombinado.toFixed(2)}% y estás intentando agregar ${pct.toFixed(2)}%.`);
      return;
    }

    setRadicarForm(prev => ({
      ...prev,
      activos_fijos: [
        ...prev.activos_fijos,
        {
          nombre: activoNombreInput.trim(),
          sucursal: activoSucursalInput,
          proyecto: activoProyectoInput.trim(),
          porcentaje: pct.toFixed(2),
          valor: val.toFixed(2),
          descripcion: activoDescripcionInput.trim()
        }
      ]
    }));

    setActivoNombreInput(""); setActivoSucursalInput(""); setActivoProyectoInput("");
    setActivoPorcentajeInput(""); setActivoValorInput(""); setActivoDescripcionInput("");
  };

  const handleActivoFijoChange = (index, field, value) => {
    setRadicarForm(prev => {
      const updated = [...prev.activos_fijos];
      updated[index] = { ...updated[index], [field]: value };
      return { ...prev, activos_fijos: updated };
    });
  };

  const handleRemoveActivoFijo = (index) => {
    setRadicarForm(prev => ({ ...prev, activos_fijos: prev.activos_fijos.filter((_, i) => i !== index) }));
  };

  const normasFiltradas = normasRepartoCatalogo.filter(n => {
    if (normaFiltroSede && n.sucursal !== normaFiltroSede) return false;
    if (normaFiltroArea && n.departamento !== normaFiltroArea) return false;
    return true;
  });

  const handleRadicarSubmit = async () => {
    if (!radicarDocId) return;
    if (!radicarForm.tipo_radicacion_id || !radicarForm.ruta_id || !radicarForm.metodo_pago_id) {
      alert("Debe seleccionar tipo de radicación, ruta y método de pago."); return;
    }

    if (totalPorcentajeCombinado > 100.01) {
      alert(`El total asignado (${totalPorcentajeCombinado.toFixed(2)}%) no puede superar el 100%.`);
      return;
    }

    const rutaSeleccionada = rutas.find(r => String(r.id) === String(radicarForm.ruta_id));
    const es_malambo_derivado = rutaSeleccionada ? rutaSeleccionada.zona === "MALAMBO" : false;

    setRadicando(true);
    const payload = {
      documento_comercial_id: radicarDocId,
      tipo_radicacion_id: parseInt(radicarForm.tipo_radicacion_id),
      ruta_id: parseInt(radicarForm.ruta_id),
      metodo_pago_id: parseInt(radicarForm.metodo_pago_id),
      numero_radicado: radicarForm.numero_radicado?.trim() || "",
      es_malambo: es_malambo_derivado,
      normas_reparto: (radicarForm.normas_reparto || []).filter(n => n.norma_reparto_id && n.porcentaje).map(n => ({
        norma_reparto_id: parseInt(n.norma_reparto_id),
        porcentaje: parseFloat(n.porcentaje),
        proyecto: n.proyecto || "",
        descripcion: n.descripcion || ""
      })),
      activos_fijos: (radicarForm.activos_fijos || []).filter(a => a.nombre && a.porcentaje).map(a => ({
        nombre: a.nombre,
        sucursal: a.sucursal,
        proyecto: a.proyecto || "",
        porcentaje: parseFloat(a.porcentaje),
        descripcion: a.descripcion || ""
      }))
    };
    try {
      const res = await fetch(`${API}/documentoradicado`, {
        method: "POST", headers: { "Content-Type": "application/json", Authorization: `Bearer ${obtenerToken()}` },
        body: JSON.stringify(payload),
      });
      if (!res.ok) { const errData = await res.json(); throw new Error(errData.error || "Error al radicar"); }
      const creado = await res.json();

      setShowRadicarModal(false); setRadicarDocId(null);
      setSelectedDocId(null); setDocDetail(null);
      setDocumentos(prev => prev.filter(d => d.id !== radicarDocId));
      setActiveTab("radicados");
    } catch (err) { alert("Error al radicar: " + err.message); }
    finally { setRadicando(false); }
  };

  return {
    showRadicarModal, setShowRadicarModal, radicarDocId, setRadicarDocId,
    radicarForm, setRadicarForm, radicando, normasRepartoAutoMsg,
    tiposRadicacion, rutas, tiposPago, metodosPago, normasRepartoCatalogo, proyectosCatalogo,
    normaFiltroSede, setNormaFiltroSede, normaFiltroArea, setNormaFiltroArea,
    normaSeleccionadaId, setNormaSeleccionadaId, normaPorcentajeInput, setNormaPorcentajeInput,
    normaValorInput, setNormaValorInput, normaProyectoInput, setNormaProyectoInput,
    normaDescripcionInput, setNormaDescripcionInput,
    // Activos Fijos inputs
    activoNombreInput, setActivoNombreInput, activoSucursalInput, setActivoSucursalInput,
    activoProyectoInput, setActivoProyectoInput, activoPorcentajeInput, setActivoPorcentajeInput,
    activoValorInput, setActivoValorInput, activoDescripcionInput, setActivoDescripcionInput,
    subtotalDoc, sedesDisponibles, areasDisponibles, normasFiltradas,
    totalPorcentajeNormas, totalPorcentajeActivos, totalPorcentajeCombinado,
    proveedorIdActual, normasPredeterminadas, usarNormasPredeterminadas,
    aceptarNormasPredeterminadas, rechazarNormasPredeterminadas,
    openRadicarModal, handleRadicarChange, handleAgregarNormaModal,
    handleNormaRepartoChange, handleRemoveNormaReparto,
    handleAgregarActivoModal, handleActivoFijoChange, handleRemoveActivoFijo,
    handleRadicarSubmit
  };
}