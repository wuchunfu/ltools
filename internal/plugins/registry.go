package plugins

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// Registry manages plugin metadata persistence
type Registry struct {
	mu        sync.RWMutex
	file      string
	plugins   map[string]*PluginMetadata
	dirty     bool
}

// NewRegistry creates a new plugin registry
func NewRegistry(dataDir string) (*Registry, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, err
	}

	file := filepath.Join(dataDir, "plugins.json")
	r := &Registry{
		file:    file,
		plugins: make(map[string]*PluginMetadata),
	}

	if err := r.load(); err != nil {
		// If file doesn't exist, that's okay - start fresh
		if !errors.Is(err, os.ErrNotExist) {
			return nil, err
		}
	}

	return r, nil
}

// load loads the registry from disk
func (r *Registry) load() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	data, err := os.ReadFile(r.file)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, &r.plugins)
}

// save saves the registry to disk
func (r *Registry) save() error {
	if !r.dirty {
		return nil
	}

	data, err := json.MarshalIndent(r.plugins, "", "  ")
	if err != nil {
		return err
	}

	// Write to temp file first, then rename for atomicity
	tmpFile := r.file + ".tmp"
	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		return err
	}

	return os.Rename(tmpFile, r.file)
}

// Register registers a plugin in the registry
func (r *Registry) Register(metadata *PluginMetadata) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// If plugin already exists in registry, preserve its state
	if existing, ok := r.plugins[metadata.ID]; ok {
		// Preserve the saved state and enabled status
		fmt.Printf("[Registry] Plugin %s already exists with state %s, preserving state\n", metadata.ID, existing.State)
		metadata.State = existing.State
		// Copy all other fields from existing metadata that should be preserved
		// (Currently only state needs to be preserved)
	} else {
		fmt.Printf("[Registry] Registering new plugin %s with state %s\n", metadata.ID, metadata.State)
	}

	r.plugins[metadata.ID] = metadata
	r.dirty = true

	return r.save()
}

// Unregister removes a plugin from the registry
func (r *Registry) Unregister(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.plugins, id)
	r.dirty = true

	return r.save()
}

// Get retrieves a plugin's metadata by ID
func (r *Registry) Get(id string) (*PluginMetadata, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	metadata, ok := r.plugins[id]
	return metadata, ok
}

// List returns all registered plugins
func (r *Registry) List() []*PluginMetadata {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*PluginMetadata, 0, len(r.plugins))
	for _, metadata := range r.plugins {
		result = append(result, metadata)
	}

	return result
}

// ListByType returns plugins of a specific type
func (r *Registry) ListByType(pluginType PluginType) []*PluginMetadata {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*PluginMetadata
	for _, metadata := range r.plugins {
		if metadata.Type == pluginType {
			result = append(result, metadata)
		}
	}

	return result
}

// ListByState returns plugins in a specific state
func (r *Registry) ListByState(state PluginState) []*PluginMetadata {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*PluginMetadata
	for _, metadata := range r.plugins {
		if metadata.State == state {
			result = append(result, metadata)
		}
	}

	return result
}

// Update updates a plugin's metadata
func (r *Registry) Update(metadata *PluginMetadata) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.plugins[metadata.ID]; !ok {
		return ErrPluginNotFound
	}

	// DEBUG: Log before update
	fmt.Printf("[Registry] Updating plugin %s, state = %s (pointer: %p)\n", metadata.ID, metadata.State, metadata)

	r.plugins[metadata.ID] = metadata
	r.dirty = true

	if err := r.save(); err != nil {
		fmt.Printf("[Registry] Failed to save: %v\n", err)
		return err
	}

	// DEBUG: Verify save worked
	fmt.Printf("[Registry] Successfully saved plugin %s\n", metadata.ID)

	return nil
}

// Search finds plugins matching the given keywords
func (r *Registry) Search(keywords ...string) []*PluginMetadata {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if len(keywords) == 0 {
		return r.List()
	}

	// Use a map to avoid duplicates
	seen := make(map[string]bool)
	var result []*PluginMetadata

	for _, metadata := range r.plugins {
		// Check if plugin matches any keyword
		for _, keyword := range keywords {
			if r.matchesKeyword(metadata, keyword) {
				if !seen[metadata.ID] {
					seen[metadata.ID] = true
					result = append(result, metadata)
				}
				break
			}
		}
	}

	return result
}

// matchesKeyword checks if a plugin matches a keyword
func (r *Registry) matchesKeyword(metadata *PluginMetadata, keyword string) bool {
	// Check name
	if contains(metadata.Name, keyword) {
		return true
	}

	// Check description
	if contains(metadata.Description, keyword) {
		return true
	}

	// Check author
	if contains(metadata.Author, keyword) {
		return true
	}

	// Check keywords
	for _, kw := range metadata.Keywords {
		if contains(kw, keyword) {
			return true
		}
	}

	return false
}

// contains is a case-insensitive substring check
func contains(s, substr string) bool {
	return containsIgnoreCase(s, substr)
}

// containsIgnoreCase checks if s contains substr (case-insensitive)
func containsIgnoreCase(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
		 len(substr) == 0 ||
		 (len(s) > 0 && containsIgnoreCaseHelper(s, substr)))
}

func containsIgnoreCaseHelper(s, substr string) bool {
	// Simple case-insensitive substring check
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			c1 := s[i+j]
			c2 := substr[j]
			if c1 >= 'A' && c1 <= 'Z' {
				c1 += 32
			}
			if c2 >= 'A' && c2 <= 'Z' {
				c2 += 32
			}
			if c1 != c2 {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

// Errors
var (
	ErrPluginNotFound = errors.New("plugin not found")
	ErrPluginExists   = errors.New("plugin already exists")
)
