package backend

import (
	"errors"
	"fmt"
	"net/url"
)

type Pool struct {
	backends []*Backend
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

	return &Pool{backends: backends}, nil
}

func (p *Pool) GetHealthy() []*Backend {
	var healthy []*Backend
	for _, b := range p.backends {
		if b.Healthy() {
			healthy = append(healthy, b)
		}
	}
	return healthy
}

func (p *Pool) GetAll() []*Backend {
	return p.backends
}
