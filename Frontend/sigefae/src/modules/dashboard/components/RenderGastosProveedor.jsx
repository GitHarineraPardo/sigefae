import { useMemo } from "react";
import "./RenderGastosProveedor.css";
import { formatCurrency } from "../helpers/formatters.js";

const CHART_COLORS = [
  "#1e3a5f",
  "#2c5282",
  "#edab00",
  "#168b8b",
  "#8055a3",
  "#df7041",
];

export default function RenderGastosProveedor({
  gastosHook,
  openSaia,
  userCargo,
}) {
  const {
    data,
    loading,
    error,
    refetch,
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
    activeSubTab,
    setActiveSubTab,
    tipoDesglose,
    setTipoDesglose,
    busquedaItem,
    setBusquedaItem,
    busquedaProveedor,
    setBusquedaProveedor,
  } = gastosHook;

  const proveedoresDisponibles = useMemo(
    () => data?.proveedores || [],
    [data?.proveedores]
  );
  const radicados = useMemo(() => data?.radicados || [], [data?.radicados]);
  const items = useMemo(() => data?.items || [], [data?.items]);
  const normas = useMemo(() => data?.normas || [], [data?.normas]);

  // Proveedores seleccionados como objetos
  const proveedoresSeleccionadosObj = useMemo(() => {
    if (proveedoresSeleccionados.length === 0) return [];
    return proveedoresDisponibles.filter((p) =>
      proveedoresSeleccionados.includes(p.id)
    );
  }, [proveedoresDisponibles, proveedoresSeleccionados]);

  // Proveedores que aún no están seleccionados para el dropdown
  const proveedoresParaAgregar = useMemo(() => {
    return proveedoresDisponibles.filter(
      (p) => !proveedoresSeleccionados.includes(p.id)
    );
  }, [proveedoresDisponibles, proveedoresSeleccionados]);

  // Filtrado de proveedores en Dashboard General
  const proveedoresFiltrados = useMemo(() => {
    if (!busquedaProveedor.trim()) return proveedoresDisponibles;
    const q = busquedaProveedor.toLowerCase();
    return proveedoresDisponibles.filter(
      (p) =>
        p.razon_social?.toLowerCase().includes(q) ||
        p.nit?.toLowerCase().includes(q)
    );
  }, [proveedoresDisponibles, busquedaProveedor]);

  // Filtrado de ítems en Desglose
  const itemsFiltrados = useMemo(() => {
    let list = items;
    // Si hay proveedores seleccionados, filtrar ítems de esos proveedores
    if (proveedoresSeleccionados.length > 0) {
      list = list.filter((it) =>
        proveedoresSeleccionados.includes(it.proveedor_id)
      );
    }
    if (busquedaItem.trim()) {
      const q = busquedaItem.toLowerCase();
      list = list.filter(
        (it) =>
          it.descripcion?.toLowerCase().includes(q) ||
          it.proveedor_nombre?.toLowerCase().includes(q)
      );
    }
    return list;
  }, [items, proveedoresSeleccionados, busquedaItem]);

  // Ranking Top 10 Proveedores para barras visuales
  const top10Proveedores = useMemo(() => {
    return proveedoresDisponibles.slice(0, 10);
  }, [proveedoresDisponibles]);

  const gastoChart = useMemo(() => {
    const proveedoresConGasto = proveedoresDisponibles.filter(
      (proveedor) => Number(proveedor.total_gastado) > 0
    );
    const principales = proveedoresConGasto.slice(0, 5).map((proveedor) => ({
      id: proveedor.id,
      nombre: proveedor.razon_social,
      total: Number(proveedor.total_gastado),
    }));
    const otrosTotal = proveedoresConGasto
      .slice(5)
      .reduce((total, proveedor) => total + Number(proveedor.total_gastado), 0);

    if (otrosTotal > 0) {
      principales.push({
        id: "otros",
        nombre: "Otros proveedores",
        total: otrosTotal,
      });
    }

    const total = principales.reduce((sum, proveedor) => sum + proveedor.total, 0);
    return principales.map((proveedor, index) => {
      const porcentaje = total > 0 ? (proveedor.total / total) * 100 : 0;
      const inicio = principales
        .slice(0, index)
        .reduce(
          (acumulado, anterior) =>
            acumulado + (total > 0 ? (anterior.total / total) * 100 : 0),
          0
        );
      return {
        ...proveedor,
        color: CHART_COLORS[index % CHART_COLORS.length],
        porcentaje,
        inicio,
        fin: inicio + porcentaje,
      };
    });
  }, [proveedoresDisponibles]);

  // Facturas agrupadas por proveedor seleccionado
  const facturasPorProveedor = useMemo(() => {
    const map = {};
    radicados.forEach((rad) => {
      if (!map[rad.proveedor_id]) {
        map[rad.proveedor_id] = [];
      }
      map[rad.proveedor_id].push(rad);
    });
    return map;
  }, [radicados]);

  const handleSelectProveedorChange = (e) => {
    const val = e.target.value;
    if (val) {
      agregarProveedor(val);
      e.target.value = "";
    }
  };

  return (
    <div className="gastos-container">
      {/* ── CARD SUPERIOR: TÍTULO Y FILTROS ── */}
      <div className="gastos-header-card">
        <div className="gastos-title-row">
          <div className="gastos-title">
            <i className="fa-solid fa-chart-pie" />
            <div>
              <h2>Gastos por Proveedor</h2>
              <p>
                Análisis financiero y control de facturas finalizadas aprobadas
                por el cargo
              </p>
            </div>
          </div>

          <div className="gastos-cargo-badge">
            <i className="fa-solid fa-briefcase" />
            <span>
              Cargo:{" "}
              <strong>
                {data?.cargo_nombre || userCargo || "Sin cargo asignado"}
              </strong>
            </span>
          </div>
        </div>

        {error && (
          <div style={{ background: "#fee2e2", border: "1px solid #ef4444", color: "#b91c1c", padding: "10px 14px", borderRadius: "8px", fontSize: "0.88rem", display: "flex", alignItems: "center", gap: "8px" }}>
            <i className="fa-solid fa-triangle-exclamation" /> {error}
          </div>
        )}

        {/* ── BARRA DE CONTROLES Y FECHAS ── */}
        <div className="gastos-filters-grid">
          <div className="filter-group">
            <label>
              <i className="fa-regular fa-calendar" /> Fecha Desde
            </label>
            <input
              type="date"
              className="filter-control"
              value={fechaDesde}
              onChange={(e) => setFechaDesde(e.target.value)}
            />
          </div>

          <div className="filter-group">
            <label>
              <i className="fa-regular fa-calendar" /> Fecha Hasta
            </label>
            <input
              type="date"
              className="filter-control"
              value={fechaHasta}
              onChange={(e) => setFechaHasta(e.target.value)}
            />
          </div>

          <div className="filter-group">
            <label>
              <i className="fa-solid fa-building" /> Añadir Proveedor
            </label>
            <select
              className="filter-control"
              defaultValue=""
              onChange={handleSelectProveedorChange}
              disabled={loading || proveedoresParaAgregar.length === 0}
            >
              <option value="" disabled>
                {proveedoresParaAgregar.length === 0
                  ? "Todos los proveedores añadidos"
                  : "-- Selecciona un proveedor para ver --"}
              </option>
              {proveedoresParaAgregar.map((p) => (
                <option key={p.id} value={p.id}>
                  {p.razon_social} ({p.nit})
                </option>
              ))}
            </select>
          </div>

          <div className="filter-actions">
            <button
              className="doc-btn doc-btn-secondary"
              onClick={limpiarFiltros}
              title="Restablecer filtros"
              style={{ padding: "8px 12px" }}
            >
              <i className="fa-solid fa-rotate-left" /> Limpiar
            </button>
            <button
              className="doc-btn doc-btn-primary"
              onClick={refetch}
              disabled={loading}
              style={{ padding: "8px 16px" }}
            >
              <i className={`fa-solid ${loading ? "fa-spinner fa-spin" : "fa-magnifying-glass"}`} />
              {loading ? " Actualizando..." : " Consultar"}
            </button>
          </div>
        </div>

        {/* ── LISTA DE PROVEEDORES SELECCIONADOS (PILLS) ── */}
        <div className="proveedores-pills-container">
          <div className="pills-header">
            <span>
              <i className="fa-solid fa-filter" /> Proveedores en análisis
              específico ({proveedoresSeleccionados.length} de{" "}
              {proveedoresDisponibles.length})
            </span>
            <div style={{ display: "flex", gap: "8px" }}>
              {proveedoresDisponibles.length > 0 && (
                <button
                  type="button"
                  className="doc-btn doc-btn-secondary"
                  style={{ padding: "3px 8px", fontSize: "0.78rem" }}
                  onClick={seleccionarTodosProveedores}
                >
                  <i className="fa-solid fa-check-double" /> Seleccionar Todos
                </button>
              )}
              {proveedoresSeleccionados.length > 0 && (
                <button
                  type="button"
                  className="doc-btn doc-btn-secondary"
                  style={{ padding: "3px 8px", fontSize: "0.78rem" }}
                  onClick={limpiarProveedores}
                >
                  <i className="fa-solid fa-xmark" /> Quitar Todos
                </button>
              )}
            </div>
          </div>

          <div className="pills-list">
            {proveedoresSeleccionadosObj.length === 0 ? (
              <span className="empty-pills-msg">
                Ningún proveedor seleccionado. Usa la lista desplegable arriba
                para agregar proveedores a tu análisis enfocado o consulta la
                pestaña <strong>Dashboard General</strong>.
              </span>
            ) : (
              proveedoresSeleccionadosObj.map((p) => (
                <span key={p.id} className="prov-pill">
                  <i className="fa-solid fa-building" />
                  {p.razon_social} ({formatCurrency(p.total_gastado)})
                  <button
                    type="button"
                    onClick={() => quitarProveedor(p.id)}
                    title="Quitar de la lista"
                  >
                    <i className="fa-solid fa-xmark" />
                  </button>
                </span>
              ))
            )}
          </div>
        </div>
      </div>

      {/* ── KPIS RESUMEN GENERAL ── */}
      <div className="gastos-kpis-grid">
        <div className="kpi-card">
          <div className="kpi-icon blue">
            <i className="fa-solid fa-sack-dollar" />
          </div>
          <div className="kpi-content">
            <span className="kpi-label">Gasto Total Aprobado</span>
            <span className="kpi-value">
              {formatCurrency(data?.total_general || 0)}
            </span>
            <span className="kpi-subtext">
              Subtotal: {formatCurrency(data?.subtotal_general || 0)} | IVA:{" "}
              {formatCurrency(data?.iva_general || 0)}
            </span>
          </div>
        </div>

        <div className="kpi-card">
          <div className="kpi-icon emerald">
            <i className="fa-solid fa-file-invoice" />
          </div>
          <div className="kpi-content">
            <span className="kpi-label">Facturas Finalizadas</span>
            <span className="kpi-value">{data?.cantidad_facturas || 0}</span>
            <span className="kpi-subtext">Aprobadas por tu cargo</span>
          </div>
        </div>

        <div className="kpi-card">
          <div className="kpi-icon amber">
            <i className="fa-solid fa-truck-field" />
          </div>
          <div className="kpi-content">
            <span className="kpi-label">Proveedores Totales</span>
            <span className="kpi-value">{data?.cantidad_proveedores || 0}</span>
            <span className="kpi-subtext">
              Con facturas aprobadas y finalizadas
            </span>
          </div>
        </div>

        <div className="kpi-card">
          <div className="kpi-icon purple">
            <i className="fa-solid fa-calculator" />
          </div>
          <div className="kpi-content">
            <span className="kpi-label">Ticket Promedio</span>
            <span className="kpi-value">
              {formatCurrency(data?.promedio_factura || 0)}
            </span>
            <span className="kpi-subtext">Por factura radicada</span>
          </div>
        </div>
      </div>

      {/* ── NAVEGACIÓN ENTRE SUB-DASHBOARDS ── */}
      <div className="gastos-nav-tabs">
        <button
          className={`gastos-tab-btn ${activeSubTab === "general" ? "active" : ""}`}
          onClick={() => setActiveSubTab("general")}
        >
          <i className="fa-solid fa-table-cells" /> Dashboard General (Todos los
          Proveedores)
        </button>
        <button
          className={`gastos-tab-btn ${activeSubTab === "por_proveedor" ? "active" : ""}`}
          onClick={() => setActiveSubTab("por_proveedor")}
        >
          <i className="fa-solid fa-building-user" /> Gasto por Proveedor
          Seleccionado ({proveedoresSeleccionados.length})
        </button>
        <button
          className={`gastos-tab-btn ${activeSubTab === "desglose" ? "active" : ""}`}
          onClick={() => setActiveSubTab("desglose")}
        >
          <i className="fa-solid fa-layer-group" /> Desglose de Gastos
          (Conceptos e Ítems)
        </button>
      </div>

      {/* ── CONTENIDO: SUB-DASHBOARD 1: GENERAL ── */}
      {activeSubTab === "general" && (
        <div className="dashboard-card">
          <div className="card-header-flex">
            <h3>
              <i className="fa-solid fa-ranking-star" /> Top 10 Proveedores con
              Mayor Gasto
            </h3>
            <span style={{ fontSize: "0.85rem", color: "#64748b" }}>
              Mostrando los 10 proveedores con mayor desembolso
            </span>
          </div>

          {top10Proveedores.length === 0 ? (
            <div className="gastos-empty">
              <i className="fa-solid fa-inbox" />
              <p>No hay facturas finalizadas aprobadas por este cargo.</p>
            </div>
          ) : (
            <>
              {gastoChart.length > 0 && (
                <section
                  className="gastos-chart-panel"
                  aria-labelledby="gastos-chart-title"
                >
                  <div className="gastos-chart-heading">
                    <span className="gastos-chart-icon">
                      <i className="fa-solid fa-chart-pie" />
                    </span>
                    <div>
                      <h4 id="gastos-chart-title">
                        Participación por proveedor
                      </h4>
                      <p>
                        Distribución del gasto aprobado para{" "}
                        {data?.cargo_nombre || userCargo || "tu cargo"}.
                      </p>
                    </div>
                  </div>
                  <div className="gastos-chart-content">
                    <div
                      className="gastos-donut"
                      role="img"
                      aria-label={`Distribución porcentual del gasto por proveedor. Total ${formatCurrency(
                        data?.total_general || 0
                      )}.`}
                      style={{
                        background: `conic-gradient(${gastoChart
                          .map(
                            (segmento) =>
                              `${segmento.color} ${segmento.inicio}% ${segmento.fin}%`
                          )
                          .join(", ")})`,
                      }}
                    >
                      <div className="gastos-donut-center">
                        <span>Total del cargo</span>
                        <strong>{formatCurrency(data?.total_general || 0)}</strong>
                      </div>
                    </div>
                    <ul className="gastos-chart-legend">
                      {gastoChart.map((segmento) => (
                        <li key={segmento.id}>
                          <span
                            className="gastos-legend-color"
                            style={{ backgroundColor: segmento.color }}
                          />
                          <span
                            className="gastos-legend-name"
                            title={segmento.nombre}
                          >
                            {segmento.nombre}
                          </span>
                          <strong>{segmento.porcentaje.toFixed(1)}%</strong>
                          <span className="gastos-legend-total">{formatCurrency(segmento.total)}</span>
                        </li>
                      ))}
                    </ul>
                  </div>
                </section>
              )}
              <div className="ranking-list">
                {top10Proveedores.map((p, index) => {
                  const maxGasto = top10Proveedores[0]?.total_gastado || 1;
                  const pctBar = Math.min(
                    100,
                    Math.round((p.total_gastado / maxGasto) * 100)
                  );
                  return (
                    <div key={p.id} className="ranking-item">
                      <div className="ranking-meta">
                        <div className="ranking-prov-name">
                          <span
                            className={`ranking-badge-pos ${
                              index === 0
                                ? "top1"
                                : index === 1
                                  ? "top2"
                                  : index === 2
                                    ? "top3"
                                    : ""
                            }`}
                          >
                            {index + 1}
                          </span>
                          <span>{p.razon_social}</span>
                          <small style={{ color: "#94a3b8" }}>({p.nit})</small>
                        </div>
                        <div
                          style={{
                            display: "flex",
                            alignItems: "center",
                            gap: "14px",
                          }}
                        >
                          <span style={{ fontSize: "0.82rem", color: "#64748b" }}>
                            {p.cantidad_facturas} fact. ({p.porcentaje_gasto}%)
                          </span>
                          <span className="ranking-amount">
                            {formatCurrency(p.total_gastado)}
                          </span>
                          {!proveedoresSeleccionados.includes(p.id) && (
                            <button
                              type="button"
                              className="doc-btn doc-btn-secondary"
                              style={{
                                padding: "4px 8px",
                                fontSize: "0.75rem",
                              }}
                              onClick={() => {
                                agregarProveedor(p.id);
                                setActiveSubTab("por_proveedor");
                              }}
                              title="Analizar a fondo este proveedor"
                            >
                              <i className="fa-solid fa-arrow-right" /> Analizar
                            </button>
                          )}
                        </div>
                      </div>
                      <div className="ranking-bar-bg">
                        <div
                          className="ranking-bar-fill"
                          style={{ width: `${pctBar}%` }}
                        />
                      </div>
                    </div>
                  );
                })}
              </div>
            </>
          )}

          {/* Tabla Comparativa Completa */}
          <div style={{ marginTop: "24px" }}>
            <div className="card-header-flex" style={{ marginBottom: "14px" }}>
              <h3>
                <i className="fa-solid fa-list-check" /> Todos los Proveedores (
                {proveedoresFiltrados.length})
              </h3>
              <input
                type="text"
                placeholder="Buscar proveedor o NIT..."
                className="filter-control gastos-search"
                value={busquedaProveedor}
                onChange={(e) => setBusquedaProveedor(e.target.value)}
              />
            </div>

            <div className="gastos-table-wrapper">
              <table className="gastos-table">
                <thead>
                  <tr>
                    <th>#</th>
                    <th>Proveedor / Razón Social</th>
                    <th>NIT</th>
                    <th className="text-right">N° Facturas</th>
                    <th className="text-right">Subtotal</th>
                    <th className="text-right">IVA</th>
                    <th className="text-right">Total Gastado</th>
                    <th className="text-right">% Participación</th>
                    <th>Acción</th>
                  </tr>
                </thead>
                <tbody>
                  {proveedoresFiltrados.map((p, idx) => (
                    <tr key={p.id}>
                      <td>{idx + 1}</td>
                      <td>
                        <strong>{p.razon_social}</strong>
                      </td>
                      <td>{p.nit}</td>
                      <td className="text-right">{p.cantidad_facturas}</td>
                      <td className="monto-col">
                        {formatCurrency(p.total_subtotal)}
                      </td>
                      <td className="monto-col">
                        {formatCurrency(p.total_iva)}
                      </td>
                      <td className="monto-col" style={{ color: "#2c5282" }}>
                        {formatCurrency(p.total_gastado)}
                      </td>
                      <td className="text-right">
                        <strong>{p.porcentaje_gasto}%</strong>
                      </td>
                      <td>
                        <button
                          type="button"
                          className="doc-btn doc-btn-secondary"
                          style={{ padding: "4px 10px", fontSize: "0.8rem" }}
                          onClick={() => {
                            agregarProveedor(p.id);
                            setActiveSubTab("por_proveedor");
                          }}
                        >
                          <i className="fa-solid fa-magnifying-glass-chart" /> Ver
                          detalle
                        </button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        </div>
      )}

      {/* ── CONTENIDO: SUB-DASHBOARD 2: POR PROVEEDOR SELECCIONADO ── */}
      {activeSubTab === "por_proveedor" && (
        <div className="dashboard-card">
          <div className="card-header-flex">
            <h3>
              <i className="fa-solid fa-building-circle-check" /> Detalle de
              Proveedores Seleccionados ({proveedoresSeleccionadosObj.length})
            </h3>
            {proveedoresSeleccionadosObj.length > 0 && (
              <button
                type="button"
                className="doc-btn doc-btn-secondary"
                onClick={limpiarProveedores}
              >
                <i className="fa-solid fa-eraser" /> Limpiar selección
              </button>
            )}
          </div>

          {proveedoresSeleccionadosObj.length === 0 ? (
            <div className="gastos-empty">
              <i className="fa-solid fa-hand-pointer" />
              <p>
                No has agregado ningún proveedor al análisis específico todavía.
              </p>
              <p style={{ fontSize: "0.85rem", color: "#64748b" }}>
                Selecciona uno o más proveedores en la lista desplegable de arriba, o
                haz clic en &quot;Seleccionar Todos&quot;.
              </p>
              <button
                type="button"
                className="doc-btn doc-btn-primary"
                onClick={seleccionarTodosProveedores}
                style={{ marginTop: "8px" }}
              >
                <i className="fa-solid fa-check-double" /> Analizar Todos los
                Proveedores ({proveedoresDisponibles.length})
              </button>
            </div>
          ) : (
            <div
              style={{ display: "flex", flexDirection: "column", gap: "20px" }}
            >
              {proveedoresSeleccionadosObj.map((p) => {
                const facturas = facturasPorProveedor[p.id] || [];
                return (
                  <div key={p.id} className="prov-detail-card">
                    <div className="prov-card-header">
                      <div className="prov-info">
                        <h4>{p.razon_social}</h4>
                        <span>NIT: {p.nit}</span>
                      </div>
                      <button
                        type="button"
                        className="doc-btn doc-btn-secondary"
                        style={{
                          padding: "4px 8px",
                          fontSize: "0.78rem",
                          color: "#ef4444",
                          borderColor: "#fecaca",
                        }}
                        onClick={() => quitarProveedor(p.id)}
                        title="Quitar este proveedor del análisis"
                      >
                        <i className="fa-solid fa-xmark" /> Quitar
                      </button>
                    </div>

                    <div className="prov-card-totals">
                      <div className="prov-total-item">
                        <span className="lbl">Total Gastado</span>
                        <span className="val" style={{ color: "#2c5282" }}>
                          {formatCurrency(p.total_gastado)}
                        </span>
                      </div>
                      <div className="prov-total-item">
                        <span className="lbl">Subtotal</span>
                        <span className="val">
                          {formatCurrency(p.total_subtotal)}
                        </span>
                      </div>
                      <div className="prov-total-item">
                        <span className="lbl">IVA</span>
                        <span className="val">
                          {formatCurrency(p.total_iva)}
                        </span>
                      </div>
                      <div className="prov-total-item">
                        <span className="lbl">Facturas Aprobadas</span>
                        <span className="val">{p.cantidad_facturas}</span>
                      </div>
                      <div className="prov-total-item">
                        <span className="lbl">% del Gasto Total</span>
                        <span className="val">{p.porcentaje_gasto}%</span>
                      </div>
                    </div>

                    {/* Tabla de Facturas de este proveedor */}
                    <div style={{ marginTop: "10px" }}>
                      <h5
                        style={{
                          margin: "0 0 10px 0",
                          fontSize: "0.85rem",
                          color: "#475569",
                        }}
                      >
                        <i className="fa-solid fa-receipt" /> Facturas Asociadas
                        ({facturas.length}):
                      </h5>
                      <div className="gastos-table-wrapper">
                        <table className="gastos-table">
                          <thead>
                            <tr>
                              <th>N° Radicado</th>
                              <th>N° Factura</th>
                              <th>Fecha Radicación</th>
                              <th>Fecha Factura</th>
                              <th className="text-right">Subtotal</th>
                              <th className="text-right">IVA</th>
                              <th className="text-right">Total</th>
                              <th>Acciones</th>
                            </tr>
                          </thead>
                          <tbody>
                            {facturas.map((f) => (
                              <tr key={f.id}>
                                <td>
                                  <strong>#{f.numero_radicado}</strong>
                                </td>
                                <td>{f.numero_documento}</td>
                                <td>
                                  {f.fecha_radicacion
                                    ? new Date(
                                        f.fecha_radicacion
                                      ).toLocaleDateString()
                                    : "—"}
                                </td>
                                <td>
                                  {f.fecha_documento
                                    ? new Date(
                                        f.fecha_documento
                                      ).toLocaleDateString()
                                    : "—"}
                                </td>
                                <td className="monto-col">
                                  {formatCurrency(f.subtotal)}
                                </td>
                                <td className="monto-col">
                                  {formatCurrency(f.iva)}
                                </td>
                                <td
                                  className="monto-col"
                                  style={{ color: "#2c5282" }}
                                >
                                  {formatCurrency(f.total)}
                                </td>
                                <td>
                                  <button
                                    type="button"
                                    className="doc-btn doc-btn-secondary"
                                    style={{
                                      padding: "3px 8px",
                                      fontSize: "0.78rem",
                                    }}
                                    onClick={() => openSaia(f, "radicados")}
                                    title="Ver expediente en SAIA"
                                  >
                                    <i className="fa-solid fa-eye" /> Ver
                                    Expediente
                                  </button>
                                </td>
                              </tr>
                            ))}
                          </tbody>
                        </table>
                      </div>
                    </div>
                  </div>
                );
              })}
            </div>
          )}
        </div>
      )}

      {/* ── CONTENIDO: SUB-DASHBOARD 3: DESGLOSE DE GASTOS ── */}
      {activeSubTab === "desglose" && (
        <div className="dashboard-card">
          <div className="card-header-flex">
            <div>
              <h3>
                <i className="fa-solid fa-sitemap" /> Desglose de en qué se ha
                gastado
              </h3>
              <p
                style={{
                  margin: "4px 0 0 0",
                  fontSize: "0.85rem",
                  color: "#64748b",
                }}
              >
                Auditoría detallada por conceptos de artículos adquiridos o por
                centros de costo/normas de reparto
              </p>
            </div>

            <div style={{ display: "flex", gap: "8px" }}>
              <button
                className={`doc-btn ${
                  tipoDesglose === "items"
                    ? "doc-btn-primary"
                    : "doc-btn-secondary"
                }`}
                onClick={() => setTipoDesglose("items")}
              >
                <i className="fa-solid fa-boxes-stacked" /> Por Artículos e
                Ítems
              </button>
              <button
                className={`doc-btn ${
                  tipoDesglose === "normas"
                    ? "doc-btn-primary"
                    : "doc-btn-secondary"
                }`}
                onClick={() => setTipoDesglose("normas")}
              >
                <i className="fa-solid fa-diagram-project" /> Por Normas /
                Centros de Costo
              </button>
            </div>
          </div>

          {tipoDesglose === "items" ? (
            <div>
              <div
                className="card-header-flex"
                style={{ marginBottom: "14px" }}
              >
                <span style={{ fontSize: "0.85rem", color: "#64748b" }}>
                  Mostrando {itemsFiltrados.length} conceptos comprados
                  {proveedoresSeleccionados.length > 0 &&
                    ` (filtrado por los ${proveedoresSeleccionados.length} proveedores seleccionados)`}
                </span>
                <input
                  type="text"
                  placeholder="Buscar ítem o descripción..."
                  className="filter-control gastos-search"
                  value={busquedaItem}
                  onChange={(e) => setBusquedaItem(e.target.value)}
                />
              </div>

              {itemsFiltrados.length === 0 ? (
                <div className="gastos-empty">
                  <i className="fa-solid fa-boxes-packing" />
                  <p>No se encontraron ítems con los filtros aplicados.</p>
                </div>
              ) : (
                <div className="gastos-table-wrapper">
                  <table className="gastos-table">
                    <thead>
                      <tr>
                        <th>#</th>
                        <th>Concepto / Descripción del Ítem</th>
                        <th>Proveedor Suministrante</th>
                        <th className="text-right">Cantidad Total</th>
                        <th className="text-right">N° Facturas</th>
                        <th className="text-right">Total Gastado</th>
                        <th className="text-right">% del Total</th>
                      </tr>
                    </thead>
                    <tbody>
                      {itemsFiltrados.map((it, idx) => (
                        <tr key={idx}>
                          <td>{idx + 1}</td>
                          <td>
                            <strong>{it.descripcion}</strong>
                          </td>
                          <td>{it.proveedor_nombre}</td>
                          <td className="text-right">{it.cantidad_total}</td>
                          <td className="text-right">
                            {it.cantidad_facturas}
                          </td>
                          <td
                            className="monto-col"
                            style={{ color: "#2c5282" }}
                          >
                            {formatCurrency(it.total_gastado)}
                          </td>
                          <td className="text-right">
                            <strong>{it.porcentaje_gasto}%</strong>
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              )}
            </div>
          ) : (
            <div>
              <div
                className="card-header-flex"
                style={{ marginBottom: "14px" }}
              >
                <span style={{ fontSize: "0.85rem", color: "#64748b" }}>
                  Mostrando {normas.length} normas de reparto y centros de costo
                  imputados
                </span>
              </div>

              {normas.length === 0 ? (
                <div className="gastos-empty">
                  <i className="fa-solid fa-network-wired" />
                  <p>
                    No se han registrado normas de reparto en los radicados
                    aprobados.
                  </p>
                </div>
              ) : (
                <div className="gastos-table-wrapper">
                  <table className="gastos-table">
                    <thead>
                      <tr>
                        <th>#</th>
                        <th>Código Norma</th>
                        <th>Nombre Cuenta / Centro</th>
                        <th>Sucursal</th>
                        <th>Departamento</th>
                        <th className="text-right">N° Facturas</th>
                        <th className="text-right">Total Asignado</th>
                        <th className="text-right">% del Total</th>
                      </tr>
                    </thead>
                    <tbody>
                      {normas.map((n, idx) => (
                        <tr key={idx}>
                          <td>{idx + 1}</td>
                          <td>
                            <span
                              style={{
                                background: "#e8eef5",
                                color: "#1e3a5f",
                                padding: "2px 8px",
                                borderRadius: "4px",
                                fontWeight: "600",
                              }}
                            >
                              {n.codigo}
                            </span>
                          </td>
                          <td>
                            <strong>{n.nombre}</strong>
                          </td>
                          <td>{n.sucursal}</td>
                          <td>{n.departamento}</td>
                          <td className="text-right">{n.cantidad_facturas}</td>
                          <td
                            className="monto-col"
                            style={{ color: "#2c5282" }}
                          >
                            {formatCurrency(n.total_gastado)}
                          </td>
                          <td className="text-right">
                            <strong>{n.porcentaje_gasto}%</strong>
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              )}
            </div>
          )}
        </div>
      )}
    </div>
  );
}
