package tunnel

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const (
	configFileName = "tunnel.json"
	configVersion  = 2
)

// ConfigManager 统一的配置管理器
type ConfigManager struct {
	configDir string
	Config    *TunnelConfig
}

// NewConfigManager 创建配置管理器
func NewConfigManager(dataDir string) *ConfigManager {
	return &ConfigManager{
		configDir: dataDir,
	}
}

// GetConfigPath 获取配置文件路径
func (cm *ConfigManager) GetConfigPath() string {
	return filepath.Join(cm.configDir, configFileName)
}

// LoadConfig 加载配置文件
func (cm *ConfigManager) LoadConfig() (*TunnelConfig, error) {
	configPath := cm.GetConfigPath()

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// 返回默认配置
			config := DefaultConfig()
			cm.Config = config
			return config, nil
		}
		return nil, err
	}

	var config TunnelConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	// 版本检查和迁移
	if config.Version < configVersion {
		cm.migrateConfig(&config)
	}

	cm.Config = &config
	return &config, nil
}

// SaveConfig 保存配置文件
func (cm *ConfigManager) SaveConfig() error {
	if cm.Config == nil {
		return nil
	}

	configPath := cm.GetConfigPath()

	// 确保目录存在
	if err := os.MkdirAll(cm.configDir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cm.Config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

// migrateConfig 配置迁移
func (cm *ConfigManager) migrateConfig(config *TunnelConfig) {
	// v1 -> v2: 添加 FRP 支持和协议选择
	config.Version = configVersion

	// 迁移现有隧道：为空协议的设置为 FRP
	for i := range config.Tunnels {
		if config.Tunnels[i].Protocol == "" {
			config.Tunnels[i].Protocol = ProtocolFRP
		}
	}
}

// DefaultConfig 创建默认配置
func DefaultConfig() *TunnelConfig {
	return &TunnelConfig{
		Version: configVersion,
		GlobalOptions: GlobalOptions{
			DefaultProtocol: ProtocolFRP,
			FRPServer: &FRPServerConfig{
				Address: "127.0.0.1:7000",
				Token:   "",
			},
		},
		Tunnels: []Tunnel{},
	}
}
