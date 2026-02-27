package vault

import (
	"errors"
	"strings"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"
)

var (
	ErrVaultLocked      = errors.New("vault is locked")
	ErrEntryNotFound    = errors.New("entry not found")
	ErrCategoryExists   = errors.New("category already exists")
	ErrCategoryNotFound = errors.New("category not found")
)

// VaultService 保险库服务
type VaultService struct {
	plugin    *VaultPlugin
	app       *application.App
	storage   *Storage
	encryptionKey []byte // 内存中的加密密钥（解锁后设置）
	isLocked  bool
}

// NewVaultService 创建新的保险库服务
func NewVaultService(plugin *VaultPlugin, app *application.App) *VaultService {
	return &VaultService{
		plugin:   plugin,
		app:      app,
		storage:  plugin.storage,
		isLocked: true,
	}
}

// Status 获取保险库状态
func (s *VaultService) Status() VaultStatus {
	return VaultStatus{
		Initialized: s.storage.IsInitialized(),
		Locked:      s.isLocked,
	}
}

// IsInitialized 检查保险库是否已初始化
func (s *VaultService) IsInitialized() bool {
	return s.storage.IsInitialized()
}

// IsLocked 检查保险库是否锁定
func (s *VaultService) IsLocked() bool {
	return s.isLocked
}

// Setup 初始化保险库（首次设置主密码）
func (s *VaultService) Setup(masterPassword string) error {
	if s.storage.IsInitialized() {
		return ErrVaultAlreadySetup
	}

	if err := s.storage.Initialize(masterPassword); err != nil {
		return err
	}

	// 设置加密密钥
	config, err := s.storage.Load()
	if err != nil {
		return err
	}

	s.encryptionKey = DeriveKey(masterPassword, config.Salt)
	s.isLocked = false

	return nil
}

// Unlock 解锁保险库
func (s *VaultService) Unlock(masterPassword string) error {
	config, err := s.storage.Load()
	if err != nil {
		return err
	}

	// 验证主密码
	if !VerifyPassword(masterPassword, config.Salt, config.VerificationHash) {
		return ErrInvalidPassword
	}

	// 设置加密密钥
	s.encryptionKey = DeriveKey(masterPassword, config.Salt)
	s.isLocked = false

	return nil
}

// Lock 锁定保险库
func (s *VaultService) Lock() error {
	// 清除内存中的敏感数据
	for i := range s.encryptionKey {
		s.encryptionKey[i] = 0
	}
	s.encryptionKey = nil
	s.storage.Clear()
	s.isLocked = true

	return nil
}

// ChangeMasterPassword 修改主密码
func (s *VaultService) ChangeMasterPassword(currentPassword, newPassword string) error {
	if s.isLocked {
		return ErrVaultLocked
	}

	config := s.storage.GetConfig()
	if config == nil {
		return ErrVaultLocked
	}

	// 验证当前主密码
	if !VerifyPassword(currentPassword, config.Salt, config.VerificationHash) {
		return ErrInvalidPassword
	}

	// 生成新的 salt
	newSalt, err := GenerateSalt()
	if err != nil {
		return err
	}

	// 使用新密码派生新密钥
	newKey := DeriveKey(newPassword, newSalt)

	// 重新加密所有密码条目
	newEntries := make([]VaultEntry, len(config.Entries))
	for i, entry := range config.Entries {
		// 解密原密码
		decrypted, err := Decrypt(entry.Password, s.encryptionKey)
		if err != nil {
			return err
		}

		// 使用新密钥加密
		encrypted, err := Encrypt(decrypted, newKey)
		if err != nil {
			return err
		}

		newEntries[i] = entry
		newEntries[i].Password = encrypted
	}

	// 更新配置
	config.Salt = newSalt
	config.VerificationHash = HashPassword(newPassword, newSalt)
	config.Entries = newEntries

	// 保存配置
	if err := s.storage.Save(config); err != nil {
		return err
	}

	// 更新内存中的密钥
	for i := range s.encryptionKey {
		s.encryptionKey[i] = 0
	}
	s.encryptionKey = newKey

	return nil
}

// ListEntries 获取所有条目（密码字段已解密）
func (s *VaultService) ListEntries() ([]VaultEntry, error) {
	if s.isLocked {
		return nil, ErrVaultLocked
	}

	config := s.storage.GetConfig()
	if config == nil {
		return nil, ErrVaultLocked
	}

	// 解密所有密码
	entries := make([]VaultEntry, len(config.Entries))
	for i, entry := range config.Entries {
		entries[i] = entry
		if entry.Password != "" {
			decrypted, err := Decrypt(entry.Password, s.encryptionKey)
			if err != nil {
				// 解密失败，返回原始值（可能是损坏的数据）
				entries[i].Password = entry.Password
			} else {
				entries[i].Password = string(decrypted)
			}
		}
	}

	return entries, nil
}

// GetEntry 获取单个条目
func (s *VaultService) GetEntry(id string) (*VaultEntry, error) {
	if s.isLocked {
		return nil, ErrVaultLocked
	}

	config := s.storage.GetConfig()
	if config == nil {
		return nil, ErrVaultLocked
	}

	for _, entry := range config.Entries {
		if entry.ID == id {
			result := entry
			if entry.Password != "" {
				decrypted, err := Decrypt(entry.Password, s.encryptionKey)
				if err != nil {
					return nil, err
				}
				result.Password = string(decrypted)
			}
			return &result, nil
		}
	}

	return nil, ErrEntryNotFound
}

// CreateEntry 创建新条目
func (s *VaultService) CreateEntry(req CreateEntryRequest) (*VaultEntry, error) {
	if s.isLocked {
		return nil, ErrVaultLocked
	}

	config := s.storage.GetConfig()
	if config == nil {
		return nil, ErrVaultLocked
	}

	// 生成 ID
	id, err := GenerateRandomID()
	if err != nil {
		return nil, err
	}

	// 加密密码
	encryptedPassword, err := Encrypt([]byte(req.Password), s.encryptionKey)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	entry := VaultEntry{
		ID:        id,
		Title:     req.Title,
		Website:   req.Website,
		Username:  req.Username,
		Password:  encryptedPassword,
		Notes:     req.Notes,
		Category:  req.Category,
		Tags:      req.Tags,
		Favorite:  req.Favorite,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// 添加到配置
	config.Entries = append(config.Entries, entry)

	// 保存
	if err := s.storage.Save(config); err != nil {
		return nil, err
	}

	// 返回解密后的条目
	result := entry
	result.Password = req.Password
	return &result, nil
}

// UpdateEntry 更新条目
func (s *VaultService) UpdateEntry(req UpdateEntryRequest) (*VaultEntry, error) {
	if s.isLocked {
		return nil, ErrVaultLocked
	}

	config := s.storage.GetConfig()
	if config == nil {
		return nil, ErrVaultLocked
	}

	// 查找条目
	var foundIndex = -1
	for i, entry := range config.Entries {
		if entry.ID == req.ID {
			foundIndex = i
			break
		}
	}

	if foundIndex == -1 {
		return nil, ErrEntryNotFound
	}

	entry := config.Entries[foundIndex]

	// 更新字段
	if req.Title != "" {
		entry.Title = req.Title
	}
	if req.Website != "" {
		entry.Website = req.Website
	}
	if req.Username != "" {
		entry.Username = req.Username
	}
	if req.Password != "" {
		encryptedPassword, err := Encrypt([]byte(req.Password), s.encryptionKey)
		if err != nil {
			return nil, err
		}
		entry.Password = encryptedPassword
	}
	if req.Notes != "" {
		entry.Notes = req.Notes
	}
	if req.Category != "" {
		entry.Category = req.Category
	}
	if req.Tags != nil {
		entry.Tags = req.Tags
	}
	if req.Favorite != nil {
		entry.Favorite = *req.Favorite
	}
	entry.UpdatedAt = time.Now()

	// 更新配置
	config.Entries[foundIndex] = entry

	// 保存
	if err := s.storage.Save(config); err != nil {
		return nil, err
	}

	// 返回解密后的条目
	result := entry
	if req.Password != "" {
		result.Password = req.Password
	} else {
		decrypted, err := Decrypt(entry.Password, s.encryptionKey)
		if err == nil {
			result.Password = string(decrypted)
		}
	}

	return &result, nil
}

// DeleteEntry 删除条目
func (s *VaultService) DeleteEntry(id string) error {
	if s.isLocked {
		return ErrVaultLocked
	}

	config := s.storage.GetConfig()
	if config == nil {
		return ErrVaultLocked
	}

	// 查找并删除
	var foundIndex = -1
	for i, entry := range config.Entries {
		if entry.ID == id {
			foundIndex = i
			break
		}
	}

	if foundIndex == -1 {
		return ErrEntryNotFound
	}

	// 删除条目
	config.Entries = append(config.Entries[:foundIndex], config.Entries[foundIndex+1:]...)

	// 保存
	return s.storage.Save(config)
}

// GetCategories 获取所有分类
func (s *VaultService) GetCategories() ([]string, error) {
	if s.isLocked {
		return nil, ErrVaultLocked
	}

	config := s.storage.GetConfig()
	if config == nil {
		return nil, ErrVaultLocked
	}

	return config.Categories, nil
}

// AddCategory 添加分类
func (s *VaultService) AddCategory(name string) error {
	if s.isLocked {
		return ErrVaultLocked
	}

	config := s.storage.GetConfig()
	if config == nil {
		return ErrVaultLocked
	}

	// 检查是否已存在
	for _, cat := range config.Categories {
		if cat == name {
			return ErrCategoryExists
		}
	}

	config.Categories = append(config.Categories, name)
	return s.storage.Save(config)
}

// DeleteCategory 删除分类
func (s *VaultService) DeleteCategory(name string) error {
	if s.isLocked {
		return ErrVaultLocked
	}

	config := s.storage.GetConfig()
	if config == nil {
		return ErrVaultLocked
	}

	// 查找并删除
	var foundIndex = -1
	for i, cat := range config.Categories {
		if cat == name {
			foundIndex = i
			break
		}
	}

	if foundIndex == -1 {
		return ErrCategoryNotFound
	}

	// 删除分类
	config.Categories = append(config.Categories[:foundIndex], config.Categories[foundIndex+1:]...)

	// 清除使用此分类的条目的分类字段
	for i, entry := range config.Entries {
		if entry.Category == name {
			config.Entries[i].Category = ""
		}
	}

	return s.storage.Save(config)
}

// SearchEntries 搜索条目
func (s *VaultService) SearchEntries(query string) (*SearchResult, error) {
	if s.isLocked {
		return nil, ErrVaultLocked
	}

	config := s.storage.GetConfig()
	if config == nil {
		return nil, ErrVaultLocked
	}

	query = strings.ToLower(query)
	var results []VaultEntry

	for _, entry := range config.Entries {
		// 搜索标题、网站、用户名、备注
		if strings.Contains(strings.ToLower(entry.Title), query) ||
			strings.Contains(strings.ToLower(entry.Website), query) ||
			strings.Contains(strings.ToLower(entry.Username), query) ||
			strings.Contains(strings.ToLower(entry.Notes), query) {
			// 解密密码
			result := entry
			if entry.Password != "" {
				decrypted, err := Decrypt(entry.Password, s.encryptionKey)
				if err == nil {
					result.Password = string(decrypted)
				}
			}
			results = append(results, result)
		}
	}

	return &SearchResult{
		Entries: results,
		Total:   len(results),
	}, nil
}
