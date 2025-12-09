package main

import (
	"github.com/extndr/load-balancer/internal/app"
	"github.com/extndr/load-balancer/internal/config"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Warn(".env file not found, using defaults")
	}

	cfg := config.Load()

	application, err := app.New(cfg)
	if err != nil {
		log.Fatalf("failed to initialize app: %v", err)
	}

	if err := application.Run(); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
