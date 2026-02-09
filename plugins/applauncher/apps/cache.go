package apps

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	// CacheFileName 缓存文件名
	CacheFileName = "apps_cache.json"
	// CacheVersion 缓存格式版本
	CacheVersion = "1.0"
	// DefaultCacheDuration 默认缓存过期时间（24小时）
	DefaultCacheDuration = 24 * time.Hour
)

// Cache 应用缓存管理
type Cache struct {
	dataDir string
	path    string
}

// NewCache 创建新的缓存实例
func NewCache(dataDir string) (*Cache, error) {
	cacheDir := filepath.Join(dataDir, "applauncher")

	// 确保缓存目录存在
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	return &Cache{
		dataDir: cacheDir,
		path:    filepath.Join(cacheDir, CacheFileName),
	}, nil
}

// Load 从缓存加载应用列表
func (c *Cache) Load() ([]*AppInfo, error) {
	data, err := os.ReadFile(c.path)
	if err != nil {
		if os.IsNotExist(err) {
			return []*AppInfo{}, nil // 缓存不存在，返回空列表
		}
		return nil, fmt.Errorf("failed to read cache file: %w", err)
	}

	var cacheFile struct {
		Info CacheInfo  `json:"info"`
		Apps []*AppInfo `json:"apps"`
	}

	if err := json.Unmarshal(data, &cacheFile); err != nil {
		return nil, fmt.Errorf("failed to parse cache file: %w", err)
	}

	// 检查版本兼容性
	if cacheFile.Info.Version != CacheVersion {
		return []*AppInfo{}, nil // 版本不匹配，忽略缓存
	}

	return cacheFile.Apps, nil
}

// Save 保存应用列表到缓存
func (c *Cache) Save(apps []*AppInfo) error {
	cacheFile := struct {
		Info CacheInfo  `json:"info"`
		Apps []*AppInfo `json:"apps"`
	}{
		Info: CacheInfo{
			LastUpdated: time.Now(),
			AppCount:    len(apps),
			Version:     CacheVersion,
		},
		Apps: apps,
	}

	// 使用 Encoder 来避免 Unicode 转义
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetEscapeHTML(false) // 不转义 Unicode 字符
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(cacheFile); err != nil {
		return fmt.Errorf("failed to marshal cache data: %w", err)
	}

	data := buf.Bytes()

	// 原子写入：先写到临时文件，再重命名
	tmpPath := c.path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}

	if err := os.Rename(tmpPath, c.path); err != nil {
		os.Remove(tmpPath) // 清理临时文件
		return fmt.Errorf("failed to rename cache file: %w", err)
	}

	return nil
}

// IsExpired 检查缓存是否过期
func (c *Cache) IsExpired(duration time.Duration) bool {
	info, err := c.getInfo()
	if err != nil {
		return true // 读取失败视为过期
	}

	return time.Since(info.LastUpdated) > duration
}

// getInfo 获取缓存信息
func (c *Cache) getInfo() (*CacheInfo, error) {
	data, err := os.ReadFile(c.path)
	if err != nil {
		return nil, err
	}

	var cacheFile struct {
		Info CacheInfo `json:"info"`
	}

	if err := json.Unmarshal(data, &cacheFile); err != nil {
		return nil, err
	}

	return &cacheFile.Info, nil
}

// Status 获取缓存状态
func (c *Cache) Status() map[string]interface{} {
	info, err := c.getInfo()
	if err != nil {
		return map[string]interface{}{
			"exists":     false,
			"expired":    true,
			"appCount":   0,
			"lastUpdate": "never",
		}
	}

	return map[string]interface{}{
		"exists":     true,
		"expired":    c.IsExpired(DefaultCacheDuration),
		"appCount":   info.AppCount,
		"lastUpdate": info.LastUpdated.Format("2006-01-02 15:04:05"),
		"version":    info.Version,
	}
}

// Clear 清除缓存
func (c *Cache) Clear() error {
	if err := os.Remove(c.path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to clear cache: %w", err)
	}
	return nil
}
