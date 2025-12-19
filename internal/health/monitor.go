package health

import (
	"net/http"
	"time"

	"github.com/extndr/load-balancer/internal/backend"
	"go.uber.org/zap"
)

type Monitor struct {
	client   *http.Client
	pool     *backend.Pool
	interval time.Duration
	stopCh   chan struct{}
	logger   *zap.Logger
}

func NewMonitor(pool *backend.Pool, timeout, interval time.Duration, logger *zap.Logger) *Monitor {
	return &Monitor{
		client:   &http.Client{Timeout: timeout},
		pool:     pool,
		interval: interval,
		stopCh:   make(chan struct{}),
		logger:   logger,
	}
}

// Start begins periodic health checks
func (m *Monitor) Start() {
	ticker := time.NewTicker(m.interval)
	m.logger.Info("health monitor started", zap.Duration("interval", m.interval))

	for {
		select {
		case <-ticker.C:
			m.check()
		case <-m.stopCh:
			ticker.Stop()
			m.logger.Info("health monitor stopped")
			return
		}
	}
}

// Stop terminates the health monitor
func (m *Monitor) Stop() {
	close(m.stopCh)
}

// check performs a health check on all backends
// and updates their status in the pool.
func (m *Monitor) check() {
	for _, b := range m.pool.GetAll() {
		resp, err := m.client.Get(b.URL.String())
		if err != nil {
			m.updateStatus(b, false)
			continue
		}

		// true if backend responded with 2xx or 3xx, false otherwise
		healthy := resp.StatusCode >= 200 && resp.StatusCode < 400
		m.updateStatus(b, healthy)

		resp.Body.Close()
	}
}

// updateStatus updates a backend in the pool if its health has changed
// and logs the update.
func (m *Monitor) updateStatus(b *backend.Backend, newStatus bool) {
	if b.Healthy() != newStatus {
		m.pool.UpdateHealth(b, newStatus)

		status := "healthy"
		level := m.logger.Info
		if !newStatus {
			status = "unhealthy"
			level = m.logger.Warn
		}

		level("backend health status changed",
			zap.String("backend", b.URL.Host),
			zap.String("url", b.URL.String()),
			zap.String("status", status),
		)
	}
}
