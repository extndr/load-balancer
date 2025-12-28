package config

import (
	"os"
	"time"
)

type Config struct {
	Port                string
	BackendURLs         []string
	ProxyTimeout        time.Duration
	EnableHealthCheck   bool
	HealthCheckTimeout  time.Duration
	HealthCheckInterval time.Duration

	HTTPTransport HTTPTransportConfig
}

type HTTPTransportConfig struct {
	MaxIdleConns        int
	MaxIdleConnsPerHost int
	IdleConnTimeout     time.Duration
}

func Load() *Config {
	cfg := &Config{
		Port:                getEnv("PORT", defaultPort),
		BackendURLs:         parseBackends(getEnv("BACKENDS", defaultBackends)),
		ProxyTimeout:        parseDuration(getEnv("PROXY_TIMEOUT", defaultProxyTimeout)),
		EnableHealthCheck:   parseBool(getEnv("ENABLE_HEALTH_CHECK", defaultEnableHealthCheck)),
		HealthCheckTimeout:  parseDuration(getEnv("HEALTH_CHECK_TIMEOUT", defaultHealthCheckTimeout)),
		HealthCheckInterval: parseDuration(getEnv("HEALTH_CHECK_INTERVAL", defaultHealthCheckInterval)),
	}

	cfg.HTTPTransport = HTTPTransportConfig{
		MaxIdleConns:        parseInt(getEnv("HTTP_MAX_IDLE_CONNS", defaultMaxIdleConns)),
		MaxIdleConnsPerHost: parseInt(getEnv("HTTP_MAX_IDLE_CONNS_PER_HOST", defaultMaxIdleConnsPerHost)),
		IdleConnTimeout:     parseDuration(getEnv("HTTP_IDLE_CONN_TIMEOUT", defaultIdleConnTimeout)),
	}
	return cfg
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
