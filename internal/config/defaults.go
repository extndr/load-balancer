package config

const (
	defaultPort                = "8080"
	defaultBackends            = "http://localhost:8081,http://localhost:8082,http://localhost:8083"
	defaultProxyTimeout        = "10s"
	defaultHealthTimeout       = "5s"
	defaultHealthCheckInterval = "30s"
)

const (
	defaultMaxIdleConns        = "30"
	defaultMaxIdleConnsPerHost = "30"
	defaultIdleConnTimeout     = "90s"
)
