import { useState, useEffect } from "react";
import { API } from "../constants/api";
import { catalogoConfig } from "../constants/catalogoConfig";

export function useCatalogos(obtenerToken, activeTab) {
  const [catalogoActivo, setCatalogoActivo] = useState("tipo-radicacion");
  const [catalogoItems, setCatalogoItems] = useState([]);
  const [catalogoLoading, setCatalogoLoading] = useState(false);
  const [showCatalogoForm, setShowCatalogoForm] = useState(false);
  const [catalogoEditing, setCatalogoEditing] = useState(null);
  const [catalogoForm, setCatalogoForm] = useState({});

  const [tiposPagoCatalogo, setTiposPagoCatalogo] = useState([]);
  const [areasCatalogo, setAreasCatalogo] = useState([]);
  const [rutasCatalogo, setRutasCatalogo] = useState([]);
  const [usuariosCatalogo, setUsuariosCatalogo] = useState([]);
  const [monedasCatalogo, setMonedasCatalogo] = useState([]);

  const cfg = catalogoConfig[catalogoActivo];

  const loadCatalogo = async (tipo) => {
    const c = catalogoConfig[tipo];
    setCatalogoLoading(true);
    try {
      const cacheBuster = `_t=${new Date().getTime()}`;
      const url = `${API}/${c.endpoint}${c.endpoint.includes('?') ? '&' : '?'}${cacheBuster}`;
      const res = await fetch(url, { headers: { Authorization: `Bearer ${obtenerToken()}` } });
      const data = await res.json();
      setCatalogoItems(Array.isArray(data) ? data : []);
    } catch (err) { console.error(err); setCatalogoItems([]); }
    finally { setCatalogoLoading(false); }
  };

  useEffect(() => { if (activeTab === "catalogos") loadCatalogo(catalogoActivo); }, [activeTab, catalogoActivo]);
  useEffect(() => {
    if (activeTab !== "catalogos") return;
    const headers = { Authorization: `Bearer ${obtenerToken()}` };
    const t = `?_t=${new Date().getTime()}`;
    if (catalogoActivo === "reglas-monto") fetch(`${API}/monedas${t}`, { headers }).then(r => r.ok ? r.json() : []).then(d => setMonedasCatalogo(Array.isArray(d) ? d : [])).catch(() => setMonedasCatalogo([]));
    if (catalogoActivo === "metodos-pago") fetch(`${API}/tipos-pago${t}`, { headers }).then(r => r.json()).then(d => setTiposPagoCatalogo(Array.isArray(d) ? d : []));
    if (catalogoActivo === "rutas" || catalogoActivo === "pasos-ruta") fetch(`${API}/areas${t}`, { headers }).then(r => r.json()).then(d => setAreasCatalogo(Array.isArray(d) ? d : []));
    if (catalogoActivo === "pasos-ruta") fetch(`${API}/rutas${t}`, { headers }).then(r => r.json()).then(d => setRutasCatalogo(Array.isArray(d) ? d : [])).catch(err => console.error(err));
    if (catalogoActivo === "pasos-ruta" || catalogoActivo === "reglas-monto") fetch(`${API}/usuarios${t}`, { headers }).then(r => r.json()).then(d => setUsuariosCatalogo(Array.isArray(d) ? d : []));
  }, [activeTab, catalogoActivo, obtenerToken]);

  const openCatalogoCreate = () => {
    setCatalogoEditing(null);
    const empty = {};
    cfg.fields.forEach(f => empty[f] = f === "orden" ? 1 : "");
    setCatalogoForm(empty);
    setShowCatalogoForm(true);
  };

  const openCatalogoEdit = (item) => {
    setCatalogoEditing(item);
    const values = {};
    cfg.fields.forEach(f => { values[f] = item[f] !== undefined ? String(item[f]) : ""; });
    setCatalogoForm(values);
    setShowCatalogoForm(true);
  };

  const handleCatalogoFormChange = (e) => {
    const { name, value, type } = e.target;
    setCatalogoForm(prev => ({ ...prev, [name]: type === "number" ? (value === "" ? 0 : parseFloat(value)) : value }));
  };
const handleCatalogoSubmit = async () => {
  const isEdit = !!catalogoEditing;
  const url = isEdit ? `${API}/${cfg.endpoint}/${catalogoEditing.id}` : `${API}/${cfg.endpoint}`;
  const method = isEdit ? cfg.method : "POST";
  const body = {};

  cfg.fields.forEach(f => {
    const val = catalogoForm[f];
    if (f === "orden" || f.includes("_id") || f.includes("monto") || f === "prioridad" || f === "ano" || f === "valor") {
      // Solo parsear si hay un valor ingresado para evitar enviar 0 en campos requeridos por Gin
      if (val !== "" && val !== undefined && val !== null) {
        body[f] = parseFloat(val);
      }
    } else {
      body[f] = val?.trim() || "";
    }
  });

  // Validaciones del Frontend
  if (!body.nombre && cfg.fields.includes("nombre")) { alert("El nombre es obligatorio"); return; }
  if (catalogoActivo === "rutas" && !body.zona) { alert("La zona es obligatoria"); return; }
  if (catalogoActivo === "rutas" && !body.area_id) { alert("El área es obligatoria"); return; }
  if (catalogoActivo === "normas-reparto" && !body.descripcion) { alert("La descripción es obligatoria"); return; }
  if (body.descripcion && body.descripcion.length > 100) { alert("La descripción no puede superar 100 caracteres"); return; }

  try {
    const res = await fetch(url, {
      method,
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${obtenerToken()}`
      },
      body: JSON.stringify(body)
    });

    const data = await res.json();

    if (!res.ok) {
      console.error("Detalle error Backend (Go):", data); // Muestra la causa exacta en consola F12
      throw new Error(data.error || JSON.stringify(data));
    }

    setShowCatalogoForm(false);
    loadCatalogo(catalogoActivo);
  } catch (err) {
    alert("Error: " + err.message);
  }
};

  const handleToggleCatalogoStatus = async (item) => {
    try {
      const res = await fetch(`${API}/${cfg.endpoint}/${item.id}/activo`, {
        method: "PATCH", headers: { "Content-Type": "application/json", Authorization: `Bearer ${obtenerToken()}` },
        body: JSON.stringify({ activo: !item.activo }),
      });
      if (!res.ok) throw new Error("Error cambiando estado");
      loadCatalogo(catalogoActivo);
    } catch (err) { alert(err.message); }
  };

  return {
    catalogoActivo, setCatalogoActivo, catalogoItems, catalogoLoading,
    showCatalogoForm, setShowCatalogoForm, catalogoEditing, catalogoForm,
    tiposPagoCatalogo, areasCatalogo, rutasCatalogo, usuariosCatalogo, monedasCatalogo,
    cfg, openCatalogoCreate, openCatalogoEdit, handleCatalogoFormChange,
    handleCatalogoSubmit, handleToggleCatalogoStatus
  };
}