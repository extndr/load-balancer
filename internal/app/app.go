package app

import (
	"context"
	"fmt"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/extndr/load-balancer/internal/balancer"
	"github.com/extndr/load-balancer/internal/config"
	"github.com/extndr/load-balancer/internal/health"
	"github.com/extndr/load-balancer/internal/middleware"
	"github.com/extndr/load-balancer/internal/pool"
	"github.com/extndr/load-balancer/internal/proxy"
	"github.com/extndr/load-balancer/internal/server"
	"go.uber.org/zap"
)

type App struct {
	server  *http.Server
	monitor *health.Monitor
	logger  *zap.Logger
}

func New(cfg *config.Config, logger *zap.Logger) (*App, error) {
	pool, err := pool.NewPool(cfg.BackendURLs)
	if err != nil {
		return nil, fmt.Errorf("failed to create backend pool: %w", err)
	}

	rproxy := proxy.NewProxy(cfg.ProxyTimeout, cfg.HTTPTransport)
	strategy := balancer.NewRoundRobin()
	director := balancer.NewDirector(pool, strategy)
	monitor := health.NewMonitor(
		pool,
		cfg.HealthCheckTimeout,
		cfg.HealthCheckInterval,
		logger,
	)

	handler := server.NewHandler(director, rproxy)
	chain := middleware.Chain(handler, middleware.Logging(logger))

	srv := server.NewServer(":"+cfg.Port, chain)

	return &App{
		server:  srv,
		monitor: monitor,
		logger:  logger,
	}, nil
}

func (a *App) Run() error {
	serverErr := make(chan error, 1)

	a.logger.Info("starting load balancer", zap.String("addr", a.server.Addr))
	go func() {
		err := a.server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			serverErr <- err
		}
	}()

	go a.monitor.Start()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	select {
	case <-ctx.Done():
		a.logger.Info("received shutdown signal")
	case err := <-serverErr:
		a.logger.Error("server error", zap.Error(err))
		a.monitor.Stop()
		return err
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return a.Stop(shutdownCtx)
}

func (a *App) Stop(ctx context.Context) error {
	a.logger.Info("stopping health monitor")
	a.monitor.Stop()

	a.logger.Info("shutting down server")
	if err := a.server.Shutdown(ctx); err != nil {
		a.logger.Error("server shutdown error", zap.Error(err))
		return err
	}

	a.logger.Info("server gracefully shutdown")
	return nil
}
