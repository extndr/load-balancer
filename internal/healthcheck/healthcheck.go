package healthcheck

import (
	"context"
	"net/http"
	"time"

	"github.com/extndr/load-balancer/internal/types"
	"go.uber.org/zap"
)

type HealthChecker struct {
	targets  []string
	client   *http.Client
	interval time.Duration
	logger   *zap.Logger
}

func NewHealthChecker(targets []string, timeout, interval time.Duration, logger *zap.Logger) *HealthChecker {
	return &HealthChecker{
		targets:  targets,
		client:   &http.Client{Timeout: timeout},
		interval: interval,
		logger:   logger,
	}
}

func (c *HealthChecker) Start(ctx context.Context) <-chan types.StatusEvent {
	events := make(chan types.StatusEvent, len(c.targets))

	c.logger.Info("starting health checker",
		zap.Strings("targets", c.targets),
		zap.Duration("interval", c.interval),
		zap.Duration("timeout", c.client.Timeout),
	)

	go func() {
		defer close(events)
		ticker := time.NewTicker(c.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				c.logger.Info("health checker stopped due to context cancellation")
				return
			case <-ticker.C:
				c.logger.Debug("starting health check round", zap.Strings("targets", c.targets))

				// Fan-out
				for _, url := range c.targets {
					go func(u string) {
						start := time.Now()
						healthy, err := c.check(ctx, u)

						select {
						case events <- types.StatusEvent{
							URL:     u,
							Healthy: healthy,
							Latency: time.Since(start),
							Error:   err,
						}:
						case <-ctx.Done():
						}
					}(url)
				}
			}
		}
	}()

	return events
}

func (c *HealthChecker) check(ctx context.Context, url string) (bool, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		c.logger.Debug("failed to create health check request",
			zap.String("url", url),
			zap.Error(err),
		)
		return false, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		c.logger.Debug("health check request failed",
			zap.String("url", url),
			zap.Error(err),
		)
		return false, err
	}
	defer resp.Body.Close()

	healthy := resp.StatusCode >= 200 && resp.StatusCode < 300

	c.logger.Debug("health check completed",
		zap.String("url", url),
		zap.Int("status_code", resp.StatusCode),
		zap.Bool("healthy", healthy),
	)

	return healthy, nil
}
