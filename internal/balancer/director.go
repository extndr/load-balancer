package balancer

import (
	"net/http"

	"github.com/extndr/load-balancer/internal/backend"
	"github.com/extndr/load-balancer/internal/proxy"
	log "github.com/sirupsen/logrus"
)

type Director struct {
	Strategy Strategy
	Proxy    *proxy.Proxy
}

func NewDirector(s Strategy, p *proxy.Proxy) *Director {
	return &Director{
		Strategy: s,
		Proxy:    p,
	}
}

func (d *Director) SelectBackend() *backend.Backend {
	target := d.Strategy.Next()
	return target
}

func (d *Director) ProxyRequest(target *backend.Backend, r *http.Request) (*http.Response, error) {
	resp, err := d.Proxy.DoRequest(target, r)
	if err != nil {
		log.Errorf("proxy request failed for %s: %v", target.URL.Host, err)
		return nil, err
	}
	return resp, nil
}
