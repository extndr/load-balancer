package app

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/extndr/load-balancer/internal/config"
	"github.com/extndr/load-balancer/internal/core"
	"github.com/extndr/load-balancer/internal/health"
	"github.com/extndr/load-balancer/internal/pool"
	"github.com/extndr/load-balancer/internal/proxy"
	log "github.com/sirupsen/logrus"

	httputil "github.com/extndr/load-balancer/internal/http"
)

type App struct {
	server        *http.Server
	healthService *health.Service
	errCh         chan error
}

func New(cfg *config.Config) (*App, error) {
	// Initialize pool
	p, err := pool.NewPool(cfg.BackendURLs)
	if err != nil {
		return nil, err
	}
	log.Infof("initialized pool with %d backends", len(cfg.BackendURLs))

	// Initialize proxy
	proxyClient := proxy.NewProxy(cfg.ProxyTimeout, cfg.HTTPTransport)
	log.Debugf("proxy client initialized with timeout=%v", cfg.ProxyTimeout)

	// Initialize round-robin strategy
	strategy := core.NewRoundRobin(p)
	log.Debugf("round-robin strategy created")

	// Initialize director
	director := core.NewDirector(strategy, proxyClient)
	log.Debugf("director initialized")

	// Initialize handler
	handler := core.NewHandler(director)

	// Initialize server
	addr := ":" + cfg.Port
	srv := httputil.NewServer(addr, handler)

	// Initialize health service
	checker := health.NewChecker(cfg.HealthTimeout)
	monitor := health.NewMonitor(p)
	healthService := health.NewService(checker, monitor, p, cfg.HealthCheckInterval)
	log.Infof("health service initialized with timeout=%v, interval=%v", cfg.HealthTimeout, cfg.HealthCheckInterval)

	return &App{
		server:        srv,
		healthService: healthService,
		errCh:         make(chan error, 1),
	}, nil
}

func (a *App) Start() error {
	go a.healthService.Start()
	log.Infof("starting load balancer on %s", a.server.Addr)
	go a.waitForShutdown()

	if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Errorf("server error: %v", err)
		a.errCh <- err
	}

	return <-a.errCh
}

func (a *App) waitForShutdown() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigChan
	log.Infof("received signal: %v, initiating graceful shutdown", sig)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	a.healthService.Stop()
	log.Infof("health service stopped")

	if err := a.server.Shutdown(ctx); err != nil {
		log.Errorf("server shutdown error: %v", err)
		a.errCh <- err
	} else {
		log.Infof("server gracefully shutdown")
		a.errCh <- nil
	}
}
