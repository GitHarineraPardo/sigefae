package proveedor_inbound

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

type Client struct {
	baseURL             string
	user                string
	password            string
	xWhoListarDescargar string
	xWhoAuthEventos     string
	httpClient          *http.Client

	mu    sync.RWMutex
	token string
}

func NewClient(baseURL, user, password, xWhoListarDescargar, xWhoAuthEventos string) *Client {
	baseURL = strings.TrimRight(baseURL, "/")
	return &Client{
		baseURL:             baseURL,
		user:                user,
		password:            password,
		xWhoListarDescargar: xWhoListarDescargar,
		xWhoAuthEventos:     xWhoAuthEventos,
		httpClient:          &http.Client{Timeout: 60 * time.Second},
	}
}

// Login realiza la autenticación ante /Auth/Login y almacena el accessToken
func (c *Client) Login() (*LoginResponse, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	endpoint := fmt.Sprintf("%s/Auth/Login", c.baseURL)
	body, err := json.Marshal(LoginRequest{
		U: c.user,
		P: c.password,
	})
	if err != nil {
		return nil, fmt.Errorf("error serializando login request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("error creando request de login: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.xWhoAuthEventos != "" {
		req.Header.Set("x-who", c.xWhoAuthEventos)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error ejecutando login en %s: %w", endpoint, err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error leyendo respuesta de login: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("login falló con HTTP %d: %s", resp.StatusCode, string(respBytes))
	}

	var loginResp LoginResponse
	if err := json.Unmarshal(respBytes, &loginResp); err != nil {
		return nil, fmt.Errorf("error deserializando respuesta de login: %w", err)
	}

	if loginResp.AccessToken == "" {
		return nil, errors.New("la respuesta de login no contiene un accessToken válido")
	}

	c.token = loginResp.AccessToken
	return &loginResp, nil
}

// GetToken retorna el token actual o se autentica si aún no existe
func (c *Client) GetToken() (string, error) {
	c.mu.RLock()
	tok := c.token
	c.mu.RUnlock()

	if tok != "" {
		return tok, nil
	}

	loginResp, err := c.Login()
	if err != nil {
		return "", err
	}
	return loginResp.AccessToken, nil
}

// ResetToken invalida el token actual en memoria para forzar re-login
func (c *Client) ResetToken() {
	c.mu.Lock()
	c.token = ""
	c.mu.Unlock()
}

// ListDocuments consulta /InboundDocument/Integrated/Document con paginación y filtros
func (c *Client) ListDocuments(req ListDocumentsRequest) (*ListDocumentsResponse, error) {
	endpoint := fmt.Sprintf("%s/InboundDocument/Integrated/Document", c.baseURL)

	doRequest := func(token string) (*http.Response, error) {
		body, err := json.Marshal(req)
		if err != nil {
			return nil, err
		}

		httpReq, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(body))
		if err != nil {
			return nil, err
		}

		httpReq.Header.Set("Content-Type", "application/json")
		if c.xWhoListarDescargar != "" {
			httpReq.Header.Set("x-who", c.xWhoListarDescargar)
		}
		httpReq.Header.Set("Authorization", "Bearer "+token)

		return c.httpClient.Do(httpReq)
	}

	token, err := c.GetToken()
	if err != nil {
		return nil, err
	}

	resp, err := doRequest(token)
	if err != nil {
		return nil, fmt.Errorf("error llamando a listar documentos: %w", err)
	}

	// Si responde 401 Unauthorized, reintentar una vez con re-login
	if resp.StatusCode == http.StatusUnauthorized {
		resp.Body.Close()
		c.ResetToken()
		newToken, err := c.GetToken()
		if err != nil {
			return nil, fmt.Errorf("error renovando token tras 401: %w", err)
		}
		resp, err = doRequest(newToken)
		if err != nil {
			return nil, fmt.Errorf("error llamando a listar documentos en reintento: %w", err)
		}
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error leyendo respuesta de listar documentos: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("listar documentos falló con HTTP %d: %s", resp.StatusCode, string(respBytes))
	}

	var listResp ListDocumentsResponse
	if err := json.Unmarshal(respBytes, &listResp); err != nil {
		return nil, fmt.Errorf("error deserializando lista de documentos: %w (respuesta: %s)", err, string(respBytes))
	}

	return &listResp, nil
}

// DownloadZip consulta /Inbound/Download/Zip con una lista de LDFs y retorna los bytes binarios del archivo .zip
func (c *Client) DownloadZip(ldfList []string) ([]byte, error) {
	if len(ldfList) == 0 {
		return nil, errors.New("la lista de LDF para descargar ZIP no puede estar vacía")
	}

	endpoint := fmt.Sprintf("%s/Inbound/Download/Zip", c.baseURL)

	doRequest := func(token string) (*http.Response, error) {
		body, err := json.Marshal(ldfList)
		if err != nil {
			return nil, err
		}

		httpReq, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(body))
		if err != nil {
			return nil, err
		}

		httpReq.Header.Set("Content-Type", "application/json")
		if c.xWhoListarDescargar != "" {
			httpReq.Header.Set("x-who", c.xWhoListarDescargar)
		}
		httpReq.Header.Set("Authorization", "Bearer "+token)

		return c.httpClient.Do(httpReq)
	}

	token, err := c.GetToken()
	if err != nil {
		return nil, err
	}

	resp, err := doRequest(token)
	if err != nil {
		return nil, fmt.Errorf("error llamando a descarga de zip: %w", err)
	}

	// Reintento en 401
	if resp.StatusCode == http.StatusUnauthorized {
		resp.Body.Close()
		c.ResetToken()
		newToken, err := c.GetToken()
		if err != nil {
			return nil, fmt.Errorf("error renovando token tras 401 en descarga zip: %w", err)
		}
		resp, err = doRequest(newToken)
		if err != nil {
			return nil, fmt.Errorf("error llamando a descarga de zip en reintento: %w", err)
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		errBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("descarga de zip falló con HTTP %d: %s", resp.StatusCode, string(errBytes))
	}

	zipBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error leyendo contenido del zip: %w", err)
	}

	if len(zipBytes) < 4 || (zipBytes[0] != 'P' && zipBytes[1] != 'K') {
		return nil, fmt.Errorf("la respuesta recibida no parece ser un archivo ZIP válido (tamaño: %d bytes)", len(zipBytes))
	}

	return zipBytes, nil
}

// postEvent ejecuta una petición POST hacia un endpoint de eventos con el Bearer token y cabecera X-Who
func (c *Client) postEvent(endpoint string, payload any) (string, error) {
	doRequest := func(token string) (*http.Response, error) {
		bodyBytes, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("error serializando body de evento: %w", err)
		}

		httpReq, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(bodyBytes))
		if err != nil {
			return nil, err
		}

		httpReq.Header.Set("Content-Type", "application/json")
		if c.xWhoAuthEventos != "" {
			httpReq.Header.Set("x-who", c.xWhoAuthEventos)
		}
		httpReq.Header.Set("Authorization", "Bearer "+token)

		return c.httpClient.Do(httpReq)
	}

	token, err := c.GetToken()
	if err != nil {
		return "", err
	}

	resp, err := doRequest(token)
	if err != nil {
		return "", fmt.Errorf("error ejecutando evento en %s: %w", endpoint, err)
	}

	// 401 reintento
	if resp.StatusCode == http.StatusUnauthorized {
		resp.Body.Close()
		c.ResetToken()
		newToken, err := c.GetToken()
		if err != nil {
			return "", fmt.Errorf("error renovando token tras 401: %w", err)
		}
		resp, err = doRequest(newToken)
		if err != nil {
			return "", fmt.Errorf("error ejecutando evento en reintento: %w", err)
		}
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error leyendo respuesta de evento: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return string(respBytes), fmt.Errorf("evento falló con HTTP %d: %s", resp.StatusCode, string(respBytes))
	}

	return string(respBytes), nil
}

// EmitirAcuse envía el evento 030 (Acuse de Recibo) ante /Issue/Acknowledgement/V2/{ldf}
func (c *Client) EmitirAcuse(ldf string, req AcuseEventRequest) (string, error) {
	endpoint := fmt.Sprintf("%s/Issue/Acknowledgement/V2/%s", c.baseURL, ldf)
	return c.postEvent(endpoint, req)
}

// EmitirReciboBienes envía el evento 032 (Recibo del Bien o Prestación del Servicio) ante /Issue/ReceiveGoods/{ldf}
func (c *Client) EmitirReciboBienes(ldf string, req ReciboGoodsEventRequest) (string, error) {
	endpoint := fmt.Sprintf("%s/Issue/ReceiveGoods/%s", c.baseURL, ldf)
	return c.postEvent(endpoint, req)
}

// EmitirAceptacion envía el evento 034 (Aceptación Expresa) ante /Issue/Accept/V2/{ldf}
func (c *Client) EmitirAceptacion(ldf string, req AceptacionEventRequest) (string, error) {
	endpoint := fmt.Sprintf("%s/Issue/Accept/V2/%s", c.baseURL, ldf)
	return c.postEvent(endpoint, req)
}

// EmitirReclamo envía el evento 031 (Reclamo / Rechazo) ante /Issue/Reject/V2/{ldf}
func (c *Client) EmitirReclamo(ldf string, req ReclamoEventRequest) (string, error) {
	endpoint := fmt.Sprintf("%s/Issue/Reject/V2/%s", c.baseURL, ldf)
	return c.postEvent(endpoint, req)
}
