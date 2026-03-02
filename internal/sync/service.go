package sync

import (
	"context"
	"fmt"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// SyncService exposes synchronization functionality to the frontend.
type SyncService struct {
	app     *application.App
	manager *SyncManager
}

// NewSyncService creates a new SyncService.
func NewSyncService(app *application.App, dataDir string) (*SyncService, error) {
	manager, err := NewSyncManager(dataDir)
	if err != nil {
		return nil, err
	}

	return &SyncService{
		app:     app,
		manager: manager,
	}, nil
}

// ServiceStartup is called when the application starts.
func (s *SyncService) ServiceStartup(app *application.App) error {
	// Start auto-sync if configured
	cfg := s.manager.GetConfig()
	if cfg.Enabled && cfg.AutoSync {
		return s.manager.StartAutoSync()
	}
	return nil
}

// ServiceShutdown is called when the application shuts down.
func (s *SyncService) ServiceShutdown(app *application.App) error {
	// Perform final sync before shutdown with timeout protection
	cfg := s.manager.GetConfig()
	if cfg.Enabled {
		// Create timeout context (10 seconds)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Run sync in goroutine
		done := make(chan struct{})
		go func() {
			result := s.manager.Sync()
			if !result.Success {
				fmt.Printf("[SyncService] Shutdown sync failed: %s\n", result.Error)
			} else {
				fmt.Printf("[SyncService] Shutdown sync completed: %s\n", result.Message)
			}
			close(done)
		}()

		// Wait for completion or timeout
		select {
		case <-done:
			// Sync completed successfully
		case <-ctx.Done():
			fmt.Println("[SyncService] Shutdown sync timeout after 10 seconds, continuing...")
		}
	}

	// Stop auto-sync
	s.manager.StopAutoSync()
	return nil
}

// GetConfig returns the current synchronization configuration.
func (s *SyncService) GetConfig() *SyncConfig {
	return s.manager.GetConfig()
}

// SetConfig updates the synchronization configuration.
func (s *SyncService) SetConfig(config *SyncConfig) error {
	return s.manager.SetConfig(config)
}

// Sync performs a manual synchronization.
func (s *SyncService) Sync() *SyncResult {
	return s.manager.Sync()
}

// GetStatus returns the current synchronization status.
func (s *SyncService) GetStatus() *SyncStatus {
	return s.manager.GetStatus()
}

// TestConnection tests the connection to a repository.
func (s *SyncService) TestConnection(url string) (*ConnectionTestResult, error) {
	return s.manager.TestConnection(url)
}

// EnableAutoSync enables automatic synchronization.
func (s *SyncService) EnableAutoSync() error {
	cfg := s.manager.GetConfig()
	cfg.AutoSync = true
	return s.manager.SetConfig(cfg)
}

// DisableAutoSync disables automatic synchronization.
func (s *SyncService) DisableAutoSync() error {
	cfg := s.manager.GetConfig()
	cfg.AutoSync = false
	return s.manager.SetConfig(cfg)
}

// EnableSync enables synchronization.
func (s *SyncService) EnableSync() error {
	cfg := s.manager.GetConfig()
	cfg.Enabled = true
	return s.manager.SetConfig(cfg)
}

// DisableSync disables synchronization.
func (s *SyncService) DisableSync() error {
	cfg := s.manager.GetConfig()
	cfg.Enabled = false
	return s.manager.SetConfig(cfg)
}

// StoreToken stores a Git access token securely.
func (s *SyncService) StoreToken(token string) error {
	return s.manager.StoreToken(token)
}

// HasToken checks if a token is stored.
func (s *SyncService) HasToken() bool {
	return s.manager.HasToken()
}

// DeleteToken removes the stored token.
func (s *SyncService) DeleteToken() error {
	return s.manager.DeleteToken()
}

// IsGitInstalled checks if Git is available on the system.
func (s *SyncService) IsGitInstalled() bool {
	return s.manager.IsGitInstalled()
}

// CheckSSHCredential checks if SSH credentials are available.
func (s *SyncService) CheckSSHCredential() bool {
	return s.manager.CheckSSHCredential()
}

// GetIgnorePatterns returns all ignore patterns.
func (s *SyncService) GetIgnorePatterns() []string {
	return s.manager.GetIgnorePatterns()
}

// AddIgnorePattern adds a custom ignore pattern.
func (s *SyncService) AddIgnorePattern(pattern string) {
	s.manager.AddIgnorePattern(pattern)
}

// SetRepoURL sets the repository URL.
func (s *SyncService) SetRepoURL(url string) error {
	cfg := s.manager.GetConfig()
	cfg.RepoURL = url
	return s.manager.SetConfig(cfg)
}

// SetSyncInterval sets the sync interval in minutes.
func (s *SyncService) SetSyncInterval(minutes int) error {
	cfg := s.manager.GetConfig()
	cfg.SyncInterval = minutes
	return s.manager.SetConfig(cfg)
}
