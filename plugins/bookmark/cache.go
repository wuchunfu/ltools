package bookmark

// Cache manages bookmark caching
// This is a stub implementation - will be fully implemented in Task 4
type Cache struct {
	dataDir string
}

// NewCache creates a new cache instance
func NewCache(dataDir string) (*Cache, error) {
	return &Cache{dataDir: dataDir}, nil
}

// Load loads cached bookmark data
func (c *Cache) Load() (*CacheData, error) {
	// Stub - will be implemented in Task 4
	return nil, nil
}

// IsExpired checks if the cache has expired
func (c *Cache) IsExpired() bool {
	// Stub - will be implemented in Task 4
	return true
}

// Save saves bookmark data to cache
func (c *Cache) Save(data *CacheData) error {
	// Stub - will be implemented in Task 4
	return nil
}
