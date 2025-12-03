package config

import "time"

type Config struct {
	Port                string
	BackendURLs         []string
	ProxyTimeout        time.Duration
	HealthTimeout       time.Duration
	HealthCheckInterval time.Duration
	HTTPTransport       HTTPTransportConfig
}

type HTTPTransportConfig struct {
	MaxIdleConns        int
	MaxIdleConnsPerHost int
	IdleConnTimeout     time.Duration
}
