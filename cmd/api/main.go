package main

import (
	"os"

	"go-boilerplate/internal/config"
	"go-boilerplate/internal/observability"
	"go-boilerplate/internal/server"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		observability.Error("failed to load configuration", "error", err)
		os.Exit(1)
	}

	observability.Info("configuration loaded successfully")

	// Create and start server
	srv := server.New(cfg)
	if err := srv.Start(); err != nil {
		observability.Error("server failed to start", "error", err)
		os.Exit(1)
	}
}
