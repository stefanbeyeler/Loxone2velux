package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds the application configuration
type Config struct {
	KLF200  KLF200Config  `yaml:"klf200"`
	Server  ServerConfig  `yaml:"server"`
	Loxone  LoxoneConfig  `yaml:"loxone"`
	Logging LoggingConfig `yaml:"logging"`
}

// KLF200Config holds KLF-200 connection settings
type KLF200Config struct {
	Host              string        `yaml:"host"`
	Port              int           `yaml:"port"`
	Password          string        `yaml:"password"`
	ReconnectInterval time.Duration `yaml:"reconnect_interval"`
	RefreshInterval   time.Duration `yaml:"refresh_interval"`
}

// ServerConfig holds HTTP server settings
type ServerConfig struct {
	Host         string        `yaml:"host"`
	Port         int           `yaml:"port"`
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`
	APIToken     string        `yaml:"api_token"`
}

// LoggingConfig holds logging settings
type LoggingConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"` // "json" or "console"
}

// LoxoneConfig holds Loxone integration settings
type LoxoneConfig struct {
	UDPFeedback UDPFeedbackConfig `yaml:"udp_feedback" json:"udp_feedback"`
	Mappings    []NodeMapping     `yaml:"mappings" json:"mappings"`
}

// UDPFeedbackConfig holds UDP feedback settings
type UDPFeedbackConfig struct {
	Enabled bool   `yaml:"enabled" json:"enabled"`
	IP      string `yaml:"ip" json:"ip"`
	Port    int    `yaml:"port" json:"port"`
}

// NodeMapping maps a KLF-200 node to a Loxone virtual input
type NodeMapping struct {
	ID       string `yaml:"id" json:"id"`
	Name     string `yaml:"name" json:"name"`
	NodeID   uint8  `yaml:"node_id" json:"node_id"`
	LoxoneID string `yaml:"loxone_id" json:"loxone_id"`
	Enabled  bool   `yaml:"enabled" json:"enabled"`
}

// DefaultConfig returns a config with default values
func DefaultConfig() *Config {
	return &Config{
		KLF200: KLF200Config{
			Host:              "192.168.1.100",
			Port:              51200,
			Password:          "",
			ReconnectInterval: 30 * time.Second,
			RefreshInterval:   5 * time.Minute,
		},
		Server: ServerConfig{
			Host:         "0.0.0.0",
			Port:         8080,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			APIToken:     "",
		},
		Loxone: LoxoneConfig{
			UDPFeedback: UDPFeedbackConfig{
				Enabled: false,
				IP:      "",
				Port:    7777,
			},
			Mappings: []NodeMapping{},
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "console",
		},
	}
}

// Load loads configuration from a YAML file
func Load(path string) (*Config, error) {
	cfg := DefaultConfig()

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

// LoadOrDefault loads configuration from file or returns defaults
func LoadOrDefault(path string) *Config {
	cfg, err := Load(path)
	if err != nil {
		return DefaultConfig()
	}
	return cfg
}

// Validate validates the configuration (allows missing KLF200 credentials for initial setup)
func (c *Config) Validate() error {
	if c.KLF200.Port <= 0 || c.KLF200.Port > 65535 {
		return fmt.Errorf("klf200.port must be between 1 and 65535")
	}
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return fmt.Errorf("server.port must be between 1 and 65535")
	}
	// API token is optional - if not set, no authentication required
	if c.Server.APIToken != "" && len(c.Server.APIToken) < 16 {
		return fmt.Errorf("server.api_token must be at least 16 characters if set")
	}
	if c.Loxone.UDPFeedback.Enabled {
		if c.Loxone.UDPFeedback.IP == "" {
			return fmt.Errorf("loxone.udp_feedback.ip is required when UDP feedback is enabled")
		}
		if c.Loxone.UDPFeedback.Port <= 0 || c.Loxone.UDPFeedback.Port > 65535 {
			return fmt.Errorf("loxone.udp_feedback.port must be between 1 and 65535")
		}
	}
	return nil
}

// IsKLF200Configured returns true if KLF200 host and password are set
func (c *Config) IsKLF200Configured() bool {
	return c.KLF200.Host != "" && c.KLF200.Password != ""
}

// Save saves the configuration to a YAML file
func (c *Config) Save(path string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
