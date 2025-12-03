package pool

import (
	"errors"
	"fmt"
	"net/url"

	log "github.com/sirupsen/logrus"
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
			log.Errorf("failed to parse backend URL %q: %v", s, err)
			return nil, fmt.Errorf("invalid backend URL %q: %w", s, err)
		}
		backends = append(backends, NewBackend(u))
	}

	return &Pool{backends: backends}, nil
}

func (p *Pool) AliveBackends() []*Backend {
	var alive []*Backend
	for _, b := range p.backends {
		if b.IsAlive() {
			alive = append(alive, b)
		}
	}
	return alive
}

func (p *Pool) AllBackends() []*Backend {
	return p.backends
}
