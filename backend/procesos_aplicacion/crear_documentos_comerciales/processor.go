package crear_documentos_comerciales

import (
	"errors"
	"time"

	"gorm.io/gorm"
	"sigefae/internal/db"
)

// InboundMetadata contiene los datos de trazabilidad provenientes de la API Inbound
type InboundMetadata struct {
	LDF                string
	EventoDIAN         string
	FechaRecepcionDIAN *time.Time
	RutaXML            string
	RutaPDF            string
}

// ProcesarFacturaInbound registra o actualiza una factura electrónica proveniente directamente de la API DIAN
func ProcesarFacturaInbound(database *gorm.DB, invoice *Invoice, meta InboundMetadata) (*db.DocumentoComercial, error) {
	var resultado *db.DocumentoComercial

	err := database.Transaction(func(tx *gorm.DB) error {
		// 0. Catálogos base requeridos
		var tipoDoc db.TipoDocumento
		if err := tx.Where("nombre = ?", "NIT").FirstOrCreate(&tipoDoc, db.TipoDocumento{Nombre: "NIT", Activo: true}).Error; err != nil {
			return err
		}

		var categoria db.CategoriaProveedor
		if err := tx.Where("nombre = ?", "General").FirstOrCreate(&categoria, db.CategoriaProveedor{Nombre: "General", Descripcion: "Generado automáticamente", Activo: true}).Error; err != nil {
			return err
		}

		var tipoPersona db.TipoPersona
		if err := tx.Where("nombre = ?", "Jurídica").FirstOrCreate(&tipoPersona, db.TipoPersona{Nombre: "Jurídica", Activo: true}).Error; err != nil {
			return err
		}

		var actividad db.ActividadEconomica
		if err := tx.Where("codigo = ?", "0000").FirstOrCreate(&actividad, db.ActividadEconomica{Nombre: "General", Codigo: "0000", Activo: true}).Error; err != nil {
			return err
		}

		var pais db.Pais
		if err := tx.Where("codigo = ?", "CO").FirstOrCreate(&pais, db.Pais{Nombre: "Colombia", Codigo: "CO", Activo: true}).Error; err != nil {
			return err
		}

		var moneda db.Moneda
		if err := tx.Where("codigo = ?", "COP").FirstOrCreate(&moneda, db.Moneda{Nombre: "Peso Colombiano", Codigo: "COP", Activo: true}).Error; err != nil {
			return err
		}

		var area db.Area
		if err := tx.Unscoped().Where("nombre = ?", "General").First(&area).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				area = db.Area{Nombre: "General", Activo: true}
				if err := tx.Create(&area).Error; err != nil {
					return err
				}
			} else {
				return err
			}
		} else if area.DeletedAt.Valid {
			_ = tx.Unscoped().Model(&area).Update("deleted_at", nil)
		}

		// 1. Receptor
		nitReceptor := invoice.AccountingCustomerParty.Party.PartyTaxScheme.CompanyID
		var receptor db.Receptor
		if err := tx.Where("numero_documento = ?", nitReceptor).First(&receptor).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				receptor = db.Receptor{
					Nombre:          invoice.AccountingCustomerParty.Party.PartyTaxScheme.RegistrationName,
					NumeroDocumento: nitReceptor,
					TipoDocumentoID: tipoDoc.ID,
					Activo:          true,
				}
				if err := tx.Create(&receptor).Error; err != nil {
					return err
				}
			} else {
				return err
			}
		}

		// 2. Geografía y Proveedor
		supplierParty := invoice.AccountingSupplierParty.Party
		addr := supplierParty.PhysicalLocation.Address
		nitProveedor := supplierParty.PartyTaxScheme.CompanyID

		var proveedor db.Proveedor
		if err := tx.Where("numero_documento = ?", nitProveedor).First(&proveedor).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				var depto db.Departamento
				tx.Where("nombre = ?", addr.CountrySubentity).FirstOrCreate(&depto, db.Departamento{Nombre: addr.CountrySubentity, PaisID: pais.ID})

				var municipio db.Municipio
				tx.Where("nombre = ?", addr.CityName).FirstOrCreate(&municipio, db.Municipio{Nombre: addr.CityName, DepartamentoID: depto.ID})

				var direccion db.Direccion
				tx.Where("linea_1 = ? AND id_municipio = ?", addr.AddressLine.Line, municipio.ID).FirstOrCreate(&direccion, db.Direccion{
					Nombre:      "Principal",
					Linea1:      addr.AddressLine.Line,
					IDMunicipio: municipio.ID,
				})

				proveedor = db.Proveedor{
					NumeroDocumento:      nitProveedor,
					TipoDocumentoID:      tipoDoc.ID,
					RazonSocial:          supplierParty.PartyTaxScheme.RegistrationName,
					NombreComercial:      supplierParty.PartyName.Name,
					DireccionID:          direccion.ID,
					CategoriaID:          categoria.ID,
					TipoPersonaID:        tipoPersona.ID,
					ActividadEconomicaID: actividad.ID,
					Activo:               true,
				}
				if err := tx.Create(&proveedor).Error; err != nil {
					return err
				}

				if supplierParty.Contact.Name != "" || supplierParty.Contact.ElectronicMail != "" {
					contacto := db.Contacto{
						ProveedorID: proveedor.ID,
						Nombre:      supplierParty.Contact.Name,
						Correo:      supplierParty.Contact.ElectronicMail,
						Telefono:    supplierParty.Contact.Telephone,
						Activo:      true,
					}
					tx.Create(&contacto)
				}

				if supplierParty.PartyTaxScheme.TaxLevelCode != "" {
					resp := db.ResponsabilidadFiscal{
						IDProveedor: proveedor.ID,
						Codigo:      supplierParty.PartyTaxScheme.TaxLevelCode,
						Activo:      true,
					}
					tx.Create(&resp)
				}
			} else {
				return err
			}
		}

		// 3. Validar si ya existe el documento comercial
		var docExistente db.DocumentoComercial
		busqueda := tx.Where("cufe = ?", invoice.UUID)
		if meta.LDF != "" {
			busqueda = tx.Where("cufe = ? OR ldf = ?", invoice.UUID, meta.LDF)
		}
		if err := busqueda.First(&docExistente).Error; err == nil {
			// Si ya existe, actualizamos trazabilidad de la API (evento, LDF, rutas)
			updates := map[string]any{}
			if meta.LDF != "" && docExistente.LDF == nil {
				updates["ldf"] = meta.LDF
			}
			if meta.EventoDIAN != "" && docExistente.EventoDIAN != meta.EventoDIAN {
				updates["evento_dian"] = meta.EventoDIAN
			}
			if meta.RutaXML != "" && docExistente.RutaXML == "" {
				updates["ruta_xml"] = meta.RutaXML
			}
			if meta.RutaPDF != "" && docExistente.RutaPDF == "" {
				updates["ruta_pdf"] = meta.RutaPDF
			}
			if meta.FechaRecepcionDIAN != nil && docExistente.FechaRecepcionDIAN == nil {
				updates["fecha_recepcion_dian"] = meta.FechaRecepcionDIAN
			}
			if len(updates) > 0 {
				_ = tx.Model(&docExistente).Updates(updates).Error
			}
			resultado = &docExistente
			return nil
		}

		// 4. Crear Documento Comercial Inbound
		issueDate, _ := time.Parse("2006-01-02", invoice.IssueDate)
		var dueDate *time.Time
		if invoice.DueDate != "" {
			d, err := time.Parse("2006-01-02", invoice.DueDate)
			if err == nil {
				dueDate = &d
			}
		}

		formaPago := "No definida"
		for _, pm := range invoice.PaymentMeans {
			if pm.ID == "1" {
				formaPago = "Contado"
			} else if pm.ID == "2" {
				formaPago = "Crédito"
			} else if pm.PaymentMeansCode != "" {
				formaPago = "Otro (" + pm.PaymentMeansCode + ")"
			}
			if pm.PaymentDueDate != "" && dueDate == nil {
				d, err := time.Parse("2006-01-02", pm.PaymentDueDate)
				if err == nil {
					dueDate = &d
				}
			}
		}

		subtotal := invoice.LegalMonetaryTotal.LineExtensionAmount
		if subtotal == 0 && len(invoice.InvoiceLines) > 0 {
			for _, line := range invoice.InvoiceLines {
				subtotal += line.LineExtensionAmount
			}
		}

		var iva float64
		for _, tax := range invoice.TaxTotal {
			iva += tax.TaxAmount
		}

		cufe := invoice.UUID
		var ldfPtr *string
		if meta.LDF != "" {
			ldfPtr = &meta.LDF
		}

		doc := db.DocumentoComercial{
			Tipo:               "FACTURA_ELECTRONICA",
			NumeroDocumento:    invoice.ID,
			IDProveedor:        proveedor.ID,
			IDReceptor:         receptor.ID,
			IDArea:             area.ID,
			FechaDocumento:     issueDate,
			FechaVencimiento:   dueDate,
			FormaPago:          formaPago,
			MonedaID:           moneda.ID,
			Subtotal:           subtotal,
			Iva:                iva,
			Total:              invoice.LegalMonetaryTotal.PayableAmount,
			Cufe:               &cufe,
			LDF:                ldfPtr,
			EventoDIAN:         meta.EventoDIAN,
			FechaRecepcionDIAN: meta.FechaRecepcionDIAN,
			RutaXML:            meta.RutaXML,
			RutaPDF:            meta.RutaPDF,
			Activo:             true,
		}

		if err := tx.Create(&doc).Error; err != nil {
			return err
		}

		// 5. Insertar Detalles de Factura
		for _, line := range invoice.InvoiceLines {
			valorUnit := line.Price.PriceAmount
			if line.InvoicedQuantity > 0 {
				valorUnit = line.LineExtensionAmount / line.InvoicedQuantity
			}

			detalle := db.DetalleDocumentoComercial{
				DocumentoComercialID: doc.ID,
				Descripcion:          line.Item.Description,
				Cantidad:             line.InvoicedQuantity,
				ValorUnit:            valorUnit,
				Total:                line.LineExtensionAmount,
				Activo:               true,
			}
			if err := tx.Create(&detalle).Error; err != nil {
				return err
			}
		}

		resultado = &doc
		return nil
	})

	return resultado, err
}
