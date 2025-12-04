package proxy

import (
	"net"
	"net/http"
	"time"

	"github.com/extndr/load-balancer/internal/config"
)

func NewTransport(cfg config.HTTPTransportConfig) *http.Transport {
	return &http.Transport{
		MaxIdleConns:        cfg.MaxIdleConns,
		MaxIdleConnsPerHost: cfg.MaxIdleConnsPerHost,
		IdleConnTimeout:     cfg.IdleConnTimeout,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
	}
}
