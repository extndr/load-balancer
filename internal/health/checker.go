package health

import (
	"net/http"
	"time"

	"github.com/extndr/load-balancer/internal/backend"
	log "github.com/sirupsen/logrus"
)

type Checker struct {
	client  *http.Client
	timeout time.Duration
}

func NewChecker(timeout time.Duration) *Checker {
	return &Checker{
		client:  &http.Client{Timeout: timeout},
		timeout: timeout,
	}
}

func (c *Checker) Check(b *backend.Backend) bool {
	url := b.URL.String()
	resp, err := c.client.Get(url)
	if err != nil {
		log.Debugf("health check failed for %s: %v", b.URL.Host, err)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 500 {
		log.Warnf("health check returned %d for %s", resp.StatusCode, b.URL.Host)
		return false
	}

	log.Debugf("health check passed for %s: %d", b.URL.Host, resp.StatusCode)
	return true
}
