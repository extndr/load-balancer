package main

import (
	"fmt"
	"os"

	"github.com/extndr/load-balancer/internal/app"
	"github.com/extndr/load-balancer/internal/config"
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

	application, err := app.New(cfg, logger)
	if err != nil {
		logger.Fatal("failed to initialize app", zap.Error(err))
	}

	if err := application.Run(); err != nil {
		logger.Fatal("server failed", zap.Error(err))
	}
}
