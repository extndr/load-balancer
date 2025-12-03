package core

import (
	"net/http"

	"github.com/extndr/load-balancer/internal/pool"
	"github.com/extndr/load-balancer/internal/proxy"
	log "github.com/sirupsen/logrus"
)

type Strategy interface {
	Next() *pool.Backend
}

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

func (d *Director) SelectBackend() *pool.Backend {
	target := d.Strategy.Next()
	return target
}

func (d *Director) ProxyRequest(target *pool.Backend, r *http.Request) (*http.Response, error) {
	resp, err := d.Proxy.DoRequest(target, r)
	if err != nil {
		log.Errorf("proxy request failed for %s: %v", target.URL.Host, err)
		return nil, err
	}
	return resp, nil
}
