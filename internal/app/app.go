package app

import (
	"context"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/extndr/load-balancer/internal/backend"
	"github.com/extndr/load-balancer/internal/balancer"
	"github.com/extndr/load-balancer/internal/config"
	"github.com/extndr/load-balancer/internal/health"
	"github.com/extndr/load-balancer/internal/middleware"
	"github.com/extndr/load-balancer/internal/proxy"
	"github.com/extndr/load-balancer/internal/server"
	log "github.com/sirupsen/logrus"
)

type App struct {
	server  *http.Server
	monitor *health.Monitor
}

func New(cfg *config.Config) (*App, error) {
	p, err := backend.NewPool(cfg.BackendURLs)
	if err != nil {
		return nil, err
	}

	proxyClient := proxy.NewProxy(cfg.ProxyTimeout, cfg.HTTPTransport)
	strategy := balancer.NewRoundRobin(p)
	director := balancer.NewDirector(strategy, proxyClient)
	monitor := health.NewMonitor(p, cfg.HealthCheckTimeout, cfg.HealthCheckInterval)
	handler := server.NewHandler(director)
	chain := middleware.Chain(handler, middleware.Logging())
	srv := server.NewServer(":"+cfg.Port, chain)

	return &App{
		server:  srv,
		monitor: monitor,
	}, nil
}

func (a *App) Run() error {
	serverErr := make(chan error, 1)

	log.Infof("starting load balancer on %s", a.server.Addr)
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
		log.Infof("received shutdown signal")
	case err := <-serverErr:
		log.Errorf("server error: %v", err)
		a.monitor.Stop()
		return err
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return a.Stop(shutdownCtx)
}

func (a *App) Stop(ctx context.Context) error {
	log.Infof("stopping health monitor")
	a.monitor.Stop()

	log.Infof("shutting down server")
	if err := a.server.Shutdown(ctx); err != nil {
		log.Errorf("server shutdown error: %v", err)
		return err
	}

	log.Info("server gracefully shutdown")
	return nil
}
