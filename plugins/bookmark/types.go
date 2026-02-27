package bookmark

import "time"

// Bookmark represents a browser bookmark
type Bookmark struct {
	ID          string    `json:"id"`           // Unique identifier (URL hash)
	Title       string    `json:"title"`        // Bookmark title
	URL         string    `json:"url"`          // Bookmark URL
	Folder      string    `json:"folder"`       // Folder path
	Browser     string    `json:"browser"`      // Source browser
	AddedAt     time.Time `json:"added_at"`     // Added timestamp
	PinyinTitle string    `json:"pinyin_title"` // Title pinyin (for search)
}

// SearchResult represents a search result
type SearchResult struct {
	Bookmark  Bookmark `json:"bookmark"`
	Score     int      `json:"score"`      // Match score (for sorting)
	MatchType string   `json:"match_type"` // "prefix" | "contains" | "pinyin"
}

// CacheData represents cached bookmark data
type CacheData struct {
	Bookmarks    []Bookmark     `json:"bookmarks"`
	LastSync     time.Time      `json:"last_sync"`
	BrowserStats map[string]int `json:"browser_stats"` // Bookmark count per browser
	Version      int            `json:"version"`
}

// CacheStatus represents the current cache status
type CacheStatus struct {
	Available    bool           `json:"available"`
	LastSync     string         `json:"last_sync"`
	TotalCount   int            `json:"total_count"`
	BrowserStats map[string]int `json:"browser_stats"`
	IsExpired    bool           `json:"is_expired"`
}

// BookmarkConfig represents plugin configuration
type BookmarkConfig struct {
	CacheExpiryDays int      `json:"cache_expiry_days"` // Cache validity period (days)
	MaxResults      int      `json:"max_results"`       // Maximum search results
	EnablePinyin    bool     `json:"enable_pinyin"`     // Enable pinyin search
	TriggerKeywords []string `json:"trigger_keywords"`  // Trigger keywords
}
