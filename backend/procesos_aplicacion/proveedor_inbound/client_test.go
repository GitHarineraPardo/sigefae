package proveedor_inbound

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_Login_And_ListDocuments(t *testing.T) {
	expectedToken := "test-token-12345"
	expectedLDF := "FACTURA-UBL(901297346;SETT2626513;2026-10-26;PRINCIPAL;PRINCIPAL)"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/Auth/Login":
			if r.Header.Get("x-who") != "who-auth" {
				t.Errorf("x-who header esperado 'who-auth', obtenido '%s'", r.Header.Get("x-who"))
			}

			var req LoginRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Fatalf("error decodificando login req: %v", err)
			}
			if req.U != "testuser" || req.P != "testpass" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(LoginResponse{
				DisplayName: "Test User",
				AccessToken: expectedToken,
				TenantID:    "tenant-001",
			})

		case "/InboundDocument/Integrated/Document":
			if r.Header.Get("x-who") != "who-list" {
				t.Errorf("x-who header esperado 'who-list', obtenido '%s'", r.Header.Get("x-who"))
			}
			if r.Header.Get("Authorization") != "Bearer "+expectedToken {
				t.Errorf("Authorization header esperado 'Bearer %s', obtenido '%s'", expectedToken, r.Header.Get("Authorization"))
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(ListDocumentsResponse{
				PageNumber:     1,
				PageSize:       100,
				TotalPages:     1,
				TotalItemCount: 1,
				Items: []ItemDocumento{
					{
						NumeroFactura:  "SETT2626513",
						NIT:            "901297346",
						RazonSocial:    "USALUX DISTRIBUTION SAS",
						LDF:            expectedLDF,
						CUFE:           "cufe-12345",
						Evento:         "ACKNOWLEDGED",
						FechaFactura:   "2026-10-26",
						FechaRecepcion: "2026-10-26",
						Monto:          "4899999.93",
						IVA:            "782352.93",
					},
				},
			})

		case "/Inbound/Download/Zip":
			if r.Header.Get("x-who") != "who-list" {
				t.Errorf("x-who header esperado 'who-list', obtenido '%s'", r.Header.Get("x-who"))
			}

			var ldfs []string
			if err := json.NewDecoder(r.Body).Decode(&ldfs); err != nil {
				t.Fatalf("error decodificando download zip req: %v", err)
			}
			if len(ldfs) != 1 || ldfs[0] != expectedLDF {
				t.Errorf("LDF inesperado en descarga zip: %v", ldfs)
			}

			// Crear un ZIP en memoria
			buf := new(bytes.Buffer)
			zw := zip.NewWriter(buf)
			f, _ := zw.Create("_E" + expectedLDF + ".xml")
			f.Write([]byte("<Invoice><ID>SETT2626513</ID></Invoice>"))
			zw.Close()

			w.Header().Set("Content-Type", "application/zip")
			w.Write(buf.Bytes())

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client := NewClient(server.URL, "testuser", "testpass", "who-list", "who-auth")

	// 1. Probar Login
	loginResp, err := client.Login()
	if err != nil {
		t.Fatalf("Login falló: %v", err)
	}
	if loginResp.AccessToken != expectedToken {
		t.Errorf("Token esperado '%s', obtenido '%s'", expectedToken, loginResp.AccessToken)
	}

	// 2. Probar ListDocuments
	listResp, err := client.ListDocuments(ListDocumentsRequest{
		PageNumber:           1,
		PageSize:             100,
		DocumentTypeCode:     "FACTURA-UBL",
		ReceiverStartingDate: "2026-09-01T00:00:00",
		ReceiverEndingDate:   "2026-11-30T23:59:00",
	})
	if err != nil {
		t.Fatalf("ListDocuments falló: %v", err)
	}
	if len(listResp.Items) != 1 {
		t.Fatalf("Se esperaba 1 item, obtenidos %d", len(listResp.Items))
	}
	if listResp.Items[0].LDF != expectedLDF {
		t.Errorf("LDF esperado '%s', obtenido '%s'", expectedLDF, listResp.Items[0].LDF)
	}

	// 3. Probar DownloadZip
	zipBytes, err := client.DownloadZip([]string{expectedLDF})
	if err != nil {
		t.Fatalf("DownloadZip falló: %v", err)
	}
	if len(zipBytes) == 0 {
		t.Fatal("Zip recibido está vacío")
	}
}
