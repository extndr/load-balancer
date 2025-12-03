package core

import (
	"sync/atomic"

	"github.com/extndr/load-balancer/internal/pool"
)

type RoundRobin struct {
	counter  uint64
	backends *pool.Pool
}

func NewRoundRobin(backends *pool.Pool) *RoundRobin {
	return &RoundRobin{
		backends: backends,
	}
}

func (rr *RoundRobin) Next() *pool.Backend {
	alive := rr.backends.AliveBackends()
	idx := atomic.AddUint64(&rr.counter, 1) - 1
	selected := alive[idx%uint64(len(alive))]
	return selected
}
