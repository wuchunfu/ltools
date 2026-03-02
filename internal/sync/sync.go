package sync

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// SyncManager coordinates the synchronization process.
type SyncManager struct {
	mu         sync.RWMutex
	config     *ConfigManager
	git        *GitClient
	keychain   Keychain
	dataDir    string
	syncDir    string // .sync subdirectory for Git repo
	ignore     *IgnoreRules
	ticker     *time.Ticker
	stopChan   chan struct{}
	running    bool
	syncing    bool
	lastError  error
}

// NewSyncManager creates a new SyncManager.
func NewSyncManager(dataDir string) (*SyncManager, error) {
	syncDir := filepath.Join(dataDir, ".sync")

	config, err := NewConfigManager(dataDir)
	if err != nil {
		return nil, fmt.Errorf("failed to create config manager: %w", err)
	}

	// Create OS keychain for credential storage
	keychain := NewOSKeychain(KeychainServiceName)

	return &SyncManager{
		config:   config,
		git:      NewGitClient(syncDir),
		keychain: keychain,
		dataDir:  dataDir,
		syncDir:  syncDir,
		ignore:   NewIgnoreRules(),
		stopChan: make(chan struct{}),
	}, nil
}

// Sync performs a full synchronization.
func (m *SyncManager) Sync() *SyncResult {
	m.mu.Lock()
	if m.syncing {
		m.mu.Unlock()
		return &SyncResult{
			Success: false,
			Error:   "sync already in progress",
		}
	}
	m.syncing = true
	m.mu.Unlock()

	defer func() {
		m.mu.Lock()
		m.syncing = false
		m.mu.Unlock()
	}()

	result := &SyncResult{}

	// Check if sync is enabled
	cfg := m.config.Get()
	if !cfg.Enabled {
		result.Success = false
		result.Error = "同步功能未启用"
		return result
	}

	// Check if git is installed
	if !m.git.IsGitInstalled() {
		result.Success = false
		result.Error = "未安装 Git"
		return result
	}

	// Ensure repository is set up
	if err := m.ensureRepo(cfg); err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("设置仓库失败: %v", err)
		m.lastError = err
		return result
	}

	// Pull remote changes first (if repo exists and has commits)
	if m.git.IsRepo() {
		if _, err := m.git.GetCommitHash(); err == nil {
			// Repo has commits, try to pull
			fmt.Printf("[SyncManager] Pulling remote changes...\n")
			if err := m.git.Pull(); err != nil {
				// Non-fatal: might be no remote commits yet
				fmt.Printf("[SyncManager] Pull warning (non-fatal): %v\n", err)
			}
		}
	}

	// Copy files from data directory to sync directory
	filesChanged, err := m.copyFilesToSyncDir()
	if err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("复制文件失败: %v", err)
		m.lastError = err
		return result
	}
	result.FilesChanged = filesChanged

	// Check if there are changes to commit
	hasChanges, err := m.git.HasChanges()
	if err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("检查变更失败: %v", err)
		return result
	}

	if !hasChanges {
		result.Success = true
		result.Message = "没有变更需要同步"
		return result
	}

	// Stage all changes
	if err := m.git.AddAll(); err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("暂存变更失败: %v", err)
		return result
	}

	// Create commit
	commitMsg := fmt.Sprintf("sync: %s", time.Now().Format("2006-01-02 15:04:05"))
	if err := m.git.Commit(commitMsg); err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("提交失败: %v", err)
		return result
	}

	// Try to push (without force first)
	if err := m.git.Push(false); err != nil {
		fmt.Printf("[SyncManager] Normal push failed, pulling and retrying: %v\n", err)

		// Normal push failed, might be remote has new commits
		// Pull again to get latest changes
		if pullErr := m.git.Pull(); pullErr != nil {
			fmt.Printf("[SyncManager] Pull failed: %v\n", pullErr)
			// If pull also fails, try force-with-lease as last resort
			if forceErr := m.git.Push(true); forceErr != nil {
				result.Success = false
				result.Error = fmt.Sprintf("推送失败（尝试强制推送也失败）: %v", forceErr)
				return result
			}
			fmt.Printf("[SyncManager] Force push succeeded\n")
		} else {
			// Pull succeeded, try normal push again
			if retryErr := m.git.Push(false); retryErr != nil {
				// Still failed, use force-with-lease as last resort
				fmt.Printf("[SyncManager] Retry push failed, using force-with-lease: %v\n", retryErr)
				if forceErr := m.git.Push(true); forceErr != nil {
					result.Success = false
					result.Error = fmt.Sprintf("推送失败: %v", forceErr)
					return result
				}
			}
		}
	}

	// Get commit hash
	hash, _ := m.git.GetShortHash()
	result.CommitHash = hash

	// Update last sync info
	m.config.UpdateLastSync(hash)
	m.lastError = nil

	result.Success = true
	result.Message = fmt.Sprintf("已同步 %d 个文件", filesChanged)
	return result
}

// ensureRepo ensures the Git repository is properly set up.
func (m *SyncManager) ensureRepo(cfg *SyncConfig) error {
	// Check if sync directory exists as a Git repo
	if m.git.IsRepo() {
		// Update remote URL if changed
		currentURL, err := m.git.GetRemoteURL()
		if err != nil || currentURL != cfg.RepoURL {
			if err := m.git.SetRemote(cfg.RepoURL); err != nil {
				return err
			}
		}
		return nil
	}

	// Remove sync directory if it exists but is not a repo
	if _, err := os.Stat(m.syncDir); err == nil {
		if err := os.RemoveAll(m.syncDir); err != nil {
			return fmt.Errorf("failed to remove existing sync dir: %w", err)
		}
	}

	// Try to clone the repository
	if err := m.git.Clone(cfg.RepoURL); err != nil {
		// If clone fails (e.g., empty repo), initialize new repo
		fmt.Printf("[SyncManager] Clone failed, initializing new repo: %v\n", err)
		if err := m.git.Init(); err != nil {
			return fmt.Errorf("failed to init repo: %w", err)
		}
		if err := m.git.SetRemote(cfg.RepoURL); err != nil {
			return fmt.Errorf("failed to set remote: %w", err)
		}
	}

	// Write .gitignore
	if err := m.git.WriteGitignore(m.ignore.ToGitignore()); err != nil {
		return fmt.Errorf("failed to write .gitignore: %w", err)
	}

	return nil
}

// copyFilesToSyncDir copies files from data directory to sync directory.
// Returns the number of files copied.
func (m *SyncManager) copyFilesToSyncDir() (int, error) {
	count := 0

	// First, remove files in sync dir that no longer exist in data dir
	// This ensures deletions are also synchronized
	err := filepath.Walk(m.syncDir, func(syncPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Get relative path from sync directory
		relPath, err := filepath.Rel(m.syncDir, syncPath)
		if err != nil {
			return err
		}

		// Skip .git directory and root
		if relPath == "." || strings.HasPrefix(relPath, ".git") {
			return nil
		}

		// Check if this file/directory exists in data directory
		dataPath := filepath.Join(m.dataDir, relPath)
		if _, err := os.Stat(dataPath); os.IsNotExist(err) {
			// File/directory no longer exists in data dir, remove from sync dir
			fmt.Printf("[SyncManager] Removing deleted file: %s\n", relPath)
			if err := os.RemoveAll(syncPath); err != nil {
				return fmt.Errorf("failed to remove %s: %w", relPath, err)
			}
			if info.IsDir() {
				return filepath.SkipDir
			}
		}

		return nil
	})

	if err != nil {
		return 0, fmt.Errorf("failed to clean sync directory: %w", err)
	}

	// Then, copy files from data directory to sync directory
	err = filepath.Walk(m.dataDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Get relative path
		relPath, err := filepath.Rel(m.dataDir, path)
		if err != nil {
			return err
		}

		// Skip root
		if relPath == "." {
			return nil
		}

		// Check if should ignore
		if m.ignore.ShouldIgnore(relPath) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Destination path
		destPath := filepath.Join(m.syncDir, relPath)

		if info.IsDir() {
			// Create directory
			return os.MkdirAll(destPath, info.Mode())
		}

		// Copy file
		if err := m.copyFile(path, destPath); err != nil {
			return fmt.Errorf("failed to copy %s: %w", relPath, err)
		}
		count++

		return nil
	})

	return count, err
}

// copyFile copies a single file.
func (m *SyncManager) copyFile(src, dst string) error {
	// Ensure destination directory exists
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	// Open source file
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	// Get source file info for permissions
	srcInfo, err := srcFile.Stat()
	if err != nil {
		return err
	}

	// Create destination file
	dstFile, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return err
	}
	defer dstFile.Close()

	// Copy content
	_, err = io.Copy(dstFile, srcFile)
	return err
}

// StartAutoSync starts automatic synchronization at the configured interval.
func (m *SyncManager) StartAutoSync() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.running {
		return nil
	}

	cfg := m.config.Get()
	if !cfg.AutoSync || cfg.SyncInterval <= 0 {
		return nil
	}

	interval := time.Duration(cfg.SyncInterval) * time.Minute
	m.ticker = time.NewTicker(interval)
	m.running = true

	go func() {
		for {
			select {
			case <-m.ticker.C:
				result := m.Sync()
				if !result.Success {
					fmt.Printf("[SyncManager] Auto-sync failed: %s\n", result.Error)
					m.mu.Lock()
					m.lastError = fmt.Errorf("%s", result.Error)
					m.mu.Unlock()
				} else if result.Message != "" {
					fmt.Printf("[SyncManager] Auto-sync succeeded: %s\n", result.Message)
				}
			case <-m.stopChan:
				return
			}
		}
	}()

	fmt.Printf("[SyncManager] Auto-sync started with interval %v\n", interval)
	return nil
}

// StopAutoSync stops automatic synchronization.
func (m *SyncManager) StopAutoSync() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.running {
		return
	}

	if m.ticker != nil {
		m.ticker.Stop()
		m.ticker = nil
	}

	close(m.stopChan)
	m.stopChan = make(chan struct{})
	m.running = false

	fmt.Println("[SyncManager] Auto-sync stopped")
}

// GetStatus returns the current synchronization status.
func (m *SyncManager) GetStatus() *SyncStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	cfg := m.config.Get()

	status := &SyncStatus{
		Syncing:     m.syncing,
		LastSyncTime: cfg.LastSyncTime,
		LastSyncHash: cfg.LastSyncHash,
		Enabled:      cfg.Enabled,
		AutoSync:     cfg.AutoSync,
		RemoteURL:    cfg.RepoURL,
	}

	if m.lastError != nil {
		status.Error = m.lastError.Error()
	}

	// Check for changes if repo exists
	if m.git.IsRepo() {
		hasChanges, _ := m.git.HasChanges()
		status.HasChanges = hasChanges
	}

	return status
}

// SetConfig updates the synchronization configuration.
func (m *SyncManager) SetConfig(cfg *SyncConfig) error {
	// Get old config before updating
	oldCfg := m.config.Get()

	// Save to memory and disk
	if err := m.config.Set(cfg); err != nil {
		return err
	}

	if err := m.config.Save(); err != nil {
		return err
	}

	// Check if we need to restart auto-sync
	// Restart is needed if:
	// 1. Auto-sync is currently running AND
	// 2. Either AutoSync flag changed OR SyncInterval changed
	needsRestart := m.running &&
		(oldCfg.AutoSync != cfg.AutoSync ||
			oldCfg.SyncInterval != cfg.SyncInterval)

	if needsRestart {
		fmt.Println("[SyncManager] Configuration changed, restarting auto-sync")
		m.StopAutoSync()
	}

	// Start auto-sync if enabled
	if cfg.AutoSync && cfg.Enabled {
		m.StartAutoSync()
	}

	return nil
}

// GetConfig returns the current configuration.
func (m *SyncManager) GetConfig() *SyncConfig {
	return m.config.Get()
}

// TestConnection tests the connection to a repository.
func (m *SyncManager) TestConnection(url string) (*ConnectionTestResult, error) {
	return m.git.TestConnection(url)
}

// StoreToken stores a Git access token securely.
func (m *SyncManager) StoreToken(token string) error {
	return m.keychain.Store(KeychainTokenKey, token)
}

// GetToken retrieves the stored Git access token.
func (m *SyncManager) GetToken() (string, error) {
	return m.keychain.Retrieve(KeychainTokenKey)
}

// HasToken checks if a token is stored.
func (m *SyncManager) HasToken() bool {
	return m.keychain.Exists(KeychainTokenKey)
}

// DeleteToken removes the stored token.
func (m *SyncManager) DeleteToken() error {
	return m.keychain.Delete(KeychainTokenKey)
}

// IsGitInstalled checks if Git is available on the system.
func (m *SyncManager) IsGitInstalled() bool {
	return m.git.IsGitInstalled()
}

// CheckSSHCredential checks if SSH credentials are available.
func (m *SyncManager) CheckSSHCredential() bool {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return false
	}

	// Check for common SSH key files
	sshDir := filepath.Join(homeDir, ".ssh")
	keyFiles := []string{"id_rsa", "id_ed25519", "id_ecdsa", "id_dsa"}

	for _, keyFile := range keyFiles {
		keyPath := filepath.Join(sshDir, keyFile)

		// Check if private key exists
		info, err := os.Stat(keyPath)
		if err != nil {
			continue
		}

		// Check if file is a regular file (not directory, symlink is ok)
		if !info.Mode().IsRegular() {
			continue
		}

		// Check file permissions (should be 0600 or 0400 for SSH keys)
		// SSH requires strict permissions on private keys
		perms := info.Mode().Perm()
		if perms&0077 != 0 {
			// Group or others have permissions, this might be rejected by SSH
			// But we still return true as the key exists
			fmt.Printf("[SyncManager] Warning: SSH key %s has loose permissions: %o\n", keyFile, perms)
		}

		// Check if corresponding public key exists (optional but good practice)
		pubKeyPath := keyPath + ".pub"
		if _, err := os.Stat(pubKeyPath); err == nil {
			// Both private and public key exist, good setup
			return true
		}

		// Only private key exists, still valid
		return true
	}

	return false
}

// AddIgnorePattern adds a custom ignore pattern.
func (m *SyncManager) AddIgnorePattern(pattern string) {
	m.ignore.AddPattern(pattern)
}

// AddIgnorePatterns adds multiple custom ignore patterns.
func (m *SyncManager) AddIgnorePatterns(patterns []string) {
	m.ignore.AddPatterns(patterns)
}

// GetIgnorePatterns returns all ignore patterns.
func (m *SyncManager) GetIgnorePatterns() []string {
	return m.ignore.GetPatterns()
}

// IsSSHURL checks if a URL is an SSH URL.
func IsSSHURL(url string) bool {
	return strings.HasPrefix(url, "git@") || strings.HasPrefix(url, "ssh://")
}

// IsHTTPSURL checks if a URL is an HTTPS URL.
func IsHTTPSURL(url string) bool {
	return strings.HasPrefix(url, "https://") || strings.HasPrefix(url, "http://")
}
