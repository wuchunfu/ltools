package imagebed

import "time"

// ImageBedConfig represents the configuration for the image bed plugin
type ImageBedConfig struct {
	Version     int    `json:"version"`
	GitHubToken string `json:"githubToken"`
	Owner       string `json:"owner"`
	Repo        string `json:"repo"`
	Path        string `json:"path"`   // Default "images/"
	Branch      string `json:"branch"` // Default "main"
}

// UploadRecord represents a single image upload record
type UploadRecord struct {
	ID          string    `json:"id"`
	FileName    string    `json:"fileName"`
	OriginalURL string    `json:"originalUrl"` // GitHub raw URL
	CDNURL      string    `json:"cdnUrl"`      // jsDelivr CDN URL
	Size        int64     `json:"size"`
	UploadTime  time.Time `json:"uploadTime"`
	Path        string    `json:"path"`
	Sha         string    `json:"sha"` // GitHub blob sha for deletion
}

// UploadHistory represents the persistent upload history
type UploadHistory struct {
	Version int            `json:"version"`
	Records []UploadRecord `json:"records"`
}

// UploadResult represents the result of an upload operation
type UploadResult struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Record  *UploadRecord `json:"record,omitempty"`
}

// UploadRequest represents a request to upload an image
type UploadRequest struct {
	FileName string `json:"fileName"`
	Content  string `json:"content"` // Base64 encoded image data
}

// GitHubUploadResponse represents the response from GitHub API
type GitHubUploadResponse struct {
	Content struct {
		Name    string `json:"name"`
		Path    string `json:"path"`
		Sha     string `json:"sha"`
		Size    int    `json:"size"`
		HTMLURL string `json:"html_url"`
	} `json:"content"`
	Commit struct {
		Message string `json:"message"`
		Sha     string `json:"sha"`
	} `json:"commit"`
}

// GitHubErrorResponse represents an error response from GitHub API
type GitHubErrorResponse struct {
	Message string `json:"message"`
	Errors  []struct {
		Resource string `json:"resource"`
		Field    string `json:"field"`
		Code     string `json:"code"`
		Message  string `json:"message,omitempty"`
	} `json:"errors,omitempty"`
}

// ConfigValidationResult represents the result of validating the configuration
type ConfigValidationResult struct {
	Valid   bool   `json:"valid"`
	Message string `json:"message,omitempty"`
}

// LinkFormat represents different formats for sharing image links
type LinkFormat string

const (
	LinkFormatRaw     LinkFormat = "raw"     // Raw URL
	LinkFormatMarkdown LinkFormat = "markdown" // Markdown: ![alt](url)
	LinkFormatHTML    LinkFormat = "html"    // HTML: <img src="url" />
)

// LinkFormats returns all available link formats
func LinkFormats() []LinkFormat {
	return []LinkFormat{LinkFormatRaw, LinkFormatMarkdown, LinkFormatHTML}
}

// FormatLink formats a URL according to the specified format
func FormatLink(url, alt string, format LinkFormat) string {
	switch format {
	case LinkFormatMarkdown:
		return "![" + alt + "](" + url + ")"
	case LinkFormatHTML:
		return "<img src=\"" + url + "\" alt=\"" + alt + "\" />"
	default:
		return url
	}
}
