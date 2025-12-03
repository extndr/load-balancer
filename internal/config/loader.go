package config

import (
	"os"

	log "github.com/sirupsen/logrus"
)

func Load() *Config {
	cfg := &Config{
		Port:                getEnv("PORT", defaultPort),
		BackendURLs:         parseBackends(getEnv("BACKENDS", defaultBackends)),
		ProxyTimeout:        parseDuration(getEnv("PROXY_TIMEOUT", defaultProxyTimeout)),
		HealthTimeout:       parseDuration(getEnv("HEALTH_TIMEOUT", defaultHealthTimeout)),
		HealthCheckInterval: parseDuration(getEnv("HEALTH_CHECK_INTERVAL", defaultHealthCheckInterval)),
	}

	cfg.HTTPTransport = HTTPTransportConfig{
		MaxIdleConns:        parseInt(getEnv("HTTP_MAX_IDLE_CONNS", defaultMaxIdleConns)),
		MaxIdleConnsPerHost: parseInt(getEnv("HTTP_MAX_IDLE_CONNS_PER_HOST", defaultMaxIdleConnsPerHost)),
		IdleConnTimeout:     parseDuration(getEnv("HTTP_IDLE_CONN_TIMEOUT", defaultIdleConnTimeout)),
	}

	log.Debugf("config loaded: %+v", cfg)
	return cfg
}

func getEnv(key, defaultVal string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultVal
}
