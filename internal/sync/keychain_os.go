//go:build darwin || windows || linux

package sync

import (
	"github.com/zalando/go-keyring"
)

// OSKeychain implements Keychain using the OS-native credential store.
type OSKeychain struct {
	service string
}

// NewOSKeychain creates a new OSKeychain with the specified service name.
func NewOSKeychain(service string) *OSKeychain {
	return &OSKeychain{service: service}
}

// Store saves a credential securely in the OS keychain.
func (k *OSKeychain) Store(key, value string) error {
	return keyring.Set(k.service, key, value)
}

// Retrieve gets a stored credential from the OS keychain.
func (k *OSKeychain) Retrieve(key string) (string, error) {
	return keyring.Get(k.service, key)
}

// Delete removes a stored credential from the OS keychain.
func (k *OSKeychain) Delete(key string) error {
	return keyring.Delete(k.service, key)
}

// Exists checks if a credential exists in the OS keychain.
func (k *OSKeychain) Exists(key string) bool {
	_, err := keyring.Get(k.service, key)
	return err == nil
}
