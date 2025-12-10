package balancer

import (
	"github.com/extndr/load-balancer/internal/backend"
)

type Director struct {
	strategy Strategy
}

func NewDirector(strategy Strategy) *Director {
	return &Director{
		strategy: strategy,
	}
}

func (d *Director) SelectBackend() *backend.Backend {
	return d.strategy.Next()
}
