package balancer

import (
	"sync/atomic"

	"github.com/extndr/load-balancer/internal/backend"
)

type RoundRobin struct {
	counter  uint64
	backends *backend.Pool
}

func NewRoundRobin(backends *backend.Pool) *RoundRobin {
	return &RoundRobin{
		backends: backends,
	}
}

func (rr *RoundRobin) Next() *backend.Backend {
	healthy := rr.backends.GetHealthy()
	if len(healthy) == 0 {
		return nil
	}
	idx := atomic.AddUint64(&rr.counter, 1) - 1
	selected := healthy[idx%uint64(len(healthy))]
	return selected
}
