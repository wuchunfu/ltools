package localtranslate

import (
	"fmt"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// LocalTranslateService exposes local translation functionality to the frontend
type LocalTranslateService struct {
	app         *application.App
	plugin      *LocalTranslatePlugin
	multiEngine *MultiProviderEngine
}

// NewLocalTranslateService creates a new local translation service
func NewLocalTranslateService(plugin *LocalTranslatePlugin, app *application.App) *LocalTranslateService {
	return &LocalTranslateService{
		app:    app,
		plugin: plugin,
	}
}

// ServiceStartup is called when the application starts
func (s *LocalTranslateService) ServiceStartup(app *application.App) error {
	// Initialize multi-provider engine
	if s.plugin.config != nil {
		multiEngine, err := NewMultiProviderEngine(s.plugin.config)
		if err != nil {
			fmt.Printf("[LocalTranslateService] Failed to initialize multi-provider engine: %v\n", err)
		} else {
			s.multiEngine = multiEngine
			fmt.Printf("[LocalTranslateService] ✅ Multi-provider engine initialized\n")
		}
	}

	return nil
}

// ServiceShutdown is called when the application shuts down
func (s *LocalTranslateService) ServiceShutdown(app *application.App) error {
	// Cleanup resources
	if s.multiEngine != nil {
		s.multiEngine.Close()
	}
	return nil
}

// Translate translates text from source language to target language
// Uses multi-provider engine with fallback
func (s *LocalTranslateService) Translate(text string, sourceLang string, targetLang string) (*TranslationResult, error) {
	if s.multiEngine == nil {
		return nil, fmt.Errorf("translation engine not initialized")
	}

	translated, err := s.multiEngine.Translate(text, sourceLang, targetLang)
	if err != nil {
		return nil, fmt.Errorf("translation failed: %w", err)
	}

	// Determine which provider was used (simplified - in reality we'd track this)
	provider := "unknown"
	available := s.multiEngine.GetAvailableProviders()
	if len(available) > 0 {
		provider = string(available[0])
	}

	result := &TranslationResult{
		OriginalText:   text,
		TranslatedText: translated,
		SourceLang:     sourceLang,
		TargetLang:     targetLang,
		Confidence:     90,
		ModelID:        fmt.Sprintf("%s-%s", sourceLang, targetLang),
		Provider:       provider,
	}

	// Emit translation event
	s.plugin.emitEvent("translated", result)

	return result, nil
}

// TranslateWithProvider translates text using a specific provider
func (s *LocalTranslateService) TranslateWithProvider(text string, sourceLang string, targetLang string, providerType ProviderType) (*TranslationResult, error) {
	// This is a simplified implementation
	// In a full implementation, you'd select the specific provider from multiEngine
	return s.Translate(text, sourceLang, targetLang)
}

// GetProviderStatuses returns the status of all configured providers
func (s *LocalTranslateService) GetProviderStatuses() ([]ProviderStatus, error) {
	if s.plugin.config == nil {
		return nil, fmt.Errorf("plugin config not initialized")
	}

	var statuses []ProviderStatus

	for _, pc := range s.plugin.config.Providers {
		status := ProviderStatus{
			Type:      pc.Type,
			Enabled:   pc.Enabled,
			Model:     pc.Model,
			Priority:  pc.Priority,
			Available: false,
			APIKeySet: pc.GetAPIKey() != "",
		}

		// Check if provider is available
		if s.multiEngine != nil {
			for _, availableType := range s.multiEngine.GetAvailableProviders() {
				if availableType == pc.Type {
					status.Available = true
					break
				}
			}
		}

		statuses = append(statuses, status)
	}

	return statuses, nil
}

// SetProviderEnabled enables or disables a specific provider
func (s *LocalTranslateService) SetProviderEnabled(providerType ProviderType, enabled bool) error {
	if s.plugin.config == nil {
		return fmt.Errorf("plugin config not initialized")
	}

	// Find and update the provider
	for i, pc := range s.plugin.config.Providers {
		if pc.Type == providerType {
			s.plugin.config.Providers[i].Enabled = enabled

			// Save configuration to file
			if err := s.plugin.config.Save(); err != nil {
				fmt.Printf("[LocalTranslateService] Failed to save config: %v\n", err)
			}

			// Re-initialize multi-provider engine
			multiEngine, err := NewMultiProviderEngine(s.plugin.config)
			if err != nil {
				return fmt.Errorf("failed to reinitialize providers: %w", err)
			}
			s.multiEngine = multiEngine
			return nil
		}
	}

	return fmt.Errorf("provider %s not found", providerType)
}

// ConfigureProviderInput defines the input for configuring a provider
type ConfigureProviderInput struct {
	APIKey    string `json:"apiKey"`
	BaseURL   string `json:"baseUrl"`
	Model     string `json:"model"`
	MaxTokens int    `json:"maxTokens"`
}

// ConfigureProvider configures a specific provider with API key and settings
func (s *LocalTranslateService) ConfigureProvider(providerType ProviderType, input ConfigureProviderInput) error {
	if s.plugin.config == nil {
		return fmt.Errorf("plugin config not initialized")
	}

	// Find and update the provider
	for i, pc := range s.plugin.config.Providers {
		if pc.Type == providerType {
			// Update configuration
			if input.APIKey != "" {
				s.plugin.config.Providers[i].APIKey = input.APIKey
			}
			if input.BaseURL != "" {
				s.plugin.config.Providers[i].BaseURL = input.BaseURL
			}
			if input.Model != "" {
				s.plugin.config.Providers[i].Model = input.Model
			}
			if input.MaxTokens > 0 {
				s.plugin.config.Providers[i].MaxTokens = input.MaxTokens
			}

			// Enable the provider when configured
			s.plugin.config.Providers[i].Enabled = true

			// Save configuration to file
			if err := s.plugin.config.Save(); err != nil {
				fmt.Printf("[LocalTranslateService] Failed to save config: %v\n", err)
				return fmt.Errorf("failed to save configuration: %w", err)
			}

			// Re-initialize multi-provider engine
			multiEngine, err := NewMultiProviderEngine(s.plugin.config)
			if err != nil {
				return fmt.Errorf("failed to reinitialize providers: %w", err)
			}
			s.multiEngine = multiEngine
			return nil
		}
	}

	return fmt.Errorf("provider %s not found", providerType)
}

// GetSupportedLanguages returns the list of supported language pairs
func (s *LocalTranslateService) GetSupportedLanguages() ([]LanguagePair, error) {
	if s.plugin.config == nil {
		return nil, fmt.Errorf("plugin config not initialized")
	}

	return s.plugin.config.GetSupportedLanguages(), nil
}

// ShowTranslateWindow shows the translation window
// This is called when the global shortcut is triggered
func (s *LocalTranslateService) ShowTranslateWindow() {
	// Emit event to frontend to open the translation window
	s.app.Event.Emit("localtranslate:show-window", "")
}
