package localtranslate

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"
)

// LLMProvider defines the interface for LLM providers
type LLMProvider interface {
	Translate(text, sourceLang, targetLang string) (string, error)
	GetType() ProviderType
	IsAvailable() bool
}

// MultiProviderEngine manages multiple LLM providers with fallback
type MultiProviderEngine struct {
	providers []LLMProvider
	config    *Config
	mu        sync.RWMutex
}

// NewMultiProviderEngine creates a new multi-provider translation engine
func NewMultiProviderEngine(config *Config) (*MultiProviderEngine, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	engine := &MultiProviderEngine{
		config: config,
	}

	// Initialize providers based on configuration
	var providers []LLMProvider

	// Sort enabled providers by priority
	enabledProviders := config.GetEnabledProviders()
	sort.Slice(enabledProviders, func(i, j int) bool {
		return enabledProviders[i].Priority < enabledProviders[j].Priority
	})

	// Initialize each enabled provider
	for _, pc := range enabledProviders {
		var provider LLMProvider
		var err error

		switch pc.Type {
		case ProviderOpenAI:
			provider, err = NewOpenAIProvider(&pc)
		case ProviderAnthropic:
			provider, err = NewAnthropicProvider(&pc)
		case ProviderDeepSeek:
			provider, err = NewDeepSeekProvider(&pc)
		case ProviderOllama:
			provider, err = NewOllamaProvider(&pc)
		}

		if err != nil {
			fmt.Printf("[MultiProvider] Failed to initialize %s provider: %v\n", pc.Type, err)
			continue
		}

		if provider != nil && provider.IsAvailable() {
			providers = append(providers, provider)
			fmt.Printf("[MultiProvider] ✅ Initialized %s provider\n", pc.Type)
		}
	}

	engine.providers = providers

	if len(providers) == 0 {
		return nil, fmt.Errorf("no available translation providers")
	}

	return engine, nil
}

// Translate tries providers in priority order until one succeeds
func (e *MultiProviderEngine) Translate(text, sourceLang, targetLang string) (string, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if len(e.providers) == 0 {
		return "", fmt.Errorf("no translation providers available")
	}

	var lastError error

	// Try each provider in order
	for i, provider := range e.providers {
		if !provider.IsAvailable() {
			continue
		}

		fmt.Printf("[MultiProvider] Trying provider %d/%d: %s\n", i+1, len(e.providers), provider.GetType())

		result, err := provider.Translate(text, sourceLang, targetLang)
		if err != nil {
			fmt.Printf("[MultiProvider] Provider %s failed: %v\n", provider.GetType(), err)
			lastError = err
			continue
		}

		fmt.Printf("[MultiProvider] ✅ Translation succeeded with %s\n", provider.GetType())
		return result, nil
	}

	return "", fmt.Errorf("all providers failed, last error: %w", lastError)
}

// GetAvailableProviders returns list of available provider types
func (e *MultiProviderEngine) GetAvailableProviders() []ProviderType {
	e.mu.RLock()
	defer e.mu.RUnlock()

	var available []ProviderType
	for _, p := range e.providers {
		if p.IsAvailable() {
			available = append(available, p.GetType())
		}
	}
	return available
}

// Close closes all providers
func (e *MultiProviderEngine) Close() error {
	// Most providers don't need explicit closing
	return nil
}

// ============================================================================
// OpenAI Provider
// ============================================================================

type OpenAIProvider struct {
	config    *ProviderConfig
	client    *http.Client
	available bool
}

func NewOpenAIProvider(config *ProviderConfig) (*OpenAIProvider, error) {
	apiKey := config.GetAPIKey()
	if apiKey == "" {
		return nil, fmt.Errorf("OpenAI API key not configured")
	}

	return &OpenAIProvider{
		config: config,
		client: &http.Client{Timeout: 30 * time.Second},
		available: true,
	}, nil
}

func (p *OpenAIProvider) GetType() ProviderType {
	return ProviderOpenAI
}

func (p *OpenAIProvider) IsAvailable() bool {
	return p.available && p.config.GetAPIKey() != ""
}

func (p *OpenAIProvider) Translate(text, sourceLang, targetLang string) (string, error) {
	targetLangName := getLanguageName(targetLang)
	prompt := fmt.Sprintf("Translate the following text to %s. Only output the translation, no explanations:\n\n%s", targetLangName, text)

	reqBody := map[string]interface{}{
		"model": p.config.Model,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
		"max_tokens": p.config.MaxTokens,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.config.GetAPIKey())

	resp, err := p.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("OpenAI API error: %s", string(bodyBytes))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	choices, ok := result["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	choice, ok := choices[0].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid choice format")
	}

	message, ok := choice["message"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid message format")
	}

	content, ok := message["content"].(string)
	if !ok {
		return "", fmt.Errorf("invalid content format")
	}

	return strings.TrimSpace(content), nil
}

// ============================================================================
// Anthropic Provider
// ============================================================================

type AnthropicProvider struct {
	config    *ProviderConfig
	client    *http.Client
	available bool
}

func NewAnthropicProvider(config *ProviderConfig) (*AnthropicProvider, error) {
	apiKey := config.GetAPIKey()
	if apiKey == "" {
		return nil, fmt.Errorf("Anthropic API key not configured")
	}

	return &AnthropicProvider{
		config: config,
		client: &http.Client{Timeout: 30 * time.Second},
		available: true,
	}, nil
}

func (p *AnthropicProvider) GetType() ProviderType {
	return ProviderAnthropic
}

func (p *AnthropicProvider) IsAvailable() bool {
	return p.available && p.config.GetAPIKey() != ""
}

func (p *AnthropicProvider) Translate(text, sourceLang, targetLang string) (string, error) {
	targetLangName := getLanguageName(targetLang)
	prompt := fmt.Sprintf("Translate the following text to %s. Only output the translation, no explanations:\n\n%s", targetLangName, text)

	reqBody := map[string]interface{}{
		"model": p.config.Model,
		"max_tokens": p.config.MaxTokens,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", bytes.NewReader(body))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", p.config.GetAPIKey())
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := p.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("Anthropic API error: %s", string(bodyBytes))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	content, ok := result["content"].([]interface{})
	if !ok || len(content) == 0 {
		return "", fmt.Errorf("no content in response")
	}

	firstContent, ok := content[0].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid content format")
	}

	text_content, ok := firstContent["text"].(string)
	if !ok {
		return "", fmt.Errorf("invalid text format")
	}

	return strings.TrimSpace(text_content), nil
}

// ============================================================================
// DeepSeek Provider
// ============================================================================

type DeepSeekProvider struct {
	config    *ProviderConfig
	client    *http.Client
	available bool
}

func NewDeepSeekProvider(config *ProviderConfig) (*DeepSeekProvider, error) {
	apiKey := config.GetAPIKey()
	if apiKey == "" {
		return nil, fmt.Errorf("DeepSeek API key not configured")
	}

	return &DeepSeekProvider{
		config: config,
		client: &http.Client{Timeout: 30 * time.Second},
		available: true,
	}, nil
}

func (p *DeepSeekProvider) GetType() ProviderType {
	return ProviderDeepSeek
}

func (p *DeepSeekProvider) IsAvailable() bool {
	return p.available && p.config.GetAPIKey() != ""
}

func (p *DeepSeekProvider) Translate(text, sourceLang, targetLang string) (string, error) {
	targetLangName := getLanguageName(targetLang)
	prompt := fmt.Sprintf("Translate the following text to %s. Only output the translation, no explanations:\n\n%s", targetLangName, text)

	reqBody := map[string]interface{}{
		"model": p.config.Model,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
		"max_tokens": p.config.MaxTokens,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	baseURL := p.config.BaseURL
	if baseURL == "" {
		baseURL = "https://api.deepseek.com"
	}

	req, err := http.NewRequest("POST", baseURL+"/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.config.GetAPIKey())

	resp, err := p.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("DeepSeek API error: %s", string(bodyBytes))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	choices, ok := result["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	choice, ok := choices[0].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid choice format")
	}

	message, ok := choice["message"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid message format")
	}

	content, ok := message["content"].(string)
	if !ok {
		return "", fmt.Errorf("invalid content format")
	}

	return strings.TrimSpace(content), nil
}

// ============================================================================
// Ollama Provider
// ============================================================================

type OllamaProvider struct {
	config    *ProviderConfig
	client    *http.Client
	available bool
}

func NewOllamaProvider(config *ProviderConfig) (*OllamaProvider, error) {
	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}

	provider := &OllamaProvider{
		config: config,
		client: &http.Client{Timeout: 60 * time.Second},
	}

	// Check if Ollama is running
	resp, err := provider.client.Get(baseURL + "/api/tags")
	if err != nil {
		return nil, fmt.Errorf("Ollama server not reachable: %w", err)
	}
	defer resp.Body.Close()

	provider.available = resp.StatusCode == http.StatusOK
	return provider, nil
}

func (p *OllamaProvider) GetType() ProviderType {
	return ProviderOllama
}

func (p *OllamaProvider) IsAvailable() bool {
	return p.available
}

func (p *OllamaProvider) Translate(text, sourceLang, targetLang string) (string, error) {
	targetLangName := getLanguageName(targetLang)
	prompt := fmt.Sprintf("Translate the following text to %s. Only output the translation, no explanations:\n\n%s", targetLangName, text)

	reqBody := map[string]interface{}{
		"model": p.config.Model,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
		"stream": false,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	baseURL := p.config.BaseURL
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}

	req, err := http.NewRequest("POST", baseURL+"/api/chat", bytes.NewReader(body))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("Ollama API error: %s", string(bodyBytes))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	message, ok := result["message"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid message format")
	}

	content, ok := message["content"].(string)
	if !ok {
		return "", fmt.Errorf("invalid content format")
	}

	return strings.TrimSpace(content), nil
}
