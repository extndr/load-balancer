package balancer

import (
	"context"
	"net/url"
	"sync"
	"sync/atomic"

	"github.com/extndr/load-balancer/internal/types"
	"go.uber.org/zap"
)

type Backend struct {
	URL *url.URL
	up  bool
}

type BackendPool struct {
	backends []*Backend
	index    map[string]*Backend
	healthy  atomic.Value
}

func (p *BackendPool) Healthy() []*Backend {
	return p.healthy.Load().([]*Backend)
}

func (p *BackendPool) rebuildCache() {
	hs := make([]*Backend, 0, len(p.backends))
	for _, be := range p.backends {
		if be.up {
			hs = append(hs, be)
		}
	}
	p.healthy.Store(hs) // publish snapshot
}

type Balancer struct {
	mu      sync.Mutex // for backend status changing
	pool    *BackendPool
	counter uint64
	logger  *zap.Logger
}

func New(backends []*Backend, logger *zap.Logger) *Balancer {
	pool := &BackendPool{
		backends: backends,
		index:    make(map[string]*Backend, len(backends)),
	}
	for _, b := range backends {
		pool.index[b.URL.Host] = b
	}
	pool.rebuildCache()

	bal := &Balancer{
		pool:   pool,
		logger: logger,
	}

	// Log initial backend status
	healthyCount := len(pool.Healthy())
	backendURLs := make([]string, len(backends))
	for i, b := range backends {
		backendURLs[i] = b.URL.String()
	}
	bal.logger.Info("load balancer initialized",
		zap.Int("total_backends", len(backends)),
		zap.Int("healthy_backends", healthyCount),
		zap.Strings("backend_urls", backendURLs),
	)

	return bal
}

func (b *Balancer) SetStatus(host string, healthy bool) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if be, ok := b.pool.index[host]; ok && be.up != healthy {
		be.up = healthy
		previousHealthy := len(b.pool.healthy.Load().([]*Backend))
		b.pool.rebuildCache()
		newHealthy := len(b.pool.healthy.Load().([]*Backend))

		status := "unhealthy"
		if healthy {
			status = "healthy"
		}

		b.logger.Info("backend status changed",
			zap.String("backend", be.URL.String()),
			zap.String("status", status),
			zap.Int("previous_healthy_count", previousHealthy),
			zap.Int("new_healthy_count", newHealthy),
		)
	}
}

func (b *Balancer) HandleHealthEvents(ctx context.Context, events <-chan types.StatusEvent) {
	for {
		select {
		case ev, ok := <-events:
			if !ok {
				b.logger.Info("health check event channel closed")
				return
			}

			if ev.Error != nil {
				b.logger.Warn("health check failed",
					zap.String("backend", ev.URL),
					zap.Duration("latency", ev.Latency),
					zap.Error(ev.Error),
				)
			} else {
				b.logger.Debug("health check completed",
					zap.String("backend", ev.URL),
					zap.Duration("latency", ev.Latency),
					zap.Bool("healthy", ev.Healthy),
				)
			}

			b.SetStatus(ev.URL, ev.Healthy)

		case <-ctx.Done():
			b.logger.Info("health event handler stopped due to context cancellation")
			return
		}
	}
}

func (b *Balancer) NextBackend() *Backend {
	cache := b.pool.healthy.Load().([]*Backend)
	if len(cache) == 0 {
		b.logger.Error("no healthy backends available for request routing")
		return nil
	}

	n := atomic.AddUint64(&b.counter, 1) - 1
	backend := cache[n%uint64(len(cache))]

	b.logger.Debug("backend selected for request",
		zap.String("backend", backend.URL.String()),
		zap.Uint64("request_count", b.counter+1),
		zap.Int("available_healthy_backends", len(cache)),
	)

	return backend
}
