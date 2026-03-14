package musicplayer

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// ConfigManager 配置管理器
type ConfigManager struct {
	configDir string
	config    *Config
}

// NewConfigManager 创建配置管理器
func NewConfigManager() (*ConfigManager, error) {
	// 获取用户主目录
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	// 配置目录路径
	configDir := filepath.Join(homeDir, ".ltools", "musicplayer")

	// 确保目录存在
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, err
	}

	cm := &ConfigManager{
		configDir: configDir,
	}

	// 加载配置
	config, err := cm.loadConfig()
	if err != nil {
		// 如果加载失败，使用默认配置
		config = &Config{
			Platform: "netease", // 默认网易云
			Volume:   80,        // 默认音量 80%
		}
	}

	cm.config = config
	return cm, nil
}

// loadConfig 从文件加载配置
func (cm *ConfigManager) loadConfig() (*Config, error) {
	configFile := filepath.Join(cm.configDir, "config.json")

	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// saveConfig 保存配置到文件
func (cm *ConfigManager) saveConfig() error {
	configFile := filepath.Join(cm.configDir, "config.json")

	data, err := json.MarshalIndent(cm.config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configFile, data, 0644)
}

// GetConfig 获取当前配置
func (cm *ConfigManager) GetConfig() *Config {
	return cm.config
}

// SetPlatform 设置当前平台
func (cm *ConfigManager) SetPlatform(platform string) error {
	cm.config.Platform = platform
	return cm.saveConfig()
}

// SetVolume 设置音量
func (cm *ConfigManager) SetVolume(volume int) error {
	if volume < 0 {
		volume = 0
	} else if volume > 100 {
		volume = 100
	}
	cm.config.Volume = volume
	return cm.saveConfig()
}

// MigrateLikeList 迁移旧格式的喜欢列表
func (cm *ConfigManager) MigrateLikeList() error {
	likesFile := filepath.Join(cm.configDir, "likes.json")

	data, err := os.ReadFile(likesFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // 文件不存在，无需迁移
		}
		return err
	}

	// 尝试解析新格式
	var newFormat LikeList
	if err := json.Unmarshal(data, &newFormat); err == nil {
		// 检查是否已经是新格式（通过检查第一个元素的 LikedAt）
		if len(newFormat.Songs) > 0 && !newFormat.Songs[0].LikedAt.IsZero() {
			return nil // 已经是新格式
		}
		// 如果是空列表,也算新格式
		if len(newFormat.Songs) == 0 {
			return nil
		}
	}

	// 尝试解析为旧格式
	var oldFormat struct {
		Songs     []Song    `json:"songs"`
		UpdatedAt time.Time `json:"updated_at"`
	}
	if err := json.Unmarshal(data, &oldFormat); err != nil {
		return err
	}

	// 转换为新格式
	now := time.Now()
	newFormat = LikeList{
		Songs:     make([]SongWithTimestamp, len(oldFormat.Songs)),
		UpdatedAt: oldFormat.UpdatedAt,
	}

	for i, song := range oldFormat.Songs {
		newFormat.Songs[i] = SongWithTimestamp{
			Song:    song,
			LikedAt: now, // 使用当前时间作为默认喜欢时间
		}
	}

	// 保存新格式
	return cm.SaveLikeList(&newFormat)
}

// GetLikeList 获取喜欢列表
func (cm *ConfigManager) GetLikeList() (*LikeList, error) {
	// 先尝试迁移旧格式数据
	cm.MigrateLikeList()

	likesFile := filepath.Join(cm.configDir, "likes.json")

	data, err := os.ReadFile(likesFile)
	if err != nil {
		// 文件不存在，返回空列表
		if os.IsNotExist(err) {
			return &LikeList{
				Songs: []SongWithTimestamp{},
			}, nil
		}
		return nil, err
	}

	var likeList LikeList
	if err := json.Unmarshal(data, &likeList); err != nil {
		return nil, err
	}

	return &likeList, nil
}

// SaveLikeList 保存喜欢列表
func (cm *ConfigManager) SaveLikeList(likeList *LikeList) error {
	likesFile := filepath.Join(cm.configDir, "likes.json")

	data, err := json.MarshalIndent(likeList, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(likesFile, data, 0644)
}
