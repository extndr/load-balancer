package backend

import (
	"net/url"
	"sync/atomic"
)

type Backend struct {
	URL     *url.URL
	healthy atomic.Bool
}

func NewBackend(u *url.URL) *Backend {
	b := &Backend{URL: u}
	b.healthy.Store(true)
	return b
}

func (b *Backend) Healthy() bool {
	return b.healthy.Load()
}

func (b *Backend) SetHealthy(status bool) {
	b.healthy.Store(status)
}
