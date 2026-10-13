import { useId, useState } from "react";

const normalize = (text = "") =>
  text.normalize("NFD").replace(/[\u0300-\u036f]/g, "").toLocaleLowerCase();

export default function ProjectInput({
  value,
  onChange,
  projects,
  listId,
  onCreateProject,
  zones,
  zoneDefault = "",
  inputStyle,
  containerStyle,
  placeholder = "Proyecto",
}) {
  const id = useId().replaceAll(":", "");
  const suggestionsId = `${listId}-${id}`;
  const [showCreateForm, setShowCreateForm] = useState(false);
  const [zone, setZone] = useState(zoneDefault);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");
  const [suggestionsOpen, setSuggestionsOpen] = useState(false);
  const [activeSuggestion, setActiveSuggestion] = useState(-1);

  const projectName = value?.trim() || "";
  const normalizedQuery = normalize(projectName);
  const searchTerms = normalizedQuery.split(/\s+/).filter(Boolean);
  const allMatchingProjects = normalizedQuery
    ? projects
      .filter((project) => searchTerms.every((term) => normalize(project.nombre).includes(term)))
      .sort((a, b) => {
        const aStartsWith = normalize(a.nombre).startsWith(normalizedQuery);
        const bStartsWith = normalize(b.nombre).startsWith(normalizedQuery);
        if (aStartsWith !== bStartsWith) return aStartsWith ? -1 : 1;
        return (a.nombre || "").localeCompare(b.nombre || "");
      })
    : [];
  const matchingProjects = allMatchingProjects.slice(0, 8);
  const projectExists = projects.some(
    (project) => normalize(project.nombre?.trim()) === normalizedQuery,
  );

  const selectProject = (project) => {
    onChange(project.nombre);
    setSuggestionsOpen(false);
    setActiveSuggestion(-1);
    setShowCreateForm(false);
    setError("");
  };

  const handleInputKeyDown = (event) => {
    if (event.key === "ArrowDown" && matchingProjects.length) {
      event.preventDefault();
      setSuggestionsOpen(true);
      setActiveSuggestion((current) => (current + 1) % matchingProjects.length);
    } else if (event.key === "ArrowUp" && matchingProjects.length) {
      event.preventDefault();
      setActiveSuggestion((current) => current <= 0 ? matchingProjects.length - 1 : current - 1);
    } else if (event.key === "Enter" && suggestionsOpen && activeSuggestion >= 0) {
      event.preventDefault();
      selectProject(matchingProjects[activeSuggestion]);
    } else if (event.key === "Escape") {
      setSuggestionsOpen(false);
      setActiveSuggestion(-1);
    }
  };

  const handleCreate = async () => {
    setSaving(true);
    setError("");
    try {
      const project = await onCreateProject(projectName, zone);
      onChange(project.nombre);
      setShowCreateForm(false);
    } catch (createError) {
      setError(createError.message || "No se pudo crear el proyecto.");
    } finally {
      setSaving(false);
    }
  };

  return (
    <div style={{ minWidth: 0, width: "100%", ...containerStyle }}>
      <div style={{ position: "relative", width: "100%" }}>
        <input
          type="text"
          className="doc-input"
          style={{ display: "block", width: "100%", minWidth: 0, boxSizing: "border-box", ...inputStyle }}
          placeholder={placeholder}
          value={value || ""}
          role="combobox"
          aria-autocomplete="list"
          aria-expanded={suggestionsOpen && matchingProjects.length > 0}
          aria-controls={suggestionsId}
          aria-activedescendant={activeSuggestion >= 0 ? `${suggestionsId}-${matchingProjects[activeSuggestion]?.id}` : undefined}
          onFocus={() => {
            if (normalizedQuery) setSuggestionsOpen(true);
          }}
          onKeyDown={handleInputKeyDown}
          onChange={(event) => {
            onChange(event.target.value);
            setSuggestionsOpen(true);
            setActiveSuggestion(-1);
            setShowCreateForm(false);
            setError("");
          }}
          onBlur={() => {
            setSuggestionsOpen(false);
            setActiveSuggestion(-1);
          }}
        />
        {suggestionsOpen && matchingProjects.length > 0 && (
          <div
            style={{
              position: "absolute",
              zIndex: 20,
              top: "calc(100% + 4px)",
              left: 0,
              width: "max(100%, min(320px, calc(100vw - 32px)))",
              maxHeight: 240,
              overflowY: "auto",
              padding: 5,
              background: "#fff",
              border: "1px solid #cbd5e1",
              borderRadius: 8,
              boxShadow: "0 8px 20px rgba(15, 23, 42, 0.16)",
            }}
          >
            <ul
              id={suggestionsId}
              role="listbox"
              style={{ margin: 0, padding: 5, listStyle: "none" }}
            >
              {matchingProjects.map((project, index) => (
                <li
                  id={`${suggestionsId}-${project.id}`}
                  key={project.id}
                  role="option"
                  aria-selected={index === activeSuggestion}
                  onMouseDown={(event) => event.preventDefault()}
                  onClick={() => selectProject(project)}
                  style={{
                    display: "flex",
                    alignItems: "center",
                    justifyContent: "space-between",
                    gap: 8,
                    padding: "10px 12px",
                    borderRadius: 6,
                    cursor: "pointer",
                    background: index === activeSuggestion ? "#eff6ff" : "#fff",
                    color: "#1e293b",
                  }}
                >
                  <span style={{ overflow: "hidden", textOverflow: "ellipsis", whiteSpace: "nowrap", fontWeight: 600, textAlign: "left" }}>
                    {project.nombre}
                  </span>
                  {project.zona && (
                    <span style={{ flexShrink: 0, color: "#475569", background: "#f1f5f9", borderRadius: 999, padding: "3px 8px", fontSize: "0.72em" }}>
                      {project.zona}
                    </span>
                  )}
                </li>
              ))}
            </ul>
            {allMatchingProjects.length > matchingProjects.length && (
              <small style={{ display: "block", padding: "3px 10px 7px", color: "#64748b" }}>
                Mostrando {matchingProjects.length} de {allMatchingProjects.length} coincidencias.
              </small>
            )}
          </div>
        )}
        {suggestionsOpen && normalizedQuery && matchingProjects.length === 0 && (
          <div style={{ padding: "9px 2px", color: "#64748b", fontSize: "0.82em" }}>
            No hay coincidencias por nombre. Puedes crear el proyecto.
          </div>
        )}
      </div>
      {projectName && !projectExists && (
        <button
          type="button"
          className="doc-btn doc-btn-secondary"
          style={{ display: "block", marginTop: 5, padding: "5px 10px", fontSize: "0.85em" }}
          onClick={() => {
            setShowCreateForm((show) => !show);
            if (zoneDefault) setZone(zoneDefault);
          }}
        >
          Crear proyecto
        </button>
      )}
      {showCreateForm && (
        <div style={{ display: "grid", gridTemplateColumns: "minmax(0, 1fr) auto", gap: 8, marginTop: 8, width: "100%" }}>
          <select
            className="doc-input"
            style={{ gridColumn: "1 / -1", width: "100%", minWidth: 0, minHeight: 42, boxSizing: "border-box", fontSize: "0.95em" }}
            value={zone}
            onChange={(event) => setZone(event.target.value)}
            required
          >
            <option value="">Zona del proyecto...</option>
            {zones.map((availableZone) => (
              <option key={availableZone} value={availableZone}>
                {availableZone}
              </option>
            ))}
          </select>
          <button
            type="button"
            className="doc-btn doc-btn-primary"
            style={{ gridColumn: "1 / -1", justifyContent: "center", minHeight: 40, padding: "8px 14px", flexShrink: 0 }}
            onClick={handleCreate}
            disabled={saving || !zone}
          >
            {saving ? "Guardando..." : "Guardar"}
          </button>
        </div>
      )}
      {error && <small role="alert" style={{ display: "block", color: "#b91c1c", marginTop: 4 }}>{error}</small>}
    </div>
  );
}
