package health

import (
	"github.com/extndr/load-balancer/internal/backend"
	log "github.com/sirupsen/logrus"
)

type Monitor struct {
	pool *backend.Pool
}

func NewMonitor(pool *backend.Pool) *Monitor {
	return &Monitor{
		pool: pool,
	}
}

func (m *Monitor) UpdateStatus(b *backend.Backend, alive bool) {
	wasAlive := b.IsAlive()

	if alive != wasAlive {
		b.SetAlive(alive)
		status := "UP"
		if !alive {
			status = "DOWN"
		}
		log.Warnf("backend status changed: %s -> %s", b.URL.Host, status)
	}
}

func (m *Monitor) GetStats() (alive, total int) {
	allBackends := m.pool.AllBackends()
	aliveCount := len(m.pool.AliveBackends())

	return aliveCount, len(allBackends)
}
