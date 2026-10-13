export default function Sidebar({ isSidebarOpen, setIsSidebarOpen, activeTab, setActiveTab, esAdmin, esUsuario, userRole }) {
  return (
    <aside className={`sidebar ${isSidebarOpen ? "open" : ""}`}>
      <div className="sidebar-header">
        <div className="sidebar-title"><i className="fa-solid fa-gear"></i> Menú</div>
      </div>
      <nav className="menu-nav">
        {esUsuario && userRole !== "Contabilidad" && userRole !== "Tesorería" && (
          <a href="#" className={`menu-item ${activeTab === "tareas" ? "active" : ""}`} onClick={(e) => { e.preventDefault(); setActiveTab("tareas"); setIsSidebarOpen(false); }}>
            <div className="item-icon"><i className="fa-solid fa-clipboard-list"></i></div>
            <div className="item-text"><span className="item-nombre">Mis Tareas</span></div>
          </a>
        )}

        <a href="#" className={`menu-item ${activeTab === "gastos-proveedor" ? "active" : ""}`} onClick={(e) => { e.preventDefault(); setActiveTab("gastos-proveedor"); setIsSidebarOpen(false); }}>
          <div className="item-icon"><i className="fa-solid fa-chart-pie"></i></div>
          <div className="item-text"><span className="item-nombre">Gastos por Proveedor</span></div>
        </a>

        {esAdmin && (
          <>
            <a href="#" className={`menu-item ${activeTab === "correos" ? "active" : ""}`} onClick={(e) => { e.preventDefault(); setActiveTab("correos"); setIsSidebarOpen(false); }}>
              <div className="item-icon"><i className="fa-solid fa-envelope"></i></div>
              <div className="item-text"><span className="item-nombre">Correos</span></div>
            </a>
            <a href="#" className={`menu-item ${activeTab === "documentos" ? "active" : ""}`} onClick={(e) => { e.preventDefault(); setActiveTab("documentos"); setIsSidebarOpen(false); }}>
              <div className="item-icon"><i className="fa-solid fa-file-invoice"></i></div>
              <div className="item-text"><span className="item-nombre">Documentos</span></div>
            </a>
          </>
        )}

        {esAdmin && (
          <a href="#" className={`menu-item ${activeTab === "radicados" ? "active" : ""}`} onClick={(e) => { e.preventDefault(); setActiveTab("radicados"); setIsSidebarOpen(false); }}>
            <div className="item-icon"><i className="fa-solid fa-stamp"></i></div>
            <div className="item-text"><span className="item-nombre">Radicados</span></div>
          </a>
        )}

        {userRole !== "Contabilidad" && userRole !== "Tesorería" && (
          <a href="#" className={`menu-item ${activeTab === "solicitudes" ? "active" : ""}`} onClick={(e) => { e.preventDefault(); setActiveTab("solicitudes"); setIsSidebarOpen(false); }}>
            <div className="item-icon"><i className="fa-solid fa-circle-exclamation"></i></div>
            <div className="item-text"><span className="item-nombre">Solicitudes</span></div>
          </a>
        )}

        {(userRole === "Contabilidad" || userRole === "Tesorería" || userRole === "Superadministrador") && (
          <a href="#" className={`menu-item ${activeTab === "finanzas" ? "active" : ""}`} onClick={(e) => { e.preventDefault(); setActiveTab("finanzas"); setIsSidebarOpen(false); }}>
            <div className="item-icon"><i className="fa-solid fa-file-invoice-dollar"></i></div>
            <div className="item-text"><span className="item-nombre">Control Financiero</span></div>
          </a>
        )}

        {esAdmin && (
          <a href="#" className={`menu-item ${activeTab === "catalogos" ? "active" : ""}`} onClick={(e) => { e.preventDefault(); setActiveTab("catalogos"); setIsSidebarOpen(false); }}>
            <div className="item-icon"><i className="fa-solid fa-sliders"></i></div>
            <div className="item-text"><span className="item-nombre">Catálogos</span></div>
          </a>
        )}

        {(userRole === "Contabilidad" || esAdmin) && (
          <a href="#" className={`menu-item ${activeTab === "trazabilidad-area" ? "active" : ""}`} onClick={(e) => { e.preventDefault(); setActiveTab("trazabilidad-area"); setIsSidebarOpen(false); }}>
            <div className="item-icon"><i className="fa-solid fa-route"></i></div>
            <div className="item-text"><span className="item-nombre">Trazabilidad por Área</span></div>
          </a>
        )}
      </nav>
    </aside>
  );
}
