package vault

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
)

const (
	configFileName = "vault.json"
	configVersion  = 1
)

var (
	ErrVaultNotInitialized = errors.New("vault not initialized")
	ErrVaultAlreadySetup   = errors.New("vault already setup")
)

// Storage 管理保险库数据的持久化
type Storage struct {
	mu       sync.RWMutex
	filePath string
	config   *VaultConfig
}

// NewStorage 创建新的存储实例
func NewStorage(dataDir string) *Storage {
	return &Storage{
		filePath: filepath.Join(dataDir, configFileName),
	}
}

// IsInitialized 检查保险库是否已初始化
func (s *Storage) IsInitialized() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.config != nil {
		return true
	}

	// 检查文件是否存在
	_, err := os.Stat(s.filePath)
	return err == nil
}

// Load 加载保险库配置
func (s *Storage) Load() (*VaultConfig, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrVaultNotInitialized
		}
		return nil, err
	}

	var config VaultConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	s.config = &config
	return &config, nil
}

// Save 保存保险库配置
func (s *Storage) Save(config *VaultConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 确保目录存在
	dir := filepath.Dir(s.filePath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	// 使用限制权限写入文件
	if err := os.WriteFile(s.filePath, data, 0600); err != nil {
		return err
	}

	s.config = config
	return nil
}

// Initialize 初始化新的保险库
func (s *Storage) Initialize(masterPassword string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 检查是否已存在
	if _, err := os.Stat(s.filePath); err == nil {
		return ErrVaultAlreadySetup
	}

	// 生成 salt
	salt, err := GenerateSalt()
	if err != nil {
		return err
	}

	// 创建验证哈希
	verificationHash := HashPassword(masterPassword, salt)

	// 创建默认配置
	config := &VaultConfig{
		Version:          configVersion,
		VerificationHash: verificationHash,
		Salt:             salt,
		Entries:          []VaultEntry{},
		Categories:       []string{"社交", "工作", "金融", "购物", "其他"},
	}

	// 确保目录存在
	dir := filepath.Dir(s.filePath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	// 保存配置
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(s.filePath, data, 0600); err != nil {
		return err
	}

	s.config = config
	return nil
}

// GetConfig 获取当前配置（内存中的副本）
func (s *Storage) GetConfig() *VaultConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.config == nil {
		return nil
	}

	// 返回副本
	config := *s.config
	return &config
}

// SetConfig 更新内存中的配置
func (s *Storage) SetConfig(config *VaultConfig) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.config = config
}

// Clear 清除内存中的配置（锁定）
func (s *Storage) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.config = nil
}
