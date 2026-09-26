export const catalogoConfig = {
  "tipo-radicacion": { endpoint: "tipo-radicacion", label: "Tipo de Radicación", method: "PUT", fields: ["nombre"] },
  "cargos":          { endpoint: "cargos",          label: "Cargo",              method: "PATCH", fields: ["nombre"] },
  "tipos-pago":      { endpoint: "tipos-pago",      label: "Tipo de Pago",      method: "PATCH", fields: ["nombre"] },
  "metodos-pago":    { endpoint: "metodos-pago",    label: "Método de Pago",    method: "PATCH", fields: ["nombre", "tipo_pago_id"] },
  "areas":           { endpoint: "areas",           label: "Área",              method: "PATCH", fields: ["nombre"] },
  "rutas":           { endpoint: "rutas",           label: "Ruta",              method: "PUT",   fields: ["nombre", "area_id", "zona"] },
  "pasos-ruta":      { endpoint: "pasos-ruta",      label: "Paso de Ruta",      method: "PUT",   fields: ["ruta_id", "orden", "nombre", "cargo_id"] },
  "reglas-monto":    { endpoint: "regla-monto-ruta",label: "Regla de Monto",    method: "PUT",   fields: ["monto_minimo_smmlv", "monto_maximo_smmlv", "posicion_insercion", "prioridad", "cargo_aprobador_id"] },
  "normas-reparto":  { endpoint: "normas-reparto",  label: "Norma de Reparto",  method: "PUT",   fields: ["codigo", "nombre", "proyecto", "sucursal", "departamento", "tipo", "tarifa_iva", "descripcion"] },
  "salario-minimo":  { endpoint: "salario-minimo",  label: "Salario Mínimo",    method: "PUT",   fields: ["ano", "valor"] },
  "proyectos":       { endpoint: "proyectos",       label: "Proyecto",          method: "PUT",   fields: ["nombre", "zona"] },
};