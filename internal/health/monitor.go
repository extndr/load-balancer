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
			m.checkBackends()
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

// checkBackends checks all backends and updates their status
func (m *Monitor) checkBackends() {
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

// updateStatus updates backend status and logs changes
func (m *Monitor) updateStatus(b *backend.Backend, healthy bool) {
	if b.Healthy() != healthy {
		b.SetHealthy(healthy)

		status := "healthy"
		level := m.logger.Info
		if !healthy {
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
