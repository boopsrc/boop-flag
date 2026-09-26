package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config carrega toda a configuração externa do serviço a partir de variáveis
// de ambiente. É validada uma única vez no boot (Load), falhando rápido se algo
// obrigatório estiver ausente, para o processo não subir num estado inválido.
type Config struct {
	Port        string
	DatabaseURL string
	FrontendURL string

	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURL  string

	JWTSecret  string
	SessionTTL time.Duration

	// AuthDevMode habilita um login mock que dispensa credenciais reais do
	// Google. Serve para desenvolvimento e demonstração; nunca deve ficar
	// ligado em produção.
	AuthDevMode bool
}

// Load lê a configuração do ambiente e a valida. Retorna erro (em vez de
// aplicar defaults silenciosos) para segredos obrigatórios ausentes.
func Load() (*Config, error) {
	cfg := &Config{
		Port:               getEnv("PORT", "8080"),
		DatabaseURL:        os.Getenv("DATABASE_URL"),
		FrontendURL:        getEnv("FRONTEND_URL", "http://localhost:5173"),
		GoogleClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		GoogleClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		GoogleRedirectURL:  getEnv("GOOGLE_REDIRECT_URL", "http://localhost:8080/api/v1/auth/google/callback"),
		JWTSecret:          os.Getenv("JWT_SECRET"),
		AuthDevMode:        getEnvBool("AUTH_DEV_MODE", false),
	}

	ttl, err := time.ParseDuration(getEnv("SESSION_TTL", "24h"))
	if err != nil {
		return nil, fmt.Errorf("invalid SESSION_TTL: %w", err)
	}
	cfg.SessionTTL = ttl

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}
	// Em modo dev o login mock dispensa o Google; fora dele, as credenciais
	// OAuth são obrigatórias para o fluxo funcionar.
	if !cfg.AuthDevMode {
		if cfg.GoogleClientID == "" || cfg.GoogleClientSecret == "" {
			return nil, fmt.Errorf("GOOGLE_CLIENT_ID and GOOGLE_CLIENT_SECRET are required when AUTH_DEV_MODE is off")
		}
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return fallback
	}
	return b
}
