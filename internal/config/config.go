package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	UniFi    UniFiConfig    `yaml:"unifi"`
	Tracker  TrackerConfig  `yaml:"tracker"`
}

// ServerConfig holds HTTP server settings
type ServerConfig struct {
	Address string `yaml:"address"`
	// Auth settings for the web UI
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

// DatabaseConfig holds database settings
type DatabaseConfig struct {
	Path string `yaml:"path"`
}

// UniFiConfig holds UniFi controller settings
type UniFiConfig struct {
	URL      string `yaml:"url"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Site     string `yaml:"site"`
	IsUDM    bool   `yaml:"is_udm"`
	Insecure bool   `yaml:"insecure"`
}

// TrackerConfig holds traffic tracker settings
type TrackerConfig struct {
	PollInterval    time.Duration `yaml:"poll_interval"`
	ActivityMinBytes int64        `yaml:"activity_min_bytes"`
}

// Load reads and parses the configuration file
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		// Defaults
		Server: ServerConfig{
			Address: ":8765",
		},
		Database: DatabaseConfig{
			Path: "zeitpolizei.db",
		},
		UniFi: UniFiConfig{
			Site: "default",
		},
		Tracker: TrackerConfig{
			PollInterval:     30 * time.Second,
			ActivityMinBytes: 1024, // 1 KB minimum to count as active
		},
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// ExampleConfig returns a sample configuration
func ExampleConfig() string {
	return `# Zeitpolizei Configuration

server:
  address: ":8765"
  username: "admin"
  password: "changeme"

database:
  path: "/data/zeitpolizei/zeitpolizei.db"

unifi:
  url: "https://192.168.1.1"
  username: "admin"
  password: "your-unifi-password"
  site: "default"
  is_udm: true    # Set to true for UDM/UDM Pro/SE
  insecure: true  # Skip TLS verification for self-signed certs

tracker:
  poll_interval: 30s
  activity_min_bytes: 1024  # Minimum bytes to count as active minute
`
}
