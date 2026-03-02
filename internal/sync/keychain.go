package sync

// Keychain provides a cross-platform interface for secure credential storage.
type Keychain interface {
	// Store saves a credential securely.
	Store(key, value string) error

	// Retrieve gets a stored credential.
	Retrieve(key string) (string, error)

	// Delete removes a stored credential.
	Delete(key string) error

	// Exists checks if a credential exists.
	Exists(key string) bool
}
