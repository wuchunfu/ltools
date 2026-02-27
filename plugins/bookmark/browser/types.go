package browser

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
