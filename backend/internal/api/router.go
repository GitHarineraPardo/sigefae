package api

import (
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"sigefae/internal/actividad_economica"
	"sigefae/internal/archivo"
	"sigefae/internal/archivo_origen"
	"sigefae/internal/area"
	"sigefae/internal/auth"
	"sigefae/internal/cargo"
	"sigefae/internal/categoria_proveedor"
	"sigefae/internal/codigo_qr"
	"sigefae/internal/comentario"
	"sigefae/internal/contacto"
	"sigefae/internal/correo"
	"sigefae/internal/db"
	"sigefae/internal/departamento"
	"sigefae/internal/detalle_documento_comercial"
	"sigefae/internal/direccion"
	"sigefae/internal/documento_comercial"
	"sigefae/internal/documento_radicado"
	"sigefae/internal/estado_correo"
	"sigefae/internal/estado_documento_radicado"
	"sigefae/internal/estado_tarea"
	"sigefae/internal/metodo_pago"
	"sigefae/internal/moneda"
	"sigefae/internal/municipio"
	"sigefae/internal/norma_reparto"
	"sigefae/internal/notificacion"
	"sigefae/internal/pais"
	"sigefae/internal/paso_ruta"
	"sigefae/internal/proveedor"
	"sigefae/internal/proyecto"
	"sigefae/internal/receptor"
	"sigefae/internal/registro_aprobacion"
	"sigefae/internal/regla_monto_ruta"
	"sigefae/internal/responsabilidad_fiscal"
	"sigefae/internal/role"
	"sigefae/internal/ruta"
	"sigefae/internal/salario_minimo"
	"sigefae/internal/tarea"
	"sigefae/internal/tipo_documento"
	"sigefae/internal/tipo_factura"
	"sigefae/internal/tipo_pago"
	"sigefae/internal/tipo_persona"
	"sigefae/internal/tipo_radicacion"
	"sigefae/internal/trazabilidad"
	"sigefae/internal/user"
)

func New(database *gorm.DB) *gin.Engine {

	router := gin.Default()
	router.Use(cors.New(cors.Config{
		AllowOrigins: []string{
			"http://localhost:5173",
			"https://app.harinerapardo.co",
		},
		AllowMethods: []string{
			"GET",
			"POST",
			"PUT",
			"PATCH",
			"DELETE",
			"OPTIONS",
		},
		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Authorization",
		},
		AllowCredentials: true,
	}))

	// ==========================
	// No Cache Middleware
	// ==========================
	router.Use(func(c *gin.Context) {
		c.Header("Cache-Control", "no-store, no-cache, must-revalidate, proxy-revalidate")
		c.Header("Pragma", "no-cache")
		c.Header("Expires", "0")
		c.Next()
	})

	// ==========================
	// Health Check
	// ==========================

	router.GET("/ping", func(c *gin.Context) {

		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})

	})

	// ==========================
	// Auth
	// ==========================

	authService := auth.New(database)
	authHandler := auth.NewHandler(authService)

	userService := user.New(database)
	userHandler := user.NewHandler(userService)
	roleService := role.New(database)
	roleHandler := role.NewHandler(roleService)

	tipoPagoService := tipo_pago.New(database)
	tipoPagoHandler := tipo_pago.NewHandler(tipoPagoService)

	metodoPagoService := metodo_pago.New(database)
	metodoPagoHandler := metodo_pago.NewHandler(metodoPagoService)

	areaService := area.New(database)
	areaHandler := area.NewHandler(areaService)

	tipoFacturaService := tipo_factura.New(database)
	tipoFacturaHandler := tipo_factura.NewHandler(tipoFacturaService)

	rutaService := ruta.New(database)
	rutaHandler := ruta.NewHandler(rutaService)

	pasoRutaService := paso_ruta.New(database)
	pasoRutaHandler := paso_ruta.NewHandler(pasoRutaService)

	monedaService := moneda.New(database)
	monedaHandler := moneda.NewHandler(monedaService)

	paisService := pais.New(database)
	paisHandler := pais.NewHandler(paisService)

	departamentoService := departamento.New(database)
	departamentoHandler := departamento.NewHandler(departamentoService)

	municipioService := municipio.New(database)
	municipioHandler := municipio.NewHandler(municipioService)

	direccionService := direccion.New(database)
	direccionHandler := direccion.NewHandler(direccionService)

	actividadEconomicaService := actividad_economica.New(database)
	actividadEconomicaHandler := actividad_economica.NewHandler(actividadEconomicaService)

	tipoDocumentoService := tipo_documento.New(database)
	tipoDocumentoHandler := tipo_documento.NewHandler(tipoDocumentoService)

	categoriaProveedorService := categoria_proveedor.New(database)
	categoriaProveedorHandler := categoria_proveedor.NewHandler(categoriaProveedorService)

	tipoPersonaService := tipo_persona.New(database)
	tipoPersonaHandler := tipo_persona.NewHandler(tipoPersonaService)

	proveedorService := proveedor.New(database)
	proveedorHandler := proveedor.NewHandler(proveedorService)

	contactoService := contacto.New(database)
	contactoHandler := contacto.NewHandler(contactoService)

	responsabilidadFiscalService := responsabilidad_fiscal.New(database)
	responsabilidadFiscalHandler := responsabilidad_fiscal.NewHandler(responsabilidadFiscalService)

	receptorService := receptor.New(database)
	receptorHandler := receptor.NewHandler(receptorService)

	estadoCorreoService := estado_correo.New(database)
	estadoCorreoHandler := estado_correo.NewHandler(estadoCorreoService)

	estadoDocumentoRadicadoService := estado_documento_radicado.New(database)
	estadoDocumentoRadicadoHandler := estado_documento_radicado.NewHandler(estadoDocumentoRadicadoService)

	tipoRadicacionService := tipo_radicacion.New(database)
	tipoRadicacionHandler := tipo_radicacion.NewHandler(tipoRadicacionService)

	estadoTareaService := estado_tarea.New(database)
	estadoTareaHandler := estado_tarea.NewHandler(estadoTareaService)

	archivoOrigenService := archivo_origen.New(database)
	archivoOrigenHandler := archivo_origen.NewHandler(archivoOrigenService)

	correoService := correo.New(database)
	correoHandler := correo.NewHandler(correoService)

	documentoComercialService := documento_comercial.New(database)
	documentoComercialHandler := documento_comercial.NewHandler(documentoComercialService)

	detalleDocumentoComercialService := detalle_documento_comercial.New(database)
	detalleDocumentoComercialHandler := detalle_documento_comercial.NewHandler(detalleDocumentoComercialService)

	codigoQrService := codigo_qr.New(database)
	codigoQrHandler := codigo_qr.NewHandler(codigoQrService)

	// ==========================
	// NOTIFICACION PRIMERO (lo usan otros handlers)
	// ==========================
	notificacionService := notificacion.New(database)
	notificacionHandler := notificacion.NewHandler(notificacionService)

	// ==========================
	// Documento Radicado (necesita notificacion)
	// ==========================
	documentoRadicadoService := documento_radicado.New(database)
	documentoRadicadoHandler := documento_radicado.NewHandler(documentoRadicadoService, notificacionService, database)

	// ==========================
	// Comentario (necesita notificacion)
	// ==========================
	comentarioService := comentario.New(database)
	comentarioHandler := comentario.NewHandler(comentarioService, notificacionService, database)

	// ==========================
	// Tarea (necesita notificacion)
	// ==========================
	tareaService := tarea.New(database)
	tareaHandler := tarea.NewHandler(tareaService, notificacionService, database)

	trazabilidadService := trazabilidad.New(database)
	trazabilidadHandler := trazabilidad.NewHandler(trazabilidadService)

	archivoService := archivo.New(database)
	archivoHandler := archivo.NewHandler(archivoService, database, notificacionService)

	registroAprobacionService := registro_aprobacion.New(database)
	registroAprobacionHandler := registro_aprobacion.NewHandler(registroAprobacionService)

	reglaMontoRutaService := regla_monto_ruta.New(database)
	reglaMontoRutaHandler := regla_monto_ruta.NewHandler(reglaMontoRutaService)

	normaRepartoService := norma_reparto.New(database)
	normaRepartoHandler := norma_reparto.NewHandler(normaRepartoService)

	proyectoService := proyecto.New(database)
	proyectoHandler := proyecto.NewHandler(proyectoService)

	salarioMinimoService := salario_minimo.New(database)
	salarioMinimoHandler := salario_minimo.NewHandler(salarioMinimoService)

	cargoService := cargo.New(database)
	cargoHandler := cargo.NewHandler(cargoService)

	api := router.Group("/api")
	{
		api.POST("/auth/login", authHandler.Login)
		api.POST("/auth/sso", authHandler.SSO) // Nuevo endpoint para Single Sign-On desde PHP
		api.GET("/documentoradicado/verificar/:numero_radicado", documentoRadicadoHandler.VerificarPublico)
	}
	// =========================
	// Rutas protegidas
	// =========================

	protected := router.Group("/api")
	protected.Use(auth.Middleware(database))
	{
		// ── Cualquier usuario logueado puede leer radicados ──
		protected.GET("/documentoradicado", documentoRadicadoHandler.List)
		protected.GET("/documentoradicado/:id", documentoRadicadoHandler.GetByID)
		protected.PATCH("/documentoradicado/:id/causar", documentoRadicadoHandler.Causar)
		protected.PATCH("/documentoradicado/:id/pagar", documentoRadicadoHandler.Pagar)
		protected.PATCH("/documentoradicado/:id/comprobantes_subidos", documentoRadicadoHandler.ToggleComprobantesSubidos)
		// ── Cualquier usuario logueado puede subir/leer anexos ──
		protected.POST("/documentoradicado/:id/anexos", archivoHandler.UploadAnexo)
		protected.GET("/archivo", archivoHandler.List)
		protected.GET("/archivo/:id/download", archivoHandler.Download)
		// tarea
		protected.GET("/documentoradicado/:id/tareas", tareaHandler.ListByRadicado)
		protected.PATCH("/tarea/:id/completar", tareaHandler.Completar)
		protected.POST("/tarea/:id/devolver", tareaHandler.Devolver)
		// moneda list
		protected.GET("/monedas", monedaHandler.List)
		// =========================
		// Comentario
		// =========================
		protected.POST("/comentario", comentarioHandler.Create)
		protected.GET("/comentario", comentarioHandler.List)

		protected.PATCH("/archivo/:id/reemplazar", archivoHandler.Reemplazar)
		// ── Cualquier usuario logueado puede eliminar sus propios anexos (handler valida propietario) ──
		protected.DELETE("/archivo/:id", archivoHandler.Delete)

		// ── Rechazar documento (Superadministrador y Contabilidad — validado en handler) ──
		protected.POST("/documentoradicado/:id/rechazar", documentoRadicadoHandler.Rechazar)

		// =========================
		// Trazabilidad
		// =========================
		protected.GET("/trazabilidad", trazabilidadHandler.List)
		protected.GET("/trazabilidad/por-area", trazabilidadHandler.ListPorArea)

		protected.GET("/rutas", rutaHandler.List)
		protected.GET("/areas", areaHandler.List) // lectura para todos (Contabilidad necesita para trazabilidad por área)
		protected.GET("/cargos", cargoHandler.ListActivos)
		protected.GET("/proveedor/:id/normas-reparto", proveedorHandler.ListNormasReparto)
		protected.GET("/normas-reparto", normaRepartoHandler.List)
		protected.GET("/normas-reparto/:id", normaRepartoHandler.GetByID)
		protected.GET("/proyectos", proyectoHandler.List)
		protected.GET("/cargos/activos", cargoHandler.ListActivos)
		// =========================
		// Notificacion (usuario logueado)
		// =========================
		protected.POST("/notificacion", notificacionHandler.Create)
		protected.GET("/notificacion/mias", notificacionHandler.ListByUsuario)
		protected.PATCH("/notificacion/:id/leida", notificacionHandler.MarkAsRead)
		protected.DELETE("/notificacion/:id", notificacionHandler.Delete)

		protected.GET("/documentoradicado/:id/normas-reparto", documentoRadicadoHandler.GetNormasReparto)
		protected.POST("/documentoradicado/:id/normas-reparto", documentoRadicadoHandler.AsignarNormasReparto)
		protected.GET("/documentoradicado/:id/activos-fijos", documentoRadicadoHandler.GetActivosFijos)
		protected.POST("/documentoradicado/:id/activos-fijos", documentoRadicadoHandler.AsignarActivosFijos)
		protected.POST("/documentoradicado/:id/solicitar-rechazo", documentoRadicadoHandler.SolicitarRechazo)
		protected.GET("/solicitud-rechazo/mias", documentoRadicadoHandler.ListMySolicitudes)
		protected.POST("/documentoradicado/:id/solicitar-cambio-norma", documentoRadicadoHandler.SolicitarCambioNormaReparto)
		protected.GET("/solicitud-cambio-norma/mias", documentoRadicadoHandler.ListMisSolicitudesCambioNorma)
		protected.POST("/solicitud-permiso", documentoRadicadoHandler.CrearSolicitudPermiso)
		protected.GET("/solicitud-permiso/mias", documentoRadicadoHandler.ListMisSolicitudesPermiso)
		protected.GET("/solicitud-permiso", documentoRadicadoHandler.ListSolicitudesPermiso)
		protected.POST("/solicitud-permiso/:id/decidir", documentoRadicadoHandler.DecidirSolicitudPermiso)

		//documento comercial
		protected.GET("/documentocomercial/:id", documentoComercialHandler.GetByID)

		protected.GET("/me", func(c *gin.Context) {
			user := c.MustGet("user").(db.Usuario)
			cargoNombre := ""
			if user.Cargo != nil {
				cargoNombre = user.Cargo.Nombre
			}
			c.JSON(http.StatusOK, gin.H{
				"id":       user.ID,
				"nombre":   user.Nombre,
				"email":    user.Email,
				"cargo_id": user.CargoID,
				"cargo":    cargoNombre,
				"rol":      user.Rol.Nombre,
			})
		})

		// =========================
		// Solo Superadministrador
		// =========================
		admin := protected.Group("")
		admin.Use(auth.RequireRole("Superadministrador"))
		{
			// =========================
			// Rechazar documento (marcar como Devuelto / rechazado)
			admin.POST("/documentoradicado/:id/completar", documentoRadicadoHandler.MarcarCompletado)
			admin.GET("/solicitud-rechazo", documentoRadicadoHandler.ListSolicitudes)
			admin.POST("/solicitud-rechazo/:id/decidir", documentoRadicadoHandler.DecidirSolicitud)
			admin.GET("/solicitud-cambio-norma", documentoRadicadoHandler.ListSolicitudesCambioNorma)
			admin.POST("/solicitud-cambio-norma/:id/decidir", documentoRadicadoHandler.DecidirSolicitudCambioNorma)
			// =========================
			// Usuarios
			// =========================
			admin.POST("/usuarios", userHandler.Create)
			admin.GET("/usuarios", userHandler.List)
			admin.GET("/cargos/todos", cargoHandler.List)
			admin.POST("/cargos", cargoHandler.Create)
			admin.PATCH("/cargos/:id", cargoHandler.Update)
			admin.PATCH("/cargos/:id/activo", cargoHandler.UpdateStatus)
			admin.DELETE("/cargos/:id", cargoHandler.Delete)
			admin.GET("/usuarios/:id", userHandler.GetByID)
			admin.PUT("/usuarios/:id", userHandler.Update)
			admin.PATCH("/usuarios/:id/activo", userHandler.UpdateStatus)
			admin.PATCH("/usuarios/:id/password", userHandler.UpdatePassword)

			// =========================
			// Roles
			// =========================
			admin.POST("/roles", roleHandler.Create)
			admin.GET("/roles", roleHandler.List)
			admin.PATCH("/roles/:id", roleHandler.Update)
			admin.PATCH("/roles/:id/activo", roleHandler.UpdateStatus)

			// =========================
			// Tipos de Pago
			// =========================
			admin.POST("/tipos-pago", tipoPagoHandler.Create)
			admin.GET("/tipos-pago", tipoPagoHandler.List)
			admin.PATCH("/tipos-pago/:id", tipoPagoHandler.Update)
			admin.PATCH("/tipos-pago/:id/activo", tipoPagoHandler.UpdateStatus)
			admin.DELETE("/tipos-pago/:id", tipoPagoHandler.Delete)

			// =========================
			// Métodos de Pago
			// =========================
			admin.POST("/metodos-pago", metodoPagoHandler.Create)
			admin.GET("/metodos-pago", metodoPagoHandler.List)
			admin.PATCH("/metodos-pago/:id", metodoPagoHandler.Update)
			admin.PATCH("/metodos-pago/:id/activo", metodoPagoHandler.UpdateStatus)
			admin.DELETE("/metodos-pago/:id", metodoPagoHandler.Delete)

			// =========================
			// Áreas (escritura solo admin, lectura en protected)
			// =========================
			admin.POST("/areas", areaHandler.Create)
			admin.PATCH("/areas/:id/activo", areaHandler.UpdateStatus)
			admin.PATCH("/areas/:id", areaHandler.Update)
			admin.DELETE("/areas/:id", areaHandler.Delete)

			// =========================
			// Tipos de Factura
			// =========================
			admin.POST("/tipos-factura", tipoFacturaHandler.Create)
			admin.GET("/tipos-factura", tipoFacturaHandler.List)
			admin.PUT("/tipos-factura/:id", tipoFacturaHandler.Update)
			admin.PATCH("/tipos-factura/:id/activo", tipoFacturaHandler.UpdateStatus)

			// =========================
			// Rutas
			// =========================
			admin.POST("/rutas", rutaHandler.Create)
			admin.PUT("/rutas/:id", rutaHandler.Update)
			admin.PATCH("/rutas/:id/activo", rutaHandler.UpdateStatus)
			admin.DELETE("/rutas/:id", rutaHandler.Delete)

			// =========================
			// Pasos de Ruta
			// =========================
			admin.POST("/pasos-ruta", pasoRutaHandler.Create)
			admin.GET("/pasos-ruta", pasoRutaHandler.List)
			admin.PUT("/pasos-ruta/:id", pasoRutaHandler.Update)
			admin.PATCH("/pasos-ruta/:id/activo", pasoRutaHandler.UpdateStatus)
			admin.DELETE("/pasos-ruta/:id", pasoRutaHandler.Delete)

			// =========================
			// Moneda
			// =========================
			admin.POST("/moneda", monedaHandler.Create)
			admin.PUT("/moneda/:id", monedaHandler.Update)
			admin.PATCH("/moneda/:id/activo", monedaHandler.UpdateStatus)

			// =========================
			// Pais
			// =========================
			admin.POST("/pais", paisHandler.Create)
			admin.GET("/pais", paisHandler.List)
			admin.PUT("/pais/:id", paisHandler.Update)
			admin.PATCH("/pais/:id/activo", paisHandler.UpdateStatus)

			// =========================
			// Departamento
			// =========================
			admin.POST("/departamento", departamentoHandler.Create)
			admin.GET("/departamento", departamentoHandler.List)
			admin.PUT("/departamento/:id", departamentoHandler.Update)
			admin.PATCH("/departamento/:id/activo", departamentoHandler.UpdateStatus)

			// =========================
			// Municipio
			// =========================
			admin.POST("/municipio", municipioHandler.Create)
			admin.GET("/municipio", municipioHandler.List)
			admin.PUT("/municipio/:id", municipioHandler.Update)
			admin.PATCH("/municipio/:id/activo", municipioHandler.UpdateStatus)

			// =========================
			// Direccion
			// =========================
			admin.POST("/direccion", direccionHandler.Create)
			admin.GET("/direccion", direccionHandler.List)
			admin.PUT("/direccion/:id", direccionHandler.Update)
			admin.PATCH("/direccion/:id/activo", direccionHandler.UpdateStatus)

			// =========================
			// Actividad Economica
			// =========================
			admin.POST("/actividad-economica", actividadEconomicaHandler.Create)
			admin.GET("/actividad-economica", actividadEconomicaHandler.List)
			admin.PUT("/actividad-economica/:id", actividadEconomicaHandler.Update)
			admin.PATCH("/actividad-economica/:id/activo", actividadEconomicaHandler.UpdateStatus)

			// =========================
			// Tipo documento
			// =========================
			admin.POST("/tipo-documento", tipoDocumentoHandler.Create)
			admin.GET("/tipo-documento", tipoDocumentoHandler.List)
			admin.PUT("/tipo-documento/:id", tipoDocumentoHandler.Update)
			admin.PATCH("/tipo-documento/:id/activo", tipoDocumentoHandler.UpdateStatus)

			// =========================
			// Categoria proveedor
			// =========================
			admin.POST("/categoria-proveedor", categoriaProveedorHandler.Create)
			admin.GET("/categoria-proveedor", categoriaProveedorHandler.List)
			admin.PUT("/categoria-proveedor/:id", categoriaProveedorHandler.Update)
			admin.PATCH("/categoria-proveedor/:id/activo", categoriaProveedorHandler.UpdateStatus)

			// =========================
			// Tipo Persona
			// =========================
			admin.POST("/tipo-persona", tipoPersonaHandler.Create)
			admin.GET("/tipo-persona", tipoPersonaHandler.List)
			admin.PUT("/tipo-persona/:id", tipoPersonaHandler.Update)
			admin.PATCH("/tipo-persona/:id/activo", tipoPersonaHandler.UpdateStatus)

			// =========================
			// Proveedor
			// =========================
			admin.POST("/proveedor", proveedorHandler.Create)
			admin.GET("/proveedor", proveedorHandler.List)
			admin.PUT("/proveedor/:id", proveedorHandler.Update)
			admin.PATCH("/proveedor/:id/activo", proveedorHandler.UpdateStatus)

			// =========================
			// Contacto
			// =========================
			admin.POST("/contacto", contactoHandler.Create)
			admin.GET("/contacto", contactoHandler.List)
			admin.PUT("/contacto/:id", contactoHandler.Update)
			admin.PATCH("/contacto/:id/activo", contactoHandler.UpdateStatus)

			// =========================
			// Responsabilidad Fiscal
			// =========================
			admin.POST("/responsabilidad-fiscal", responsabilidadFiscalHandler.Create)
			admin.GET("/responsabilidad-fiscal", responsabilidadFiscalHandler.List)
			admin.PUT("/responsabilidad-fiscal/:id", responsabilidadFiscalHandler.Update)
			admin.PATCH("/responsabilidad-fiscal/:id/activo", responsabilidadFiscalHandler.UpdateStatus)

			// =========================
			// Receptor
			// =========================
			admin.POST("/receptor", receptorHandler.Create)
			admin.GET("/receptor", receptorHandler.List)
			admin.PUT("/receptor/:id", receptorHandler.Update)
			admin.PATCH("/receptor/:id/activo", receptorHandler.UpdateStatus)

			// =========================
			// Estado Correo
			// =========================
			admin.POST("/estado-correo", estadoCorreoHandler.Create)
			admin.GET("/estado-correo", estadoCorreoHandler.List)
			admin.PUT("/estado-correo/:id", estadoCorreoHandler.Update)
			admin.PATCH("/estado-correo/:id/activo", estadoCorreoHandler.UpdateStatus)

			// =========================
			// Estado Documento Radicado
			// =========================
			admin.POST("/estado-documento-radicado", estadoDocumentoRadicadoHandler.Create)
			admin.GET("/estado-documento-radicado", estadoDocumentoRadicadoHandler.List)
			admin.PUT("/estado-documento-radicado/:id", estadoDocumentoRadicadoHandler.Update)
			admin.PATCH("/estado-documento-radicado/:id/activo", estadoDocumentoRadicadoHandler.UpdateStatus)

			// =========================
			// Tipo Radicación
			// =========================
			admin.POST("/tipo-radicacion", tipoRadicacionHandler.Create)
			admin.GET("/tipo-radicacion", tipoRadicacionHandler.List)
			admin.PUT("/tipo-radicacion/:id", tipoRadicacionHandler.Update)
			admin.PATCH("/tipo-radicacion/:id/activo", tipoRadicacionHandler.UpdateStatus)
			admin.DELETE("/tipo-radicacion/:id", tipoRadicacionHandler.Delete)

			// =========================
			// Estado Tarea
			// =========================
			admin.POST("/estado-tarea", estadoTareaHandler.Create)
			admin.GET("/estado-tarea", estadoTareaHandler.List)
			admin.PUT("/estado-tarea/:id", estadoTareaHandler.Update)
			admin.PATCH("/estado-tarea/:id/activo", estadoTareaHandler.UpdateStatus)

			// =========================
			// Archivo Origen
			// =========================
			admin.POST("/archivo-origen", archivoOrigenHandler.Create)
			admin.GET("/archivo-origen", archivoOrigenHandler.List)
			admin.PUT("/archivo-origen/:id", archivoOrigenHandler.Update)
			admin.PATCH("/archivo-origen/:id/activo", archivoOrigenHandler.UpdateStatus)

			// =========================
			// Correos
			// =========================
			admin.POST("/correo", correoHandler.Create)
			admin.GET("/correo", correoHandler.List)
			admin.GET("/correo/:id", correoHandler.GetByID)
			admin.PUT("/correo/:id", correoHandler.Update)
			admin.PATCH("/correo/:id/estado", correoHandler.UpdateStatus)
			admin.PATCH("/correos/:id/activo", correoHandler.Delete)

			// =========================
			// Documento Comercial
			// =========================
			admin.POST("/documentocomercial", documentoComercialHandler.Create)
			admin.GET("/documentocomercial", documentoComercialHandler.List)
			admin.GET("/documentocomercial/pendientes", documentoComercialHandler.ListPendientes)
			admin.PUT("/documentocomercial/:id", documentoComercialHandler.Update)
			admin.PATCH("/documentocomercial/:id/activo", documentoComercialHandler.Delete)

			// ==============================
			// Detalle Documento Comercial
			// ==============================
			admin.POST("/detalledocumentocomercial", detalleDocumentoComercialHandler.Create)
			admin.GET("/detalledocumentocomercial", detalleDocumentoComercialHandler.List)
			admin.PUT("/detalledocumentocomercial/:id", detalleDocumentoComercialHandler.Update)
			admin.PATCH("/detalledocumentocomercial/:id/activo", detalleDocumentoComercialHandler.Delete)

			// ===============
			// Codigo QR
			// ===============
			admin.POST("/codigoqr", codigoQrHandler.Create)
			admin.GET("/codigoqr", codigoQrHandler.List)
			admin.PUT("/codigoqr/:id", codigoQrHandler.Update)
			admin.PATCH("/codigoqr/:id/activo", codigoQrHandler.Delete)

			// =========================
			// Documento Radicado
			// =========================
			admin.POST("/documentoradicado", documentoRadicadoHandler.Create)
			admin.PUT("/documentoradicado/:id", documentoRadicadoHandler.Update)

			// =========================
			// Notificacion (admin - listar todas)
			// =========================
			admin.GET("/notificacion", notificacionHandler.List)
			admin.PUT("/notificacion/:id", notificacionHandler.Update)

			// =========================
			// Trazabilidad
			// =========================
			admin.POST("/trazabilidad", trazabilidadHandler.Create)

			// =========================
			// Archivo
			// =========================
			admin.POST("/archivo", archivoHandler.Create)
			admin.PUT("/archivo/:id", archivoHandler.Update)

			// =========================
			// Tarea
			// =========================
			admin.POST("/tarea", tareaHandler.Create)
			admin.GET("/tarea", tareaHandler.List)
			admin.PUT("/tarea/:id", tareaHandler.Update)

			// =========================
			// Registro Aprobación
			// =========================
			admin.POST("/registroaprobacion", registroAprobacionHandler.Create)
			admin.GET("/registroaprobacion", registroAprobacionHandler.List)

			// =========================
			// Regla Monto Ruta
			// =========================
			admin.POST("/regla-monto-ruta", reglaMontoRutaHandler.Create)
			admin.GET("/regla-monto-ruta", reglaMontoRutaHandler.List)
			admin.PUT("/regla-monto-ruta/:id", reglaMontoRutaHandler.Update)
			admin.PATCH("/regla-monto-ruta/:id/activo", reglaMontoRutaHandler.UpdateStatus)
			admin.DELETE("/regla-monto-ruta/:id", reglaMontoRutaHandler.Delete)

			// Normas de Reparto
			admin.POST("/normas-reparto", normaRepartoHandler.Create)
			admin.PUT("/normas-reparto/:id", normaRepartoHandler.Update)
			admin.PATCH("/normas-reparto/:id/activo", normaRepartoHandler.UpdateStatus)
			admin.DELETE("/normas-reparto/:id", normaRepartoHandler.Delete)

			// Proyectos
			admin.POST("/proyectos", proyectoHandler.Create)
			admin.PUT("/proyectos/:id", proyectoHandler.Update)
			admin.PATCH("/proyectos/:id/activo", proyectoHandler.UpdateStatus)
			admin.DELETE("/proyectos/:id", proyectoHandler.Delete)

			// Salario Mínimo
			admin.POST("/salario-minimo", salarioMinimoHandler.Create)
			admin.GET("/salario-minimo", salarioMinimoHandler.List)
			admin.PUT("/salario-minimo/:id", salarioMinimoHandler.Update)
			admin.PATCH("/salario-minimo/:id/activo", salarioMinimoHandler.UpdateStatus)
			admin.DELETE("/salario-minimo/:id", salarioMinimoHandler.Delete)

		}
	}

	// =========================
	// Servir archivos estáticos
	// =========================
	router.Static("/api/storage", "storage")

	return router
}
