package app

import (
	"context"
	"net/http"
	"os"
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
	errCh   chan error
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
		errCh:   make(chan error, 1),
	}, nil
}

func (a *App) Start() error {
	go a.monitor.Start()
	log.Infof("starting load balancer on %s", a.server.Addr)
	go a.waitForShutdown()

	if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Errorf("server error: %v", err)
		a.errCh <- err
	}

	return <-a.errCh
}

func (a *App) Stop(ctx context.Context) error {
	a.monitor.Stop()

	if err := a.server.Shutdown(ctx); err != nil {
		log.Errorf("server shutdown error: %v", err)
		return err
	}

	log.Infof("server gracefully shutdown")
	return nil
}

func (a *App) waitForShutdown() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigChan
	log.Infof("received signal: %v, initiating graceful shutdown", sig)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := a.Stop(ctx); err != nil {
		a.errCh <- err
	} else {
		a.errCh <- nil
	}
}
