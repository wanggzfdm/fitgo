package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// Config represents the standalone MCP service configuration.
type Config struct {
	Coros CorosConfig `json:"coros"`
}

// CorosConfig represents the Coros service configuration
type CorosConfig struct {
	Account     string `json:"account"`
	Username    string `json:"username"`
	AccountType int    `json:"accountType"`
	Password    string `json:"password"`
	P1          string `json:"p1"`
	P2          string `json:"p2"`
	Address     string `json:"address"`
}

func (c CorosConfig) LoginAccount() string {
	if c.Account != "" {
		return c.Account
	}
	return c.Username
}

func (c CorosConfig) LoginAccountType() int {
	if c.AccountType != 0 {
		return c.AccountType
	}
	return 2
}

// LoadConfig loads the configuration from a JSON file
func LoadConfig(filepath string) (*Config, error) {
	// Read the config file
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse the JSON data
	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

// LoadConfigWithDefaults loads the configuration with fallback paths
func LoadConfigWithDefaults(primaryPath, fallbackPath string) (*Config, error) {
	// Try primary path first
	cfg, err := LoadConfig(primaryPath)
	if err == nil {
		return cfg, nil
	}

	// Try fallback path
	cfg, err = LoadConfig(fallbackPath)
	if err == nil {
		return cfg, nil
	}

	// Return the error from the fallback attempt
	return nil, fmt.Errorf("failed to load config from both %s and %s: %w", primaryPath, fallbackPath, err)
}

// LoadDefaultConfig loads the configuration using default paths.
func LoadDefaultConfig() (*Config, error) {
	if override := os.Getenv("COROS_CONFIG_PATH"); override != "" {
		return LoadConfig(override)
	}
	return LoadConfigWithDefaults("configs/config.json", "../../configs/config.json")
}
