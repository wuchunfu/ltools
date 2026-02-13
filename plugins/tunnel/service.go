package tunnel

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// TunnelService 内网穿透服务
type TunnelService struct {
	plugin  *TunnelPlugin
	app     *application.App
	dataDir string
}

// NewTunnelService 创建内网穿透服务
func NewTunnelService(plugin *TunnelPlugin, app *application.App, dataDir string) *TunnelService {
	return &TunnelService{
		plugin:  plugin,
		app:     app,
		dataDir: dataDir,
	}
}

// SetApp 设置应用实例
func (s *TunnelService) SetApp(app *application.App) {
	s.app = app
}

// GetInstallationStatus 获取安装状态
func (s *TunnelService) GetInstallationStatus() *InstallationStatus {
	status := &InstallationStatus{
		FRPInstalled: false,
	}

	// 检查 FRP 是否已安装
	if s.plugin.frpInstaller != nil {
		if path, err := s.plugin.frpInstaller.CheckInstallation(); err == nil {
			status.FRPInstalled = true
			status.FRPVersion = path
		}
	}

	return status
}

// InstallFRP 安装 FRP
func (s *TunnelService) InstallFRP() (*OperationResult, error) {
	if s.plugin.frpInstaller == nil {
		return &OperationResult{
			Success: false,
			Error:   "FRP 安装器未初始化",
		}, nil
	}

	err := s.plugin.frpInstaller.Install(func(msg string) {
		s.plugin.emitEvent("install:progress", msg)
	})

	if err != nil {
		return &OperationResult{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &OperationResult{Success: true}, nil
}

// GetTunnels 获取所有隧道配置
func (s *TunnelService) GetTunnels() []Tunnel {
	if s.plugin.configMgr == nil || s.plugin.configMgr.Config == nil {
		return []Tunnel{}
	}
	return s.plugin.configMgr.Config.Tunnels
}

// CreateTunnel 创建隧道
func (s *TunnelService) CreateTunnel(req *CreateTunnelRequest) (*OperationResult, error) {
	// 默认使用 FRP 协议
	if req.Protocol == "" {
		req.Protocol = ProtocolFRP
	}

	tunnel := Tunnel{
		ID:        generateTunnelID(req.Name),
		Name:      req.Name,
		Protocol:  req.Protocol,
		LocalHost: req.LocalHost,
		LocalPort: req.LocalPort,
		Enabled:   true,
		AutoStart: req.AutoStart,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 设置默认代理类型
	if tunnel.Protocol == ProtocolFRP && tunnel.ProxyType == "" {
		tunnel.ProxyType = ProxyTypeHTTP
	}

	s.plugin.configMgr.Config.Tunnels = append(s.plugin.configMgr.Config.Tunnels, tunnel)

	if err := s.plugin.configMgr.SaveConfig(); err != nil {
		return &OperationResult{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	s.plugin.emitEvent("created", tunnel.ID)

	return &OperationResult{Success: true}, nil
}

// UpdateTunnel 更新隧道
func (s *TunnelService) UpdateTunnel(id string, req *UpdateTunnelRequest) (*OperationResult, error) {
	// 查找隧道
	for i := range s.plugin.configMgr.Config.Tunnels {
		if s.plugin.configMgr.Config.Tunnels[i].ID == id {
			tunnel := &s.plugin.configMgr.Config.Tunnels[i]

			// 更新通用字段
			if req.Name != "" {
				tunnel.Name = req.Name
			}
			if req.LocalHost != "" {
				tunnel.LocalHost = req.LocalHost
			}
			if req.LocalPort > 0 {
				tunnel.LocalPort = req.LocalPort
			}

			// 更新 FRP 专用字段
			if req.ProxyType != "" {
				tunnel.ProxyType = req.ProxyType
			}
			if req.FRPServer != nil {
				tunnel.FRPServer = req.FRPServer
			}
			// 更新子域名（允许清空）
			tunnel.Subdomain = req.Subdomain

			tunnel.UpdatedAt = time.Now()
			break
		}
	}

	if err := s.plugin.configMgr.SaveConfig(); err != nil {
		return &OperationResult{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	s.plugin.emitEvent("updated", id)

	return &OperationResult{Success: true}, nil
}

// DeleteTunnel 删除隧道
func (s *TunnelService) DeleteTunnel(id string) (*OperationResult, error) {
	// 停止隧道（如果正在运行）
	if s.plugin.frpMgr != nil {
		s.plugin.frpMgr.StopTunnel(id)
	}

	// 查找并删除
	for i, tunnel := range s.plugin.configMgr.Config.Tunnels {
		if tunnel.ID == id {
			s.plugin.configMgr.Config.Tunnels = append(
				s.plugin.configMgr.Config.Tunnels[:i],
				s.plugin.configMgr.Config.Tunnels[i+1:]...)
			break
		}
	}

	if err := s.plugin.configMgr.SaveConfig(); err != nil {
		return &OperationResult{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	s.plugin.emitEvent("deleted", id)

	return &OperationResult{Success: true}, nil
}

// StartTunnel 启动隧道
func (s *TunnelService) StartTunnel(id string) (*OperationResult, error) {
	// 查找隧道
	var tunnelConfig *Tunnel
	for i := range s.plugin.configMgr.Config.Tunnels {
		if s.plugin.configMgr.Config.Tunnels[i].ID == id {
			tunnelConfig = &s.plugin.configMgr.Config.Tunnels[i]
			break
		}
	}
	if tunnelConfig == nil {
		return &OperationResult{
			Success: false,
			Error:   "隧道不存在",
		}, nil
	}

	// 使用 FRP 启动
	if s.plugin.frpMgr == nil {
		return &OperationResult{
			Success: false,
			Error:   "FRP 管理器未初始化",
		}, nil
	}

	return s.plugin.frpMgr.StartTunnel(tunnelConfig)
}

// StopTunnel 停止隧道
func (s *TunnelService) StopTunnel(id string) (*OperationResult, error) {
	if s.plugin.frpMgr != nil {
		s.plugin.frpMgr.StopTunnel(id)
	}

	return &OperationResult{Success: true}, nil
}

// RestartTunnel 重启隧道
func (s *TunnelService) RestartTunnel(id string) (*OperationResult, error) {
	// 查找隧道
	var tunnelConfig *Tunnel
	for i := range s.plugin.configMgr.Config.Tunnels {
		if s.plugin.configMgr.Config.Tunnels[i].ID == id {
			tunnelConfig = &s.plugin.configMgr.Config.Tunnels[i]
			break
		}
	}
	if tunnelConfig == nil {
		return &OperationResult{
			Success: false,
			Error:   "隧道不存在",
		}, nil
	}

	if s.plugin.frpMgr == nil {
		return &OperationResult{
			Success: false,
			Error:   "FRP 管理器未初始化",
		}, nil
	}

	return s.plugin.frpMgr.RestartTunnel(tunnelConfig)
}

// GetAllTunnelStatuses 获取所有隧道状态
func (s *TunnelService) GetAllTunnelStatuses() []TunnelRuntimeInfo {
	if s.plugin.frpMgr == nil {
		return []TunnelRuntimeInfo{}
	}

	return s.plugin.frpMgr.GetAllStatuses()
}

// GetGlobalOptions 获取全局配置
func (s *TunnelService) GetGlobalOptions() *GlobalOptions {
	if s.plugin.configMgr == nil || s.plugin.configMgr.Config == nil {
		return &GlobalOptions{}
	}
	return &s.plugin.configMgr.Config.GlobalOptions
}

// SetGlobalOptions 设置全局配置
func (s *TunnelService) SetGlobalOptions(opts *GlobalOptions) (*OperationResult, error) {
	s.plugin.configMgr.Config.GlobalOptions = *opts

	if err := s.plugin.configMgr.SaveConfig(); err != nil {
		return &OperationResult{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	s.plugin.emitEvent("options:updated", nil)

	return &OperationResult{Success: true}, nil
}

// GetTunnelLog 获取隧道日志内容
func (s *TunnelService) GetTunnelLog(tunnelID string, lines int) (string, error) {
	if s.plugin.frpMgr == nil {
		return "", fmt.Errorf("FRP 管理器未初始化")
	}

	// 获取隧道状态
	status := s.plugin.frpMgr.GetTunnelStatus(tunnelID)
	if status == nil || status.LogPath == "" {
		return "", fmt.Errorf("暂无日志文件")
	}

	// 读取日志文件
	content, err := os.ReadFile(status.LogPath)
	if err != nil {
		return "", fmt.Errorf("读取日志失败: %v", err)
	}

	logContent := string(content)

	// 如果指定了行数限制，返回最后 N 行
	if lines > 0 {
		logLines := strings.Split(logContent, "\n")
		if len(logLines) > lines {
			logLines = logLines[len(logLines)-lines:]
		}
		logContent = strings.Join(logLines, "\n")
	}

	return logContent, nil
}

// generateTunnelID 生成隧道 ID
func generateTunnelID(name string) string {
	return fmt.Sprintf("%s-%d", name, time.Now().Unix())
}
