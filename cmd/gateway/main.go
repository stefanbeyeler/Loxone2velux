package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"

	"github.com/stefanbeyeler/loxone2velux/internal/api"
	"github.com/stefanbeyeler/loxone2velux/internal/config"
	"github.com/stefanbeyeler/loxone2velux/internal/gateway"
)

func main() {
	// Parse command line flags
	configPath := flag.String("config", "config.yaml", "Path to configuration file")
	flag.Parse()

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		// Try to use defaults if config file doesn't exist
		if os.IsNotExist(err) {
			cfg = config.DefaultConfig()
			// Check for environment variables
			if host := os.Getenv("KLF200_HOST"); host != "" {
				cfg.KLF200.Host = host
			}
			if password := os.Getenv("KLF200_PASSWORD"); password != "" {
				cfg.KLF200.Password = password
			}
		} else {
			panic("Failed to load configuration: " + err.Error())
		}
	}

	// Validate config
	if err := cfg.Validate(); err != nil {
		panic("Invalid configuration: " + err.Error())
	}

	// Setup logger
	logger := setupLogger(cfg.Logging)

	logger.Info().
		Str("klf200_host", cfg.KLF200.Host).
		Int("klf200_port", cfg.KLF200.Port).
		Int("server_port", cfg.Server.Port).
		Msg("Starting Loxone2Velux Gateway")

	// Create gateway service
	gw := gateway.NewService(&cfg.KLF200, logger)

	// Start gateway
	ctx := context.Background()
	if err := gw.Start(ctx); err != nil {
		logger.Fatal().Err(err).Msg("Failed to start gateway service")
	}

	// Create and start API server
	server := api.NewServer(&cfg.Server, gw, logger)

	// Start server in goroutine
	go func() {
		if err := server.Start(); err != nil {
			logger.Fatal().Err(err).Msg("Server error")
		}
	}()

	// Wait for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info().Msg("Shutting down...")

	// Graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error().Err(err).Msg("Server shutdown error")
	}

	if err := gw.Stop(); err != nil {
		logger.Error().Err(err).Msg("Gateway shutdown error")
	}

	logger.Info().Msg("Goodbye!")
}

func setupLogger(cfg config.LoggingConfig) zerolog.Logger {
	// Set log level
	level, err := zerolog.ParseLevel(cfg.Level)
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)

	// Set output format
	var logger zerolog.Logger
	if cfg.Format == "json" {
		logger = zerolog.New(os.Stdout)
	} else {
		logger = zerolog.New(zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		})
	}

	return logger.With().Timestamp().Caller().Logger()
}
