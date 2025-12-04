package backend

import (
	"net/url"
	"sync/atomic"
)

type Backend struct {
	URL   *url.URL
	Alive atomic.Bool
}

func NewBackend(u *url.URL) *Backend {
	b := &Backend{URL: u}
	b.Alive.Store(true)
	return b
}

func (b *Backend) SetAlive(alive bool) {
	b.Alive.Store(alive)
}

func (b *Backend) IsAlive() bool {
	return b.Alive.Load()
}
