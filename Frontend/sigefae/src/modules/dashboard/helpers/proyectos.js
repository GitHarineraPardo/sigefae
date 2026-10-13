import { API } from "../constants/api";

export async function crearProyectoCatalogo(nombre, zona, obtenerToken) {
  const response = await fetch(`${API}/proyectos`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${obtenerToken()}`,
    },
    body: JSON.stringify({ nombre: nombre.trim(), zona }),
  });

  const project = await response.json();
  if (!response.ok) {
    throw new Error(project.error || "No se pudo crear el proyecto.");
  }
  if (!project?.id || !project.nombre) {
    throw new Error("El servidor no devolvió el proyecto creado.");
  }

  return project;
}
