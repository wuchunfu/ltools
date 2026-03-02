package localtranslate

import (
	"fmt"
	"sync"
)

// EngineOptions configures the translation engine
type EngineOptions struct {
	Temperature float64
	MaxTokens   int
}

// DefaultEngineOptions returns default engine configuration
func DefaultEngineOptions() *EngineOptions {
	return &EngineOptions{
		Temperature: 0.1,
		MaxTokens:   1024,
	}
}

// TranslateEngine interface for translation engines
type TranslateEngine interface {
	Translate(text, sourceLang, targetLang string) (string, error)
	Close() error
}

// MockTranslateEngine is a mock implementation for development and testing
type MockTranslateEngine struct {
	opts   *EngineOptions
	mu     sync.Mutex
	closed bool
}

// NewMockTranslateEngine creates a new mock translation engine
func NewMockTranslateEngine(opts *EngineOptions) (*MockTranslateEngine, error) {
	if opts == nil {
		opts = DefaultEngineOptions()
	}
	return &MockTranslateEngine{opts: opts, closed: false}, nil
}

// Translate performs mock translation
func (e *MockTranslateEngine) Translate(text, sourceLang, targetLang string) (string, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.closed {
		return "", fmt.Errorf("engine is closed")
	}
	return fmt.Sprintf("[Mock翻译 %s->%s] %s", sourceLang, targetLang, text), nil
}

// Close releases engine resources
func (e *MockTranslateEngine) Close() error {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.closed = true
	return nil
}

// getLanguageName returns the human-readable language name
func getLanguageName(langCode string) string {
	languages := map[string]string{
		"zh": "中文",
		"en": "英文",
		"ja": "日文",
		"ko": "韩文",
		"fr": "法文",
		"de": "德文",
		"es": "西班牙文",
		"ru": "俄文",
	}
	if name, ok := languages[langCode]; ok {
		return name
	}
	return langCode
}
