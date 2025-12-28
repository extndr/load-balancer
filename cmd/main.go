package main

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/extndr/load-balancer/internal/balancer"
	"github.com/extndr/load-balancer/internal/config"
	"github.com/extndr/load-balancer/internal/healthcheck"
	"github.com/extndr/load-balancer/internal/middleware"
	"github.com/extndr/load-balancer/internal/proxy"
	"github.com/extndr/load-balancer/internal/server"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	if err := godotenv.Load(); err != nil {
		logger.Warn(".env file not found, using defaults")
	}

	cfg := config.Load()

	// Convert backend URLs to []*balancer.Backend
	backends := make([]*balancer.Backend, 0, len(cfg.BackendURLs))
	for _, backendURL := range cfg.BackendURLs {
		parsedURL, err := url.Parse(backendURL)
		if err != nil {
			logger.Fatal("failed to parse backend URL", zap.String("url", backendURL), zap.Error(err))
		}
		backends = append(backends, &balancer.Backend{URL: parsedURL})
	}

	bl := balancer.New(backends, logger)

	var hc *healthcheck.HealthChecker
	if cfg.EnableHealthCheck {
		hc = healthcheck.NewHealthChecker(
			cfg.BackendURLs,
			cfg.HealthCheckTimeout,
			cfg.HealthCheckInterval,
			logger,
		)
	}

	rproxy := proxy.NewProxy(cfg.ProxyTimeout, cfg.HTTPTransport)
	handler := server.NewHandler(bl, rproxy)
	chain := middleware.Chain(handler, middleware.Logging(logger))

	srv := server.NewServer(":"+cfg.Port, chain)

	serverErr := make(chan error, 1)

	logger.Info("starting load balancer", zap.String("addr", srv.Addr))
	go func() {
		err := srv.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			serverErr <- err
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if hc != nil {
		events := hc.Start(ctx)
		go bl.HandleHealthEvents(ctx, events)
	}

	select {
	case <-ctx.Done():
		logger.Info("received shutdown signal")
	case err := <-serverErr:
		logger.Error("server error", zap.Error(err))
		os.Exit(1)
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	logger.Info("stopping health check")
	if hc != nil {
		// Health checker will stop when context is cancelled
	}

	logger.Info("shutting down server")
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("server shutdown error", zap.Error(err))
		os.Exit(1)
	}

	logger.Info("server gracefully shutdown")
}
