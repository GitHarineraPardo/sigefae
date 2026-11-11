package proveedor_inbound

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gorm.io/gorm"

	"sigefae/internal/db"
	"sigefae/procesos_aplicacion/crear_documentos_comerciales"
)

type Service struct {
	db             *gorm.DB
	client         *Client
	storageBaseDir string
}

func New(database *gorm.DB, client *Client) *Service {
	return &Service{
		db:             database,
		client:         client,
		storageBaseDir: filepath.Join("storage", "documentos"),
	}
}

func (s *Service) Client() *Client {
	return s.client
}

type SyncStats struct {
	TotalEncontrados int `json:"total_encontrados"`
	NuevosProcesados int `json:"nuevos_procesados"`
	Actualizados     int `json:"actualizados"`
	NoProcesados     int `json:"no_procesados"`
	DetallesFallos   []string `json:"detalles_fallos,omitempty"`
}

// SyncRange sincroniza los documentos recibidos en un rango de fechas dado.
// Si el rango supera los 80 días, divide automáticamente las consultas para respetar el límite de 3 meses de la API.
func (s *Service) SyncRange(startDate, endDate time.Time) (SyncStats, error) {
	var stats SyncStats
	maxWindow := 80 * 24 * time.Hour
	currentStart := startDate

	for currentStart.Before(endDate) {
		currentEnd := currentStart.Add(maxWindow)
		if currentEnd.After(endDate) {
			currentEnd = endDate
		}

		log.Printf("[Proveedor Inbound] Sincronizando ventana: %s hasta %s\n",
			currentStart.Format("2006-01-02"), currentEnd.Format("2006-01-02"))

		if err := s.syncWindow(currentStart, currentEnd, &stats); err != nil {
			log.Printf("[Proveedor Inbound] Error en ventana %s - %s: %v\n",
				currentStart.Format("2006-01-02"), currentEnd.Format("2006-01-02"), err)
		}

		currentStart = currentEnd.Add(time.Second)
	}

	return stats, nil
}

// SyncRecent sincroniza los últimos N días hacia atrás desde hoy
func (s *Service) SyncRecent(dias int) (SyncStats, error) {
	if dias <= 0 {
		dias = 30
	}
	now := time.Now()
	end := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, now.Location())
	startDay := end.AddDate(0, 0, -dias)
	start := time.Date(startDay.Year(), startDay.Month(), startDay.Day(), 0, 0, 0, 0, now.Location())
	return s.SyncRange(start, end)
}

// SyncAuto sincroniza de forma inteligente: si existen facturas en la base de datos,
// busca desde la última fecha de recepción DIAN (con 3 días de margen por actualizaciones de eventos).
// Si la base de datos está vacía, sincroniza los últimos 30 días por defecto.
func (s *Service) SyncAuto() (SyncStats, error) {
	now := time.Now()
	var maxDate *time.Time
	err := s.db.Model(&db.DocumentoComercial{}).
		Where("ldf IS NOT NULL").
		Select("MAX(COALESCE(fecha_recepcion_dian, fecha_documento))").
		Row().
		Scan(&maxDate)

	if err == nil && maxDate != nil && !maxDate.IsZero() && !maxDate.After(now) {
		start := maxDate.AddDate(0, 0, -3)
		log.Printf("[Proveedor Inbound] Sincronización automática desde última factura en BD: %s\n", start.Format("2006-01-02"))
		return s.SyncRange(start, now)
	}

	if maxDate != nil && maxDate.After(now) {
		log.Printf("[Proveedor Inbound] La última fecha registrada (%s) está en el futuro. Se ignora para evitar un rango vacío.\n", maxDate.Format(time.RFC3339))
	}
	log.Println("[Proveedor Inbound] Sincronizando últimos 30 días.")
	return s.SyncRecent(30)
}

func (s *Service) syncWindow(startDate, endDate time.Time, stats *SyncStats) error {
	pageSize := 100
	pageNumber := 1
	totalPages := 1

	for pageNumber <= totalPages {
		req := ListDocumentsRequest{
			PageNumber:           pageNumber,
			PageSize:             pageSize,
			DocumentTypeCode:     "FACTURA-UBL",
			ReceiverStartingDate: startDate.Format("2006-01-02T15:04:05"),
			ReceiverEndingDate:   endDate.Format("2006-01-02T15:04:05"),
		}

		resp, err := s.client.ListDocuments(req)
		if err != nil {
			return fmt.Errorf("error listando página %d: %w", pageNumber, err)
		}

		totalPages = resp.TotalPages
		if len(resp.Items) == 0 {
			break
		}

		stats.TotalEncontrados += len(resp.Items)
		log.Printf("[Proveedor Inbound] Página %d de %d (%d items recibidos)\n",
			resp.PageNumber, resp.TotalPages, len(resp.Items))

		// Filtrar items pendientes de descarga
		var itemsPendientes []ItemDocumento
		for _, item := range resp.Items {
			var doc db.DocumentoComercial
			busqueda := s.db.Where("cufe = ?", item.CUFE)
			if item.LDF != "" {
				busqueda = s.db.Where("cufe = ? OR ldf = ?", item.CUFE, item.LDF)
			}

			if err := busqueda.First(&doc).Error; err == nil {
				// Ya existe: actualizar estado del evento DIAN si cambió
				if item.Evento != "" && doc.EventoDIAN != item.Evento {
					_ = s.db.Model(&doc).Update("evento_dian", item.Evento).Error
					stats.Actualizados++
				}
				continue
			}

			if item.LDF != "" {
				itemsPendientes = append(itemsPendientes, item)
			}
		}

		if len(itemsPendientes) > 0 {
			log.Printf("[Proveedor Inbound] Descargando ZIP para %d facturas nuevas...\n", len(itemsPendientes))
			procesados, noProcesados, detallesFallos, err := s.processBatch(itemsPendientes)
			stats.NuevosProcesados += procesados
			stats.NoProcesados += noProcesados
			stats.DetallesFallos = append(stats.DetallesFallos, detallesFallos...)
			if err != nil {
				log.Printf("[Proveedor Inbound] Error procesando lote de %d facturas: %v\n", len(itemsPendientes), err)
			}
		}

		pageNumber++
	}

	return nil
}

// processBatch descarga los ZIPs en bloques de hasta 15 documentos y los procesa
func (s *Service) processBatch(items []ItemDocumento) (int, int, []string, error) {
	batchSize := 15
	totalProcesados := 0
	totalNoProcesados := 0
	var detallesFallos []string
	var lastErr error
	for i := 0; i < len(items); i += batchSize {
		end := i + batchSize
		if end > len(items) {
			end = len(items)
		}

		subLote := items[i:end]
		var ldfList []string
		itemMap := make(map[string]ItemDocumento)
		for _, it := range subLote {
			ldfList = append(ldfList, it.LDF)
			itemMap[it.LDF] = it
		}

		zipBytes, err := s.client.DownloadZip(ldfList)
		if err != nil {
			log.Printf("[Proveedor Inbound] Error descargando ZIP para lote (%d LDFs): %v\n", len(ldfList), err)
			lastErr = err
			for _, item := range subLote {
				retryZip, retryErr := s.client.DownloadZip([]string{item.LDF})
				if retryErr != nil {
					totalNoProcesados++
					detallesFallos = append(detallesFallos, fmt.Sprintf("%s: descarga ZIP falló: %v", item.NumeroFactura, retryErr))
					log.Printf("[Proveedor Inbound] Error reintentando factura %s: %v\n", item.NumeroFactura, retryErr)
					continue
				}
				retryMap := map[string]ItemDocumento{item.LDF: item}
				n, processErr := s.extractAndProcessZip(retryZip, retryMap)
				totalProcesados += n
				if processErr != nil || n == 0 {
					totalNoProcesados++
					if processErr != nil {
						lastErr = processErr
						detallesFallos = append(detallesFallos, fmt.Sprintf("%s: procesamiento falló: %v", item.NumeroFactura, processErr))
					} else {
						detallesFallos = append(detallesFallos, fmt.Sprintf("%s: el ZIP no contenía un XML procesable", item.NumeroFactura))
					}
				}
			}
			if len(detallesFallos) > 0 {
				log.Printf("[Proveedor Inbound] Detalles de fallos: %s\n", strings.Join(detallesFallos, " | "))
			}
			continue
		}

		n, err := s.extractAndProcessZip(zipBytes, itemMap)
		totalProcesados += n
		if err != nil {
			lastErr = err
			totalNoProcesados += len(subLote) - n
			detallesFallos = append(detallesFallos, fmt.Sprintf("Lote desde %s: %v", subLote[0].NumeroFactura, err))
			log.Printf("[Proveedor Inbound] Error extrayendo lote ZIP: %v\n", err)
		}
	}
	if len(detallesFallos) > 0 {
		lastErr = fmt.Errorf("%s", strings.Join(detallesFallos, " | "))
	}
	return totalProcesados, totalNoProcesados, detallesFallos, lastErr
}

// extractAndProcessZip descomprime el ZIP en memoria, guarda los archivos en storage y registra los documentos comerciales
func (s *Service) extractAndProcessZip(zipBytes []byte, itemMap map[string]ItemDocumento) (int, error) {
	zipReader, err := zip.NewReader(bytes.NewReader(zipBytes), int64(len(zipBytes)))
	if err != nil {
		return 0, fmt.Errorf("error abriendo ZIP: %w", err)
	}

	// Agrupamos archivos por LDF: map[LDF]map[ext]content
	type fileEntry struct {
		name string
		data []byte
	}
	extractedByLDF := make(map[string][]fileEntry)

	for _, file := range zipReader.File {
		if file.FileInfo().IsDir() {
			continue
		}

		rc, err := file.Open()
		if err != nil {
			log.Printf("[Proveedor Inbound] No se pudo abrir archivo %s dentro del zip: %v\n", file.Name, err)
			continue
		}
		data, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			log.Printf("[Proveedor Inbound] Error leyendo archivo %s: %v\n", file.Name, err)
			continue
		}

		// El nombre suele ser: _E{LDF}.xml o _E{LDF}.pdf
		baseName := filepath.Base(file.Name)
		cleanBase := strings.TrimPrefix(baseName, "_E")
		cleanBase = strings.TrimPrefix(cleanBase, "_")

		// Identificar a qué LDF pertenece
		matchedLDF := ""
		for ldf := range itemMap {
			if strings.Contains(baseName, ldf) || strings.Contains(cleanBase, ldf) {
				matchedLDF = ldf
				break
			}
		}

		if matchedLDF != "" {
			extractedByLDF[matchedLDF] = append(extractedByLDF[matchedLDF], fileEntry{
				name: baseName,
				data: data,
			})
		}
	}

	guardados := 0
	// Procesar cada LDF
	for ldf, item := range itemMap {
		files, ok := extractedByLDF[ldf]
		if !ok || len(files) == 0 {
			log.Printf("[Proveedor Inbound] No se encontraron archivos en el ZIP para LDF: %s\n", ldf)
			continue
		}

		// Crear carpeta de destino para el documento en storage/documentos/{CUFE}/
		docFolderKey := item.CUFE
		if docFolderKey == "" {
			docFolderKey = sanitizeFolderName(ldf)
		}
		docDir := filepath.Join(s.storageBaseDir, docFolderKey)
		if err := os.MkdirAll(docDir, 0755); err != nil {
			log.Printf("[Proveedor Inbound] No se pudo crear directorio %s: %v\n", docDir, err)
			continue
		}

		var rutaXML, rutaPDF string
		var xmlFilePath string

		for _, fe := range files {
			destPath := filepath.Join(docDir, fe.name)
			if err := os.WriteFile(destPath, fe.data, 0644); err != nil {
				log.Printf("[Proveedor Inbound] No se pudo escribir archivo %s: %v\n", destPath, err)
				continue
			}

			ext := strings.ToLower(filepath.Ext(fe.name))
			if ext == ".xml" {
				rutaXML = destPath
				xmlFilePath = destPath
			} else if ext == ".pdf" {
				rutaPDF = destPath
			}
		}

		if xmlFilePath == "" {
			log.Printf("[Proveedor Inbound] No se encontró archivo .xml para LDF: %s\n", ldf)
			continue
		}

		// Parsear el XML con el parser UBL 2.1
		invoice, err := crear_documentos_comerciales.ParseXMLFile(xmlFilePath)
		if err != nil {
			log.Printf("[Proveedor Inbound] Error parseando XML %s: %v\n", xmlFilePath, err)
			continue
		}

		var fechaRecep *time.Time
		if item.FechaRecepcion != "" {
			if t, err := time.Parse("2006-01-02", item.FechaRecepcion); err == nil {
				fechaRecep = &t
			} else if t, err := time.Parse(time.RFC3339, item.FechaRecepcion); err == nil {
				fechaRecep = &t
			}
		}

		meta := crear_documentos_comerciales.InboundMetadata{
			LDF:                item.LDF,
			EventoDIAN:         item.Evento,
			FechaRecepcionDIAN: fechaRecep,
			RutaXML:            rutaXML,
			RutaPDF:            rutaPDF,
		}

		docCreado, err := crear_documentos_comerciales.ProcesarFacturaInbound(s.db, invoice, meta)
		if err != nil {
			log.Printf("[Proveedor Inbound] Error registrando factura %s en base de datos: %v\n", item.NumeroFactura, err)
			continue
		}

		guardados++
		log.Printf("[Proveedor Inbound] Factura registrada exitosamente: ID=%d, Doc=%s, CUFE=%s, Evento=%s\n",
			docCreado.ID, docCreado.NumeroDocumento, *docCreado.Cufe, docCreado.EventoDIAN)
	}

	return guardados, nil
}

func sanitizeFolderName(s string) string {
	replacer := strings.NewReplacer("(", "_", ")", "_", ";", "_", "/", "_", "\\", "_", ":", "_")
	return replacer.Replace(s)
}

// StartBackgroundTask inicia la sincronización periódica en segundo plano
func StartBackgroundTask(database *gorm.DB, client *Client, interval time.Duration) {
	srv := New(database, client)
	StartBackgroundTaskWithService(srv, interval)
}

// StartBackgroundTaskWithService inicia la sincronización periódica usando un servicio existente
func StartBackgroundTaskWithService(srv *Service, interval time.Duration) {
	if interval < 5*time.Minute {
		interval = 5 * time.Minute
	}

	log.Printf("[Proveedor Inbound] Servicio en segundo plano iniciado (intervalo: %v)\n", interval)

	// Primera sincronización inteligente al iniciar
	if _, err := srv.SyncAuto(); err != nil {
		log.Printf("[Proveedor Inbound] Error en sincronización inicial: %v\n", err)
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		log.Println("[Proveedor Inbound] Iniciando sincronización periódica...")
		if _, err := srv.SyncAuto(); err != nil {
			log.Printf("[Proveedor Inbound] Error en sincronización periódica: %v\n", err)
		} else {
			log.Println("[Proveedor Inbound] Sincronización periódica finalizada con éxito.")
		}
	}
}
