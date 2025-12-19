package balancer

import "github.com/extndr/load-balancer/internal/pool"

type Strategy interface {
	Next(backends []*pool.Backend) *pool.Backend
}
