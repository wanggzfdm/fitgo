package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// Config represents the application configuration
type Config struct {
	Server ServerConfig `json:"server"`
	App    AppConfig    `json:"app"`
	Coros  CorosConfig  `json:"coros"`
	AI     AIConfig     `json:"ai"`
}

// ServerConfig represents the server configuration
type ServerConfig struct {
	Port string `json:"port"`
	Host string `json:"host"`
}

// AppConfig represents the application configuration
type AppConfig struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// CorosConfig represents the Coros service configuration
type CorosConfig struct {
	Username int    `json:"username"`
	Password string `json:"password"`
	Address  string `json:"address"`
}
type AIConfig struct {
	Provider string `json:"provider"` // 提供者，如 "qwen"
	Config   struct {
		BaseURL string `json:"base_url"` // API 基础地址
		APIKey  string `json:"api_key"`  // API 密钥
		Model   string `json:"model"`    // 模型名称
		Timeout int    `json:"timeout"`  // 超时时间(秒)
	} `json:"config"`
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

// LoadDefaultConfig loads the configuration using default paths
func LoadDefaultConfig() (*Config, error) {
	return LoadConfigWithDefaults("configs/config.json", "../../configs/config.json")
}
