package localtranslate

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// ProviderType defines the type of translation provider
type ProviderType string

const (
	ProviderOpenAI    ProviderType = "openai"
	ProviderAnthropic ProviderType = "anthropic"
	ProviderDeepSeek  ProviderType = "deepseek"
	ProviderOllama    ProviderType = "ollama"
)

// ProviderConfig configures a specific translation provider
type ProviderConfig struct {
	Type      ProviderType `json:"type"`
	Enabled   bool         `json:"enabled"`
	APIKey    string       `json:"apiKey,omitempty"`    // Encrypted or from env
	BaseURL   string       `json:"baseUrl,omitempty"`   // Custom API endpoint
	Model     string       `json:"model,omitempty"`     // Model identifier
	MaxTokens int          `json:"maxTokens,omitempty"` // Max tokens for generation
	Priority  int          `json:"priority"`            // Lower = higher priority
}

// GetAPIKey returns the API key, preferring environment variable over config
func (c *ProviderConfig) GetAPIKey() string {
	// First, check environment variables based on provider type
	var envKey string
	switch c.Type {
	case ProviderOpenAI:
		envKey = os.Getenv("OPENAI_API_KEY")
	case ProviderAnthropic:
		envKey = os.Getenv("ANTHROPIC_API_KEY")
	case ProviderDeepSeek:
		envKey = os.Getenv("DEEPSEEK_API_KEY")
	}

	// Return environment variable if set
	if envKey != "" {
		return envKey
	}

	// Fall back to config value
	return c.APIKey
}

// Config manages local translation plugin configuration
type Config struct {
	Languages []LanguagePair   `json:"languages"` // Supported language pairs
	Providers []ProviderConfig `json:"providers"` // Multi-provider configuration
}

// GetConfigPath returns the path to the configuration file
func GetConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	configDir := filepath.Join(homeDir, ".ltools", "localtranslate")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", err
	}
	return filepath.Join(configDir, "config.json"), nil
}

// LoadConfig loads configuration from file or returns default
func LoadConfig() (*Config, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return default config if file doesn't exist
			return DefaultConfig(), nil
		}
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// Save saves the configuration to file
func (c *Config) Save() error {
	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		Languages: []LanguagePair{
			{SourceLang: "zh", TargetLang: "en", Name: "Chinese to English"},
			{SourceLang: "en", TargetLang: "zh", Name: "English to Chinese"},
			{SourceLang: "zh", TargetLang: "ja", Name: "Chinese to Japanese"},
			{SourceLang: "ja", TargetLang: "zh", Name: "Japanese to Chinese"},
			{SourceLang: "zh", TargetLang: "ko", Name: "Chinese to Korean"},
			{SourceLang: "ko", TargetLang: "zh", Name: "Korean to Chinese"},
		},
		Providers: []ProviderConfig{
			{
				Type:      ProviderOllama,
				Enabled:   false, // Default disabled - requires local Ollama service
				BaseURL:   "http://localhost:11434",
				Model:     "qwen2.5:3b",
				MaxTokens: 1024,
				Priority:  1,
			},
			{
				Type:      ProviderOpenAI,
				Enabled:   false,
				Model:     "gpt-4o-mini",
				MaxTokens: 1024,
				Priority:  2,
			},
			{
				Type:      ProviderAnthropic,
				Enabled:   false,
				Model:     "claude-3-5-sonnet-20241022",
				MaxTokens: 1024,
				Priority:  3,
			},
			{
				Type:      ProviderDeepSeek,
				Enabled:   false,
				Model:     "deepseek-chat",
				MaxTokens: 1024,
				Priority:  4,
			},
		},
	}
}

// GetSupportedLanguages returns the list of supported language pairs
func (c *Config) GetSupportedLanguages() []LanguagePair {
	return c.Languages
}

// GetEnabledProviders returns enabled providers sorted by priority
func (c *Config) GetEnabledProviders() []ProviderConfig {
	var enabled []ProviderConfig
	for _, p := range c.Providers {
		if p.Enabled {
			enabled = append(enabled, p)
		}
	}
	return enabled
}
