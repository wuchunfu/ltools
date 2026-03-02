package localtranslate

// ModelStatus represents the status of a translation model
type ModelStatus struct {
	ID           string `json:"id"`           // Model identifier (e.g., "zh-en", "en-zh")
	Name         string `json:"name"`         // Display name (e.g., "Chinese to English")
	SourceLang   string `json:"sourceLang"`   // Source language code
	TargetLang   string `json:"targetLang"`   // Target language code
	IsDownloaded bool   `json:"isDownloaded"` // Whether model is downloaded
	IsLoading    bool   `json:"isLoading"`    // Whether model is currently loading
	Size         int64  `json:"size"`         // Model size in bytes
	Version      string `json:"version"`      // Model version
}

// TranslationResult represents the result of a translation
type TranslationResult struct {
	OriginalText   string `json:"originalText"`   // Original input text
	TranslatedText string `json:"translatedText"` // Translated text
	SourceLang     string `json:"sourceLang"`     // Detected or specified source language
	TargetLang     string `json:"targetLang"`     // Target language
	Confidence     int    `json:"confidence"`     // Confidence score (0-100)
	ModelID        string `json:"modelId"`        // Model used for translation
	Provider       string `json:"provider"`       // Provider used (e.g., "openai", "local")
}

// LanguagePair represents a supported language pair for translation
type LanguagePair struct {
	SourceLang string `json:"sourceLang"` // Source language code (e.g., "zh", "en")
	TargetLang string `json:"targetLang"` // Target language code
	Name       string `json:"name"`       // Display name (e.g., "Chinese to English")
}

// DownloadProgress represents the progress of a model download
type DownloadProgress struct {
	ModelID      string  `json:"modelId"`      // Model being downloaded
	BytesLoaded  int64   `json:"bytesLoaded"`  // Bytes downloaded so far
	BytesTotal   int64   `json:"bytesTotal"`   // Total bytes to download
	Progress     float64 `json:"progress"`     // Progress percentage (0-100)
	IsComplete   bool    `json:"isComplete"`   // Whether download is complete
	ErrorMessage string  `json:"errorMessage"` // Error message if download failed
}

// ProviderStatus represents the status of a translation provider
type ProviderStatus struct {
	Type       ProviderType `json:"type"`
	Enabled    bool         `json:"enabled"`
	Available  bool         `json:"available"`  // Whether provider is ready to use
	Model      string       `json:"model"`      // Model being used
	APIKeySet  bool         `json:"apiKeySet"`  // Whether API key is configured (without revealing the key)
	Priority   int          `json:"priority"`   // Priority in fallback chain
	Error      string       `json:"error"`      // Error message if unavailable
}

