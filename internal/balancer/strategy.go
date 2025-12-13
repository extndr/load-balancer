package balancer

import "github.com/extndr/load-balancer/internal/backend"

type Strategy interface {
	Next(backends []*backend.Backend) *backend.Backend
}
