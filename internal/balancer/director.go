package balancer

import (
	"github.com/extndr/load-balancer/internal/backend"
)

type Director struct {
	pool     *backend.Pool
	strategy Strategy
}

func NewDirector(pool *backend.Pool, strategy Strategy) *Director {
	return &Director{
		pool:     pool,
		strategy: strategy,
	}
}

func (d *Director) SelectBackend() *backend.Backend {
	healthy := d.pool.GetHealthy()
	if len(healthy) == 0 {
		return nil
	}
	return d.strategy.Next(healthy)
}
