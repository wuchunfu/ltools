package bookmark

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	cacheFileName = "bookmark_cache.json"
	cacheVersion  = 1
)

// Cache manages bookmark caching
type Cache struct {
	dataDir    string
	cacheFile  string
	expiryDays int
}

// NewCache creates a new cache instance
func NewCache(dataDir string) (*Cache, error) {
	cacheDir := filepath.Join(dataDir, "bookmark")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	return &Cache{
		dataDir:    dataDir,
		cacheFile:  filepath.Join(cacheDir, cacheFileName),
		expiryDays: 7, // 默认 7 天
	}, nil
}

// Load loads cached bookmark data
func (c *Cache) Load() (*CacheData, error) {
	data, err := os.ReadFile(c.cacheFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // 缓存不存在
		}
		return nil, fmt.Errorf("failed to read cache file: %w", err)
	}

	var cacheData CacheData
	if err := json.Unmarshal(data, &cacheData); err != nil {
		return nil, fmt.Errorf("failed to parse cache file: %w", err)
	}

	return &cacheData, nil
}

// Save saves bookmark data to cache
func (c *Cache) Save(data *CacheData) error {
	data.Version = cacheVersion
	data.LastSync = time.Now()

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal cache data: %w", err)
	}

	if err := os.WriteFile(c.cacheFile, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}

	return nil
}

// IsExpired checks if the cache has expired
func (c *Cache) IsExpired() bool {
	data, err := c.Load()
	if err != nil || data == nil {
		return true
	}

	expiryTime := data.LastSync.AddDate(0, 0, c.expiryDays)
	return time.Now().After(expiryTime)
}

// Clear removes the cache file
func (c *Cache) Clear() error {
	if err := os.Remove(c.cacheFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove cache file: %w", err)
	}
	return nil
}

// Status returns the current cache status
func (c *Cache) Status() map[string]interface{} {
	data, err := c.Load()
	if err != nil || data == nil {
		return map[string]interface{}{
			"available": false,
		}
	}

	return map[string]interface{}{
		"available":     true,
		"last_sync":     data.LastSync.Format("2006-01-02 15:04:05"),
		"total_count":   len(data.Bookmarks),
		"browser_stats": data.BrowserStats,
		"is_expired":    c.IsExpired(),
	}
}
