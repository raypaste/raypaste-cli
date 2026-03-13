/*
Copyright © 2026 Raypaste
*/
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/raypaste/raypaste-cli/pkg/types"

	"github.com/spf13/viper"
)

// Config represents the application configuration
type Config struct {
	APIKey           string             `mapstructure:"api_key"`
	OpenRouterAPIKey string             `mapstructure:"openrouter_api_key"`
	CerebrasAPIKey   string             `mapstructure:"cerebras_api_key"`
	DefaultModel     string             `mapstructure:"default_model"`
	DefaultLength    types.OutputLength `mapstructure:"default_length"`
	AutoCopy         bool               `mapstructure:"auto_copy"`    // Deprecated: kept for backward compatibility
	DisableCopy      bool               `mapstructure:"disable_copy"` // New field to disable clipboard copying
	Models           map[string]Model   `mapstructure:"models"`
	Temperature      float64            `mapstructure:"temperature"`
}

var globalConfig *Config

// LoadConfig initializes and loads the configuration
func LoadConfig(configPath string) (*Config, error) {
	v := viper.New()

	// Set defaults
	v.SetDefault("default_model", "cerebras-gpt-oss-120b")
	v.SetDefault("default_length", "medium")
	v.SetDefault("auto_copy", false)    // Deprecated: kept for backward compatibility
	v.SetDefault("disable_copy", false) // Default: clipboard copying is enabled
	v.SetDefault("temperature", 0.7)

	// Set config name and paths
	v.SetConfigName("config")
	v.SetConfigType("yaml")

	// Get home directory
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	// Create ~/.raypaste directory if it doesn't exist
	configDir := filepath.Join(home, ".raypaste")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	// Create prompts directory
	promptsDir := filepath.Join(configDir, "prompts")
	if err := os.MkdirAll(promptsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create prompts directory: %w", err)
	}

	// Add config paths
	v.AddConfigPath(configDir)
	if configPath != "" {
		v.SetConfigFile(configPath)
	}

	// Environment variables
	v.SetEnvPrefix("RAYPASTE")
	v.AutomaticEnv()

	// Bind specific env vars
	_ = v.BindEnv("api_key", "RAYPASTE_API_KEY")
	_ = v.BindEnv("openrouter_api_key", "OPENROUTER_API_KEY")
	_ = v.BindEnv("cerebras_api_key", "CEREBRAS_API_KEY")
	_ = v.BindEnv("default_model", "RAYPASTE_DEFAULT_MODEL")
	_ = v.BindEnv("default_length", "RAYPASTE_DEFAULT_LENGTH")
	_ = v.BindEnv("disable_copy", "RAYPASTE_DISABLE_COPY")

	// Read config file (ignore if not found)
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	// Unmarshal config
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Initialize models map if nil
	if cfg.Models == nil {
		cfg.Models = make(map[string]Model)
	}

	globalConfig = &cfg
	return &cfg, nil
}

// GetConfig returns the global config instance
func GetConfig() *Config {
	if globalConfig == nil {
		// Try to load with defaults
		cfg, err := LoadConfig("")
		if err != nil {
			// Return a default config
			return &Config{
				DefaultModel:  "cerebras-gpt-oss-120b",
				DefaultLength: types.OutputLengthMedium,
				AutoCopy:      false,
				DisableCopy:   false,
				Temperature:   0.7,
				Models:        make(map[string]Model),
			}
		}
		return cfg
	}
	return globalConfig
}

// GetAPIKey retrieves the API key from config or environment
func (c *Config) GetAPIKey() string {
	if c.APIKey != "" {
		return c.APIKey
	}
	return os.Getenv("RAYPASTE_API_KEY")
}

// ProviderKey holds the resolved provider name and API key for a request
type ProviderKey struct {
	Provider string
	APIKey   string
}

// ResolveProviderKey determines which provider and API key to use for a given model.
func (c *Config) ResolveProviderKey(modelAlias string) (ProviderKey, error) {
	model, err := ResolveModel(modelAlias, c.Models)
	if err != nil {
		return ProviderKey{}, err
	}

	// 1. If model provider is "cerebras" and we have a Cerebras key, go direct
	if model.Provider == "cerebras" {
		if key := c.GetCerebrasAPIKey(); key != "" {
			return ProviderKey{Provider: "cerebras", APIKey: key}, nil
		}
	}

	// 2. If we have an OpenRouter key, route through OpenRouter
	if key := c.GetOpenRouterAPIKey(); key != "" {
		return ProviderKey{Provider: "openrouter", APIKey: key}, nil
	}

	// 3. Legacy fallback: treat api_key as OpenRouter if it looks like one
	if key := c.GetAPIKey(); key != "" {
		if strings.HasPrefix(key, "sk-or-") {
			fmt.Fprintln(os.Stderr, "Warning: using api_key as OpenRouter key. Please run: raypaste config set openrouter-api-key "+key)
			return ProviderKey{Provider: "openrouter", APIKey: key}, nil
		}
	}

	return ProviderKey{}, fmt.Errorf("no API key configured for provider %q. Set OPENROUTER_API_KEY or CEREBRAS_API_KEY", model.Provider)
}

// GetOpenRouterAPIKey retrieves the OpenRouter API key from config or environment
func (c *Config) GetOpenRouterAPIKey() string {
	if c.OpenRouterAPIKey != "" {
		return c.OpenRouterAPIKey
	}
	return os.Getenv("OPENROUTER_API_KEY")
}

// GetCerebrasAPIKey retrieves the Cerebras API key from config or environment
func (c *Config) GetCerebrasAPIKey() string {
	if c.CerebrasAPIKey != "" {
		return c.CerebrasAPIKey
	}
	return os.Getenv("CEREBRAS_API_KEY")
}

// HasAnyAPIKey returns true if any API key is configured (provider-specific or legacy)
func (c *Config) HasAnyAPIKey() bool {
	return c.GetOpenRouterAPIKey() != "" || c.GetCerebrasAPIKey() != "" || c.GetAPIKey() != ""
}

// GetDefaultModel returns the default model setting
func (c *Config) GetDefaultModel() string {
	if c.DefaultModel != "" {
		return c.DefaultModel
	}
	return "cerebras-gpt-oss-120b"
}

// GetDefaultLength returns the default output length setting
func (c *Config) GetDefaultLength() types.OutputLength {
	if c.DefaultLength != "" {
		return c.DefaultLength
	}
	return types.OutputLengthMedium
}

// ValidateOutputLength validates and returns a valid OutputLength
func ValidateOutputLength(length string) (types.OutputLength, error) {
	switch types.OutputLength(length) {
	case types.OutputLengthShort, types.OutputLengthMedium, types.OutputLengthLong:
		return types.OutputLength(length), nil
	default:
		return "", fmt.Errorf("invalid output length: %s (must be short, medium, or long)", length)
	}
}

// GetConfigDir returns the configuration directory path
func GetConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(home, ".raypaste"), nil
}

// GetPromptsDir returns the prompts directory path
func GetPromptsDir() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "prompts"), nil
}

// Save writes the current configuration to the config file
func (c *Config) Save() error {
	return c.SaveTo("")
}

// SaveTo writes the current configuration to a specific config file path
// If path is empty, writes to the default config location
func (c *Config) SaveTo(path string) error {
	configPath := path
	if configPath == "" {
		configDir, err := GetConfigDir()
		if err != nil {
			return err
		}
		configPath = filepath.Join(configDir, "config.yaml")
	}

	v := viper.New()
	v.SetConfigFile(configPath)

	// Set all current values
	v.Set("api_key", c.APIKey)
	v.Set("openrouter_api_key", c.OpenRouterAPIKey)
	v.Set("cerebras_api_key", c.CerebrasAPIKey)
	v.Set("default_model", c.DefaultModel)
	v.Set("default_length", c.DefaultLength)
	v.Set("disable_copy", c.DisableCopy)
	v.Set("temperature", c.Temperature)
	if c.Models != nil {
		v.Set("models", c.Models)
	}

	if err := v.WriteConfigAs(configPath); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// SetAPIKey sets the API key and saves to config
func (c *Config) SetAPIKey(key string) error {
	c.APIKey = key
	return c.Save()
}

// SetDefaultModel sets the default model and saves to config
func (c *Config) SetDefaultModel(model string) error {
	c.DefaultModel = model
	return c.Save()
}

// SetDefaultLength sets the default output length and saves to config
func (c *Config) SetDefaultLength(length types.OutputLength) error {
	c.DefaultLength = length
	return c.Save()
}

// SetDisableCopy sets the disable copy setting and saves to config
func (c *Config) SetDisableCopy(disable bool) error {
	c.DisableCopy = disable
	return c.Save()
}

// SetTemperature sets the temperature and saves to config
func (c *Config) SetTemperature(temp float64) error {
	c.Temperature = temp
	return c.Save()
}
