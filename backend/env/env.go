package env

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Env struct {
	// Database
	DBHost      string
	DBPort      string
	DBUser      string
	DBPassword  string
	DBName      string
	DBCharset   string
	DBParseTime string
	DBLoc       string

	// PT API (Proveedor Tecnológico Inbound & Eventos)
	PtAPIBaseURL             string
	PtAPIUser                string
	PtAPIPassword            string
	PtAPIXWhoListarDescargar string
	PtAPIXWhoAuthEventos     string
	PtAPITenantID            string
}

func getEnv(keys ...string) string {
	for _, key := range keys {
		if val := os.Getenv(key); val != "" {
			return val
		}
	}
	return ""
}

func Load(paths ...string) (*Env, error) {
	candidates := []string{".env", "../.env", "../../.env"}
	if len(paths) > 0 && paths[0] != "" {
		candidates = append([]string{paths[0]}, candidates...)
	}

	var errs []string
	loaded := false
	for _, p := range candidates {
		if err := godotenv.Load(p); err == nil {
			loaded = true
			break
		} else {
			errs = append(errs, fmt.Sprintf("[%s: %v]", p, err))
		}
	}
	if !loaded && os.Getenv("DB_SIGEFAE_HOST") == "" {
		return nil, fmt.Errorf("fallaron los intentos de cargar .env: %s", strings.Join(errs, ", "))
	}

	cfg := &Env{
		// DB
		DBHost:      os.Getenv("DB_SIGEFAE_HOST"),
		DBPort:      os.Getenv("DB_SIGEFAE_PORT"),
		DBUser:      os.Getenv("DB_SIGEFAE_USER"),
		DBPassword:  os.Getenv("DB_SIGEFAE_PASSWORD"),
		DBName:      os.Getenv("DB_SIGEFAE_NAME"),
		DBCharset:   os.Getenv("DB_SIGEFAE_CHARSET"),
		DBParseTime: os.Getenv("DB_SIGEFAE_PARSE_TIME"),
		DBLoc:       os.Getenv("DB_SIGEFAE_LOC"),

		// PT API
		PtAPIBaseURL:             getEnv("PT_API_BASE_URL"),
		PtAPIUser:                getEnv("PT_API_USER"),
		PtAPIPassword:            getEnv("PT_API_PASSWORD"),
		PtAPIXWhoListarDescargar: getEnv("PT_API_X_WHO_LISTAR-DESCARGAR", "PT_API_X_WHO_LISTAR_DESCARGAR"),
		PtAPIXWhoAuthEventos:     getEnv("PT_API_X_WHO_AUTENTICAR-EVENTOS", "PT_API_X_WHO_AUTENTICAR_EVENTOS"),
		PtAPITenantID:            getEnv("PT_API_TENANT_ID"),
	}

	if cfg.DBHost == "" ||
		cfg.DBUser == "" ||
		cfg.DBPassword == "" ||
		cfg.DBName == "" {
		return nil, fmt.Errorf("missing database environment variables")
	}

	return cfg, nil
}
