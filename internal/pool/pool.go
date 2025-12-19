package pool

import (
	"errors"
	"fmt"
	"net/url"
	"sync/atomic"
)

type Pool struct {
	backends []*Backend
	cache    atomic.Value
}

func NewPool(urls []string) (*Pool, error) {
	if len(urls) < 1 {
		return nil, errors.New("at least one backend required")
	}

	var backends []*Backend
	for _, s := range urls {
		u, err := url.Parse(s)
		if err != nil {
			return nil, fmt.Errorf("invalid backend URL %q: %w", s, err)
		}
		backends = append(backends, NewBackend(u))
	}

	p := &Pool{backends: backends}
	p.updateCache()
	return p, nil
}

func (p *Pool) GetHealthy() []*Backend {
	return p.cache.Load().([]*Backend)
}

func (p *Pool) GetAll() []*Backend {
	return p.backends
}

// UpdateHealth sets the backend's health and updates the cached healthy list.
func (p *Pool) UpdateHealth(b *Backend, healthy bool) {
	b.setHealthy(healthy)
	p.updateCache()
}

// updateCache rebuilds the list of healthy backends in the cache.
func (p *Pool) updateCache() {
	healthy := make([]*Backend, 0, len(p.backends))
	for _, b := range p.backends {
		if b.Healthy() {
			healthy = append(healthy, b)
		}
	}
	p.cache.Store(healthy)
}
