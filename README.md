# SIGEFAE

SIGEFAE es una aplicación para la gestión de documentos, radicados, tareas,
aprobaciones, gastos por proveedor y eventos electrónicos DIAN.

## Requisitos

- Go 1.25 o superior
- Node.js compatible con las versiones declaradas en `Frontend/sigefae/package.json`
- MySQL configurado para el backend

## Configuración

1. Copia `backend/.env.example` a `backend/.env`.
2. Completa las variables de base de datos, JWT, CORS, correo y proveedor
   tecnológico.
3. Copia `Frontend/sigefae/.env.example` a `Frontend/sigefae/.env` y define
   `VITE_API_URL`.

No agregues archivos `.env` al repositorio. Las credenciales y tokens deben
permanecer únicamente en el entorno de ejecución.

## Ejecución local

En una terminal:

```bash
cd backend
go run ./cmd
```

En otra terminal:

```bash
cd Frontend/sigefae
npm ci
npm run dev
```

El backend escucha por defecto en el puerto `8080`. El frontend utiliza la
base `/sigefae/Frontend/`, configurada en `vite.config.js`.

## Validación

```bash
cd backend
go test ./...

cd ../Frontend/sigefae
npm run lint
npm run build
```

## Estructura

- `backend/`: API Go, persistencia, autenticación, procesos en segundo plano e
  integración con el proveedor tecnológico.
- `Frontend/sigefae/`: aplicación React/Vite.
- `backend/internal/evento_dian/`: consulta y gestión de eventos DIAN.
- `backend/internal/gasto_proveedor/`: gestión de gastos por proveedor.
- `backend/procesos_aplicacion/proveedor_inbound/`: sincronización de archivos y
  eventos desde el proveedor tecnológico.

Las migraciones se ejecutan al iniciar el backend. La ejecución en producción
debe usar un gestor de procesos, secretos administrados y una base de datos
respaldada.
