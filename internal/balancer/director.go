package balancer

import "github.com/extndr/load-balancer/internal/pool"

type Director struct {
	pool     *pool.Pool
	strategy Strategy
}

func NewDirector(pool *pool.Pool, strategy Strategy) *Director {
	return &Director{
		pool:     pool,
		strategy: strategy,
	}
}

func (d *Director) SelectBackend() *pool.Backend {
	healthy := d.pool.GetHealthy()
	if len(healthy) == 0 {
		return nil
	}
	return d.strategy.Next(healthy)
}
