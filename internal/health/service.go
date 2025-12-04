package health

import (
	"time"

	"github.com/extndr/load-balancer/internal/backend"
	log "github.com/sirupsen/logrus"
)

type Service struct {
	checker  *Checker
	monitor  *Monitor
	pool     *backend.Pool
	interval time.Duration
	ticker   *time.Ticker
	stopCh   chan struct{}
}

func NewService(checker *Checker, monitor *Monitor, pool *backend.Pool, interval time.Duration) *Service {
	return &Service{
		checker:  checker,
		monitor:  monitor,
		pool:     pool,
		interval: interval,
		stopCh:   make(chan struct{}),
	}
}

func (s *Service) Start() {
	s.ticker = time.NewTicker(s.interval)

	log.Infof("starting health checks every %v", s.interval)

	for {
		select {
		case <-s.ticker.C:
			s.checkAll()
		case <-s.stopCh:
			s.ticker.Stop()
			log.Infof("health checks stopped")
			return
		}
	}
}

func (s *Service) Stop() {
	close(s.stopCh)
}

func (s *Service) checkAll() {
	allBackends := s.pool.AllBackends()
	log.Debugf("running health check for %d backends", len(allBackends))

	for _, b := range allBackends {
		alive := s.checker.Check(b)
		s.monitor.UpdateStatus(b, alive)
	}

	aliveCount, totalCount := s.monitor.GetStats()
	log.Debugf("health check completed: %d/%d backends alive", aliveCount, totalCount)
}
