package config

import (
	"errors"
	"fmt"
	"net"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration.
type Config struct {
	AppName         string        `yaml:"app_name"`
	ScanCIDR        string        `yaml:"scan_cidr"`
	ScanIntervalStr string        `yaml:"scan_interval"`
	TimeoutStr      string        `yaml:"timeout"`
	Concurrency     int           `yaml:"concurrency"`
	PortScanEnabled bool          `yaml:"port_scan_enabled"`
	PortsList       []int         `yaml:"ports_list"`
	DBPath          string        `yaml:"db_path"`
	LogLevel        string        `yaml:"log_level"`
	Theme           string        `yaml:"theme"`
	ScanInterval    time.Duration `yaml:"-"`
	Timeout         time.Duration `yaml:"-"`
}

// Default returns a Config with sensible default settings.
func Default() *Config {
	return &Config{
		AppName:         "WebMonkey",
		ScanCIDR:        "192.168.1.0/24",
		ScanIntervalStr: "5m",
		TimeoutStr:      "1s",
		Concurrency:     64,
		PortScanEnabled: false,
		PortsList:       []int{21, 22, 23, 25, 80, 110, 135, 139, 443, 445, 1433, 3306, 3389, 8080},
		DBPath:          "webmonkey.db",
		LogLevel:        "info",
		Theme:           "dark",
		ScanInterval:    5 * time.Minute,
		Timeout:         1 * time.Second,
	}
}

// Load reads and parses a YAML configuration file.
func Load(path string) (*Config, error) {
	cfg := Default()
	if path == "" {
		return cfg, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil // fallback to defaults if config file is absent
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	err = yaml.Unmarshal(data, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config YAML: %w", err)
	}

	// Parse duration strings
	if cfg.ScanIntervalStr != "" {
		d, err := time.ParseDuration(cfg.ScanIntervalStr)
		if err != nil {
			return nil, fmt.Errorf("invalid scan_interval: %w", err)
		}
		cfg.ScanInterval = d
	}

	if cfg.TimeoutStr != "" {
		d, err := time.ParseDuration(cfg.TimeoutStr)
		if err != nil {
			return nil, fmt.Errorf("invalid timeout: %w", err)
		}
		cfg.Timeout = d
	}

	// Validate config parameters
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return cfg, nil
}

// Validate checks configuration settings for correctness.
func (c *Config) Validate() error {
	if c.ScanCIDR != "" {
		_, _, err := net.ParseCIDR(c.ScanCIDR)
		if err != nil {
			return fmt.Errorf("invalid scan CIDR %q: %w", c.ScanCIDR, err)
		}
	}

	if c.Concurrency <= 0 {
		return errors.New("concurrency must be greater than 0")
	}

	if c.Timeout <= 0 {
		return errors.New("timeout must be greater than 0")
	}

	if c.DBPath == "" {
		return errors.New("db_path cannot be empty")
	}

	return nil
}
