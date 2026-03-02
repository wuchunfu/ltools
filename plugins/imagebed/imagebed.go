package imagebed

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"
	"ltools/internal/plugins"
)

const (
	PluginID      = "imagebed.builtin"
	PluginName    = "图床"
	PluginVersion = "1.0.0"
)

// ImageBedPlugin provides image hosting functionality via GitHub + jsDelivr
type ImageBedPlugin struct {
	*plugins.BasePlugin
	app      *application.App
	config   *ImageBedConfig
	history  *UploadHistory
	dataDir  string
	uploader *Uploader
}

// NewImageBedPlugin creates a new image bed plugin
func NewImageBedPlugin() *ImageBedPlugin {
	metadata := &plugins.PluginMetadata{
		ID:          PluginID,
		Name:        PluginName,
		Version:     PluginVersion,
		Author:      "LTools Team",
		Description: "使用 GitHub + jsDelivr 搭建个人图床，支持拖拽上传和剪贴板粘贴",
		Icon:        "photo",
		Type:        plugins.PluginTypeBuiltIn,
		State:       plugins.PluginStateInstalled,
		Permissions: []plugins.Permission{
			plugins.PermissionNetwork,
		},
		Keywords:   []string{"图床", "图片", "上传", "github", "jsdelivr", "image", "upload", "hosting"},
		ShowInMenu: plugins.BoolPtr(true),
		HasPage:    plugins.BoolPtr(true),
	}

	return &ImageBedPlugin{
		BasePlugin: plugins.NewBasePlugin(metadata),
		config: &ImageBedConfig{
			Version: 1,
			Path:    "images",
			Branch:  "main",
		},
		history: &UploadHistory{
			Version: 1,
			Records: []UploadRecord{},
		},
	}
}

// Init initializes the plugin
func (p *ImageBedPlugin) Init(app *application.App) error {
	if err := p.BasePlugin.Init(app); err != nil {
		return err
	}
	p.app = app
	return nil
}

// SetDataDir sets the data directory for persistence
func (p *ImageBedPlugin) SetDataDir(dataDir string) error {
	p.dataDir = dataDir
	// Load both config and history
	if err := p.loadConfig(); err != nil {
		return err
	}
	return p.LoadHistory(dataDir)
}

// ServiceStartup is called when the application starts
func (p *ImageBedPlugin) ServiceStartup(app *application.App) error {
	if err := p.BasePlugin.ServiceStartup(app); err != nil {
		return err
	}
	p.app = app
	return nil
}

// ServiceShutdown is called when the application shuts down
func (p *ImageBedPlugin) ServiceShutdown(app *application.App) error {
	return p.BasePlugin.ServiceShutdown(app)
}

// Enabled returns true if the plugin is enabled
func (p *ImageBedPlugin) Enabled() bool {
	return p.BasePlugin.Enabled()
}

// SetEnabled enables or disables the plugin
func (p *ImageBedPlugin) SetEnabled(enabled bool) error {
	return p.BasePlugin.SetEnabled(enabled)
}

// GetConfig returns the current configuration
func (p *ImageBedPlugin) GetConfig() *ImageBedConfig {
	return p.config
}

// SetConfig updates the configuration
func (p *ImageBedPlugin) SetConfig(config *ImageBedConfig) error {
	p.config = config
	p.uploader = NewUploader(config)
	return p.saveConfig()
}

// loadConfig loads configuration from file
func (p *ImageBedPlugin) loadConfig() error {
	if p.dataDir == "" {
		return nil
	}

	configPath := filepath.Join(p.dataDir, "imagebed", "config.json")
	if data, err := os.ReadFile(configPath); err == nil {
		if err := json.Unmarshal(data, p.config); err != nil {
			return err
		}
	}
	return nil
}

// saveConfig saves configuration to file
func (p *ImageBedPlugin) saveConfig() error {
	if p.dataDir == "" {
		return nil
	}

	configDir := filepath.Join(p.dataDir, "imagebed")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	configPath := filepath.Join(configDir, "config.json")
	data, err := json.MarshalIndent(p.config, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(configPath, data, 0644)
}

// UploadImage uploads an image from base64 content
func (p *ImageBedPlugin) UploadImage(fileName string, base64Content string) (*UploadResult, error) {
	if p.config.GitHubToken == "" {
		return &UploadResult{
			Success: false,
			Message: "GitHub Token 未配置，请先配置图床设置",
		}, nil
	}

	// Remove data URI prefix if present
	content := base64Content
	if idx := strings.Index(base64Content, ","); idx != -1 {
		content = base64Content[idx+1:]
	}

	// Decode base64
	data, err := base64.StdEncoding.DecodeString(content)
	if err != nil {
		return &UploadResult{
			Success: false,
			Message: fmt.Sprintf("解码图片失败: %v", err),
		}, nil
	}

	// Validate image size (max 10MB)
	if len(data) > 10*1024*1024 {
		return &UploadResult{
			Success: false,
			Message: "图片大小超过 10MB 限制",
		}, nil
	}

	// Sanitize filename
	fileName = SanitizeFileName(fileName)

	// Initialize uploader if needed
	if p.uploader == nil {
		p.uploader = NewUploader(p.config)
	}

	// Upload to GitHub
	record, err := p.uploader.Upload(fileName, data)
	if err != nil {
		return &UploadResult{
			Success: false,
			Message: fmt.Sprintf("上传失败: %v", err),
		}, nil
	}

	// Add to history
	if err := p.AddRecord(record, p.dataDir); err != nil {
		fmt.Printf("[ImageBed] Failed to save history: %v\n", err)
	}

	// Emit event
	if p.app != nil {
		p.app.Event.Emit("imagebed:uploaded", record.ID)
	}

	return &UploadResult{
		Success: true,
		Message: "上传成功",
		Record:  record,
	}, nil
}

// DeleteImage deletes an image from GitHub and history
func (p *ImageBedPlugin) DeleteImage(id string) (*UploadResult, error) {
	record := p.GetRecordByID(id)
	if record == nil {
		return &UploadResult{
			Success: false,
			Message: "记录不存在",
		}, nil
	}

	// Delete from GitHub if we have the sha
	if record.Sha != "" && p.config.GitHubToken != "" {
		if p.uploader == nil {
			p.uploader = NewUploader(p.config)
		}

		if err := p.uploader.Delete(record.Path, record.Sha); err != nil {
			// Log error but still remove from history
			fmt.Printf("[ImageBed] Failed to delete from GitHub: %v\n", err)
		}
	}

	// Remove from history
	if err := p.DeleteRecord(id, p.dataDir); err != nil {
		return &UploadResult{
			Success: false,
			Message: fmt.Sprintf("删除历史记录失败: %v", err),
		}, nil
	}

	// Emit event
	if p.app != nil {
		p.app.Event.Emit("imagebed:deleted", id)
	}

	return &UploadResult{
		Success: true,
		Message: "删除成功",
	}, nil
}

// ValidateConfig validates the current configuration
func (p *ImageBedPlugin) ValidateConfig() (*ConfigValidationResult, error) {
	if p.uploader == nil {
		p.uploader = NewUploader(p.config)
	}
	return p.uploader.ValidateConfig()
}

// RenameImage renames an image file in GitHub repository
func (p *ImageBedPlugin) RenameImage(id string, newFileName string) (*UploadResult, error) {
	record := p.GetRecordByID(id)
	if record == nil {
		return &UploadResult{
			Success: false,
			Message: "记录不存在",
		}, nil
	}

	if p.config.GitHubToken == "" {
		return &UploadResult{
			Success: false,
			Message: "GitHub Token 未配置",
		}, nil
	}

	// Sanitize new filename
	newFileName = SanitizeFileName(newFileName)

	// Initialize uploader if needed
	if p.uploader == nil {
		p.uploader = NewUploader(p.config)
	}

	// Download existing image content
	content, err := p.uploader.DownloadFile(record.Path)
	if err != nil {
		return &UploadResult{
			Success: false,
			Message: fmt.Sprintf("下载原图片失败: %v", err),
		}, nil
	}

	// Build new path with new filename
	newPath := fmt.Sprintf("%s/%s/%s",
		strings.Trim(p.config.Path, "/"),
		time.Now().Format("20060102"),
		newFileName)

	// Upload with new name
	newRecord, err := p.uploader.UploadToPath(newPath, content)
	if err != nil {
		return &UploadResult{
			Success: false,
			Message: fmt.Sprintf("上传新文件失败: %v", err),
		}, nil
	}

	// Delete old file
	if err := p.uploader.Delete(record.Path, record.Sha); err != nil {
		fmt.Printf("[ImageBed] Failed to delete old file: %v\n", err)
		// Don't fail the operation, just log the error
	}

	// Update history record
	if err := p.DeleteRecord(id, p.dataDir); err != nil {
		fmt.Printf("[ImageBed] Failed to delete old record: %v\n", err)
	}

	// Add new record
	if err := p.AddRecord(newRecord, p.dataDir); err != nil {
		fmt.Printf("[ImageBed] Failed to save new record: %v\n", err)
	}

	// Emit event
	if p.app != nil {
		p.app.Event.Emit("imagebed:renamed", id)
	}

	return &UploadResult{
		Success: true,
		Message: "重命名成功",
		Record:  newRecord,
	}, nil
}
