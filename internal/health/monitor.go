package health

import (
	"net/http"
	"time"

	"github.com/extndr/load-balancer/internal/backend"
	log "github.com/sirupsen/logrus"
)

type Monitor struct {
	client   *http.Client
	pool     *backend.Pool
	interval time.Duration
	stopCh   chan struct{}
}

func NewMonitor(pool *backend.Pool, timeout, interval time.Duration) *Monitor {
	return &Monitor{
		client:   &http.Client{Timeout: timeout},
		pool:     pool,
		interval: interval,
		stopCh:   make(chan struct{}),
	}
}

// Start begins periodic health checks
func (m *Monitor) Start() {
	ticker := time.NewTicker(m.interval)
	log.Infof("health monitor started, checking every %v", m.interval)

	for {
		select {
		case <-ticker.C:
			m.checkBackends()
		case <-m.stopCh:
			ticker.Stop()
			log.Info("health monitor stopped")
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
		newStatus := resp.StatusCode >= 200 && resp.StatusCode < 400
		m.updateStatus(b, newStatus)

		resp.Body.Close()
	}
}

// updateStatus updates backend status and logs changes
func (m *Monitor) updateStatus(b *backend.Backend, newStatus bool) {
	if b.Healthy() != newStatus {
		b.SetHealthy(newStatus)
		if !newStatus {
			log.Warnf("backend %s is now DOWN", b.URL.Host)
			return
		}
		log.Infof("backend %s is now UP", b.URL.Host)
	}
}
