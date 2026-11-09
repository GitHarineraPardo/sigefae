import { useState, useEffect } from "react";
import { API } from "../constants/api";

export function useDocumentos(obtenerToken, activeTab) {
  const [documentos, setDocumentos] = useState([]);
  const [selectedDocId, setSelectedDocId] = useState(null);
  const [docDetail, setDocDetail] = useState(null);
  const [loadingDocs, setLoadingDocs] = useState(false);
  const [searchDocs, setSearchDocs] = useState("");
  const [sortDocs, setSortDocs] = useState("fecha_desc");
  const [syncingInbound, setSyncingInbound] = useState(false);
  const [syncResult, setSyncResult] = useState(null);

  // Crear documento manual
  const [showCrearDocModal, setShowCrearDocModal] = useState(false);
  const [docForm, setDocForm] = useState({
    tipo: "FACTURA_FISICA", numero_documento: "", id_proveedor: "",
    id_receptor: "", id_area: "", tipo_factura_id: "", asunto: "",
    fecha_documento: new Date().toISOString().split("T")[0],
    moneda_id: "", subtotal: 0, iva: 0, total: 0, detalles: []
  });
  const [creandoDoc, setCreandoDoc] = useState(false);

  // Crear proveedor rápido
  const [showCrearProveedorModal, setShowCrearProveedorModal] = useState(false);
  const [proveedorForm, setProveedorForm] = useState({
    razon_social: "", numero_documento: "", tipo_documento_id: "", email: "", telefono: ""
  });
  const [creandoProveedor, setCreandoProveedor] = useState(false);

  // Catálogos auxiliares
  const [proveedoresCatalogo, setProveedoresCatalogo] = useState([]);
  const [receptoresCatalogo, setReceptoresCatalogo] = useState([]);
  const [tiposDocumentoCatalogo, setTiposDocumentoCatalogo] = useState([]);
  const [areasCatalogo, setAreasCatalogo] = useState([]);
  const [monedasCatalogo, setMonedasCatalogo] = useState([]);

  useEffect(() => {
    if (activeTab !== "documentos") return;
    const headers = { Authorization: `Bearer ${obtenerToken()}` };
    const controller = new AbortController();
    const fetchData = async () => {
      setLoadingDocs(true);
      const t1 = `?_t=${new Date().getTime()}`;
      try {
        const res = await fetch(`${API}/documentocomercial/pendientes${t1}`, {
          headers,
          signal: controller.signal,
        });
        const data = await res.json();
        if (Array.isArray(data)) setDocumentos(data);
      } catch (err) {
        if (!controller.signal.aborted) console.error(err);
      } finally {
        if (!controller.signal.aborted) setLoadingDocs(false);
      }
    };
    void fetchData();

    const t2 = `_t=${new Date().getTime()}`;
    fetch(`${API}/proveedor?${t2}`, { headers }).then(r => r.json()).then(d => setProveedoresCatalogo(Array.isArray(d) ? d : [])).catch(() => setProveedoresCatalogo([]));
    fetch(`${API}/receptor?${t2}`, { headers }).then(r => r.json()).then(d => setReceptoresCatalogo(Array.isArray(d) ? d : [])).catch(() => setReceptoresCatalogo([]));
    if (areasCatalogo.length === 0) fetch(`${API}/areas?${t2}`, { headers }).then(r => r.json()).then(d => setAreasCatalogo(Array.isArray(d) ? d : [])).catch(() => {});
    if (monedasCatalogo.length === 0) fetch(`${API}/monedas?${t2}`, { headers }).then(r => r.json()).then(d => setMonedasCatalogo(Array.isArray(d) ? d : [])).catch(() => {});
    fetch(`${API}/tipo-documento?${t2}`, { headers }).then(r => r.json()).then(d => setTiposDocumentoCatalogo(Array.isArray(d) ? d : [])).catch(() => setTiposDocumentoCatalogo([]));

    return () => controller.abort();
  }, [activeTab, obtenerToken, areasCatalogo.length, monedasCatalogo.length]);

  useEffect(() => {
    if (!selectedDocId) return;
    const controller = new AbortController();
    const fetchDetail = async () => {
      setDocDetail(null);
      const t = `?_t=${new Date().getTime()}`;
      try {
        const res = await fetch(`${API}/documentocomercial/${selectedDocId}${t}`, {
          headers: { Authorization: `Bearer ${obtenerToken()}` },
          signal: controller.signal,
        });
        const data = await res.json();
        if (data?.id) setDocDetail(data);
      } catch (err) {
        if (!controller.signal.aborted) console.error(err);
      }
    };
    void fetchDetail();
    return () => controller.abort();
  }, [selectedDocId, obtenerToken]);

  const recalcularTotales = (detalles) => {
    const subtotal = detalles.reduce((s, d) => s + (parseFloat(d.valor_unitario) || 0) * (parseFloat(d.cantidad) || 0), 0);
    const iva = detalles.reduce((s, d) => s + (parseFloat(d.iva_unitario) || 0) * (parseFloat(d.cantidad) || 0), 0);
    return { subtotal, iva, total: subtotal + iva };
  };

  const handleDocFormChange = (e) => {
    const { name, value } = e.target;
    setDocForm(prev => ({ ...prev, [name]: value }));
  };

  const handleAddDetalle = () => {
    setDocForm(prev => {
      const detalles = [...prev.detalles, { descripcion: "", cantidad: 1, valor_unitario: 0, iva_unitario: 0, total: 0 }];
      const t = recalcularTotales(detalles);
      return { ...prev, detalles, ...t };
    });
  };

  const handleDetalleChange = (idx, field, value) => {
    setDocForm(prev => {
      const detalles = prev.detalles.map((d, i) => {
        if (i !== idx) return d;
        const upd = { ...d, [field]: field === "descripcion" ? value : parseFloat(value) || 0 };
        upd.total = ((upd.valor_unitario || 0) + (upd.iva_unitario || 0)) * (upd.cantidad || 0);
        return upd;
      });
      const t = recalcularTotales(detalles);
      return { ...prev, detalles, ...t };
    });
  };

  const handleRemoveDetalle = (idx) => {
    setDocForm(prev => {
      const detalles = prev.detalles.filter((_, i) => i !== idx);
      const t = recalcularTotales(detalles);
      return { ...prev, detalles, ...t };
    });
  };

  const openCrearDocModal = () => {
    setSelectedDocId(null); setDocDetail(null);
    setDocForm({ tipo: "FACTURA_FISICA", numero_documento: "", id_proveedor: "", id_receptor: "", id_area: "", tipo_factura_id: "", asunto: "", fecha_documento: new Date().toISOString().split("T")[0], moneda_id: "", subtotal: 0, iva: 0, total: 0, detalles: [] });
    setShowCrearDocModal(true);
  };

  const openCrearProveedorModal = () => {
    setProveedorForm({ razon_social: "", numero_documento: "", tipo_documento_id: "", email: "", telefono: "" });
    setShowCrearProveedorModal(true);
  };

  const handleProveedorFormChange = (e) => {
    const { name, value } = e.target;
    setProveedorForm(prev => ({ ...prev, [name]: value }));
  };

  const handleCrearProveedorSubmit = async () => {
    if (!proveedorForm.razon_social.trim() || !proveedorForm.numero_documento.trim() || !proveedorForm.tipo_documento_id) {
      alert("Complete Razón Social, Número de Documento y Tipo de Documento"); return;
    }
    setCreandoProveedor(true);
    const payload = {
      razon_social: proveedorForm.razon_social.trim(), numero_documento: proveedorForm.numero_documento.trim(),
      tipo_documento_id: parseInt(proveedorForm.tipo_documento_id), categoria_id: 0, tipo_persona_id: 0,
      actividad_economica_id: 0, direccion_id: 0, nombre_comercial: proveedorForm.razon_social.trim(),
      email: proveedorForm.email.trim() || null, telefono: proveedorForm.telefono.trim() || null, ruta_predeterminada_id: null
    };
    try {
      const res = await fetch(`${API}/proveedor`, {
        method: "POST", headers: { "Content-Type": "application/json", Authorization: `Bearer ${obtenerToken()}` },
        body: JSON.stringify(payload)
      });
      if (!res.ok) { const err = await res.json(); throw new Error(err.error || "Error creando proveedor"); }
      const nuevo = await res.json(); alert("Proveedor creado correctamente");
      const listRes = await fetch(`${API}/proveedor`, { headers: { Authorization: `Bearer ${obtenerToken()}` } });
      const listData = await listRes.json(); setProveedoresCatalogo(Array.isArray(listData) ? listData : []);
      if (nuevo.id) setDocForm(prev => ({ ...prev, id_proveedor: String(nuevo.id) }));
      setShowCrearProveedorModal(false);
    } catch (err) { alert("Error: " + err.message); }
    finally { setCreandoProveedor(false); }
  };

  const handleCrearDocumentoSubmit = async () => {
    if (!docForm.numero_documento || !docForm.id_proveedor || !docForm.id_receptor || !docForm.id_area || !docForm.moneda_id || !docForm.fecha_documento) {
      alert("Complete todos los campos obligatorios (*)"); return;
    }
    if (docForm.detalles.length === 0) { alert("Agregue al menos un ítem"); return; }
    setCreandoDoc(true);
    const payload = {
      tipo: docForm.tipo, numero_documento: docForm.numero_documento.trim(), orden_compra: "",
      id_proveedor: parseInt(docForm.id_proveedor), id_receptor: parseInt(docForm.id_receptor),
      id_area: parseInt(docForm.id_area), tipo_factura_id: docForm.tipo_factura_id ? parseInt(docForm.tipo_factura_id) : null,
      asunto: docForm.asunto.trim(), fecha_documento: docForm.fecha_documento + "T00:00:00Z", fecha_vencimiento: null,
      moneda_id: parseInt(docForm.moneda_id), subtotal: parseFloat(docForm.subtotal) || 0,
      iva: parseFloat(docForm.iva) || 0, total: parseFloat(docForm.total) || 0,
      numero_folios: 1, orientacion_sello_recibido: "", activo: true,
      detalles: docForm.detalles.map(d => ({ descripcion: d.descripcion, cantidad: parseFloat(d.cantidad) || 0, valor_unitario: parseFloat(d.valor_unitario) || 0, iva_unitario: parseFloat(d.iva_unitario) || 0, total: parseFloat(d.total) || 0 }))
    };
    try {
      const res = await fetch(`${API}/documentocomercial`, {
        method: "POST", headers: { "Content-Type": "application/json", Authorization: `Bearer ${obtenerToken()}` },
        body: JSON.stringify(payload)
      });
      if (!res.ok) { const err = await res.json(); throw new Error(err.error || "Error creando documento"); }
      alert("Documento creado correctamente"); setShowCrearDocModal(false);
      setLoadingDocs(true);
      fetch(`${API}/documentocomercial/pendientes`, { headers: { Authorization: `Bearer ${obtenerToken()}` } })
        .then(r => r.json()).then(data => { if (Array.isArray(data)) setDocumentos(data); }).finally(() => setLoadingDocs(false));
    } catch (err) { alert("Error: " + err.message); }
    finally { setCreandoDoc(false); }
  };

  const getFilteredDocs = () => {
    let result = [...documentos];
    if (searchDocs) {
      const q = searchDocs.toLowerCase();
      result = result.filter(d => (d.numero_documento?.toLowerCase().includes(q)) || (d.proveedor?.razon_social?.toLowerCase().includes(q)));
    }
    result.sort((a, b) => {
      if (sortDocs === 'fecha_desc') return new Date(b.created_at) - new Date(a.created_at);
      if (sortDocs === 'fecha_asc') return new Date(a.created_at) - new Date(b.created_at);
      if (sortDocs === 'estado') return a.activo === b.activo ? 0 : (a.activo ? -1 : 1);
      return 0;
    });
    return result;
  };

  const handleSyncInbound = async (dias = 30) => {
    setSyncingInbound(true);
    setSyncResult(null);
    try {
      const res = await fetch(`${API}/inbound/sync`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${obtenerToken()}`,
        },
        body: JSON.stringify({ dias }),
      });
      const data = await res.json();
      if (!res.ok) {
        throw new Error(data.error || "Error al sincronizar con DIAN");
      }
      setSyncResult(data.stats);
      // Recargar documentos pendientes
      setLoadingDocs(true);
      const listRes = await fetch(`${API}/documentocomercial/pendientes?_t=${Date.now()}`, {
        headers: { Authorization: `Bearer ${obtenerToken()}` },
      });
      const listData = await listRes.json();
      if (Array.isArray(listData)) setDocumentos(listData);
    } catch (err) {
      alert("Error en sincronización DIAN: " + err.message);
    } finally {
      setSyncingInbound(false);
      setLoadingDocs(false);
    }
  };

  return {
    documentos, setDocumentos, selectedDocId, setSelectedDocId, docDetail, setDocDetail, loadingDocs,
    searchDocs, setSearchDocs, sortDocs, setSortDocs, getFilteredDocs,
    showCrearDocModal, setShowCrearDocModal, docForm, setDocForm, creandoDoc,
    showCrearProveedorModal, setShowCrearProveedorModal, proveedorForm, creandoProveedor,
    proveedoresCatalogo, receptoresCatalogo, tiposDocumentoCatalogo, areasCatalogo, monedasCatalogo,
    openCrearDocModal, openCrearProveedorModal,
    handleDocFormChange, handleAddDetalle, handleDetalleChange, handleRemoveDetalle,
    handleProveedorFormChange, handleCrearProveedorSubmit, handleCrearDocumentoSubmit,
    syncingInbound, syncResult, handleSyncInbound
  };
}