// Package sync provides Git-based synchronization for application data.
package sync

import "time"

// AuthMethod represents the authentication method for Git operations.
type AuthMethod string

const (
	AuthMethodAuto   AuthMethod = "auto"   // Auto-detect (SSH first, then HTTPS)
	AuthMethodSSH    AuthMethod = "ssh"    // SSH key authentication
	AuthMethodHTTPS  AuthMethod = "https"  // HTTPS with personal access token
)

// SyncStatus represents the current synchronization status.
type SyncStatus struct {
	// Syncing indicates if a sync operation is in progress.
	Syncing bool `json:"syncing"`

	// LastSyncTime is the timestamp of the last successful sync.
	LastSyncTime time.Time `json:"lastSyncTime"`

	// LastSyncHash is the Git commit hash of the last sync.
	LastSyncHash string `json:"lastSyncHash"`

	// Error contains the last error message, if any.
	Error string `json:"error,omitempty"`

	// HasChanges indicates if there are local changes to sync.
	HasChanges bool `json:"hasChanges"`

	// RemoteURL is the configured remote repository URL.
	RemoteURL string `json:"remoteUrl"`

	// Enabled indicates if sync is enabled.
	Enabled bool `json:"enabled"`

	// AutoSync indicates if automatic sync is enabled.
	AutoSync bool `json:"autoSync"`
}

// SyncResult represents the result of a sync operation.
type SyncResult struct {
	// Success indicates if the sync was successful.
	Success bool `json:"success"`

	// Message contains a human-readable result message.
	Message string `json:"message"`

	// FilesChanged is the number of files changed in this sync.
	FilesChanged int `json:"filesChanged"`

	// CommitHash is the new commit hash, if a commit was made.
	CommitHash string `json:"commitHash,omitempty"`

	// Error contains error details if sync failed.
	Error string `json:"error,omitempty"`
}

// ConnectionTestResult represents the result of a connection test.
type ConnectionTestResult struct {
	// Success indicates if the connection test was successful.
	Success bool `json:"success"`

	// Message contains a human-readable result message.
	Message string `json:"message"`

	// AuthMethod is the detected or used authentication method.
	AuthMethod AuthMethod `json:"authMethod"`
}

// SyncConfig represents the synchronization configuration.
type SyncConfig struct {
	// Enabled indicates if synchronization is enabled.
	Enabled bool `json:"enabled"`

	// RepoURL is the Git repository URL.
	RepoURL string `json:"repoUrl"`

	// AutoSync indicates if automatic synchronization is enabled.
	AutoSync bool `json:"autoSync"`

	// SyncInterval is the interval between automatic syncs in minutes.
	SyncInterval int `json:"syncInterval"`

	// LastSyncTime is the timestamp of the last successful sync.
	LastSyncTime time.Time `json:"lastSyncTime"`

	// LastSyncHash is the Git commit hash of the last sync.
	LastSyncHash string `json:"lastSyncHash"`

	// AuthMethod is the preferred authentication method.
	AuthMethod AuthMethod `json:"authMethod"`

	// IgnorePatterns are additional patterns to ignore.
	IgnorePatterns []string `json:"ignorePatterns,omitempty"`
}

// DefaultSyncConfig returns the default synchronization configuration.
func DefaultSyncConfig() *SyncConfig {
	return &SyncConfig{
		Enabled:       false,
		AutoSync:      true,
		SyncInterval:  5, // 5 minutes
		AuthMethod:    AuthMethodAuto,
		IgnorePatterns: []string{},
	}
}

// Keychain keys for storing credentials.
const (
	KeychainServiceName = "ltools-sync"
	KeychainTokenKey    = "git-token"
)
