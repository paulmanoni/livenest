package core

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// Config holds application configuration
type Config struct {
	Debug          bool   `json:"debug" toml:"debug"`
	TemplateDir    string `json:"template_dir" toml:"template_dir"`
	StaticDir      string `json:"static_dir" toml:"static_dir"`
	SecretKey      string `json:"secret_key" toml:"secret_key"`
	LiveViewSecret string `json:"liveview_secret" toml:"liveview_secret"`

	Database DatabaseConfig `json:"database" toml:"database"`
	Server   ServerConfig   `json:"server" toml:"server"`
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Driver   string `json:"driver" toml:"driver"`
	Host     string `json:"host" toml:"host"`
	Port     int    `json:"port" toml:"port"`
	Database string `json:"database" toml:"database"`
	Username string `json:"username" toml:"username"`
	Password string `json:"password" toml:"password"`
	SSLMode  string `json:"ssl_mode" toml:"ssl_mode"`
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Host string `json:"host" toml:"host"`
	Port int    `json:"port" toml:"port"`
}

// DefaultConfig returns default configuration
func DefaultConfig() *Config {
	return &Config{
		Debug:          true,
		TemplateDir:    "templates",
		StaticDir:      "static",
		SecretKey:      "change-me-in-production",
		LiveViewSecret: "change-me-in-production",
		Database: DatabaseConfig{
			Driver:   "sqlite",
			Database: "livenest.db",
		},
		Server: ServerConfig{
			Host: "0.0.0.0",
			Port: 8080,
		},
	}
}

// LoadConfig loads configuration from a file (supports JSON and TOML)
func LoadConfig(path string) (*Config, error) {
	config := DefaultConfig()

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".json":
		if err := json.Unmarshal(data, config); err != nil {
			return nil, err
		}
	case ".toml":
		// TOML support will be added when network is available
		// For now, use JSON or implement custom TOML parser
		return nil, nil
	default:
		// Try JSON as default
		if err := json.Unmarshal(data, config); err != nil {
			return nil, err
		}
	}

	return config, nil
}

// LoadConfigOrDefault loads config from file or returns default if file doesn't exist
func LoadConfigOrDefault(path string) *Config {
	config, err := LoadConfig(path)
	if err != nil || config == nil {
		return DefaultConfig()
	}
	return config
}
