package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/rs/zerolog"

	"github.com/stefanbeyeler/loxone2velux/internal/api"
	"github.com/stefanbeyeler/loxone2velux/internal/config"
	"github.com/stefanbeyeler/loxone2velux/internal/gateway"
)

var version = "dev"

// ConfigManager manages configuration with persistence
type ConfigManager struct {
	cfg        *config.Config
	configPath string
	gateway    *gateway.Service
	mu         sync.RWMutex
	logger     zerolog.Logger
}

// NewConfigManager creates a new ConfigManager
func NewConfigManager(cfg *config.Config, configPath string, gw *gateway.Service, logger zerolog.Logger) *ConfigManager {
	return &ConfigManager{
		cfg:        cfg,
		configPath: configPath,
		gateway:    gw,
		logger:     logger,
	}
}

// GetConfig returns the current configuration
func (m *ConfigManager) GetConfig() *config.Config {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.cfg
}

// GetConfigPath returns the path to the config file
func (m *ConfigManager) GetConfigPath() string {
	return m.configPath
}

// UpdateConfig updates and saves the configuration
func (m *ConfigManager) UpdateConfig(cfg *config.Config) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Validate
	if err := cfg.Validate(); err != nil {
		return err
	}

	// Save to file
	if err := cfg.Save(m.configPath); err != nil {
		m.logger.Error().Err(err).Msg("Failed to save config file")
		// Continue anyway - config is updated in memory
	} else {
		m.logger.Info().Str("path", m.configPath).Msg("Configuration saved")
	}

	// Update gateway config if KLF-200 settings changed
	if m.cfg.KLF200.Host != cfg.KLF200.Host ||
		m.cfg.KLF200.Port != cfg.KLF200.Port ||
		m.cfg.KLF200.Password != cfg.KLF200.Password {
		m.gateway.UpdateConfig(&cfg.KLF200)
	}

	m.cfg = cfg
	return nil
}

func main() {
	// Parse command line flags
	configPath := flag.String("config", "config.yaml", "Path to configuration file")
	showVersion := flag.Bool("version", false, "Show version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Println("loxone2velux", version)
		os.Exit(0)
	}

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
		Str("version", version).
		Str("klf200_host", cfg.KLF200.Host).
		Int("klf200_port", cfg.KLF200.Port).
		Int("server_port", cfg.Server.Port).
		Msg("Starting Loxone2Velux Gateway")

	// Create gateway service
	gw := gateway.NewService(&cfg.KLF200, logger)

	// Start gateway (non-blocking, connects in background)
	ctx := context.Background()
	if err := gw.Start(ctx); err != nil {
		// Don't fail - will retry in background
		logger.Warn().Err(err).Msg("Initial KLF-200 connection failed, will retry in background")
	}

	// Create config manager
	configMgr := NewConfigManager(cfg, *configPath, gw, logger)

	// Create and start API server
	server := api.NewServer(&cfg.Server, gw, logger, configMgr)

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
