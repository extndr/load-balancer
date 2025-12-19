package balancer

import (
	"sync/atomic"

	"github.com/extndr/load-balancer/internal/pool"
)

type RoundRobin struct {
	counter uint64
}

func NewRoundRobin() *RoundRobin {
	return &RoundRobin{}
}

func (rr *RoundRobin) Next(backends []*pool.Backend) *pool.Backend {
	n := atomic.AddUint64(&rr.counter, 1) - 1
	return backends[n%uint64(len(backends))]
}
