package sync

import (
	"errors"
	"sync"
)

// MemoryKeychain implements Keychain using in-memory storage.
// Useful for testing or environments without OS keychain support.
type MemoryKeychain struct {
	mu     sync.RWMutex
	store  map[string]string
}

// NewMemoryKeychain creates a new MemoryKeychain.
func NewMemoryKeychain() *MemoryKeychain {
	return &MemoryKeychain{
		store: make(map[string]string),
	}
}

// Store saves a credential in memory.
func (k *MemoryKeychain) Store(key, value string) error {
	k.mu.Lock()
	defer k.mu.Unlock()
	k.store[key] = value
	return nil
}

// Retrieve gets a stored credential from memory.
func (k *MemoryKeychain) Retrieve(key string) (string, error) {
	k.mu.RLock()
	defer k.mu.RUnlock()
	value, exists := k.store[key]
	if !exists {
		return "", ErrCredentialNotFound
	}
	return value, nil
}

// Delete removes a stored credential from memory.
func (k *MemoryKeychain) Delete(key string) error {
	k.mu.Lock()
	defer k.mu.Unlock()
	delete(k.store, key)
	return nil
}

// Exists checks if a credential exists in memory.
func (k *MemoryKeychain) Exists(key string) bool {
	k.mu.RLock()
	defer k.mu.RUnlock()
	_, exists := k.store[key]
	return exists
}

// ErrCredentialNotFound is returned when a credential is not found.
var ErrCredentialNotFound = errors.New("credential not found")
