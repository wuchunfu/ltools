package imagebed

import (
	"github.com/wailsapp/wails/v3/pkg/application"
)

// ImageBedService exposes image bed functionality to the frontend
type ImageBedService struct {
	plugin *ImageBedPlugin
	app    *application.App
}

// NewImageBedService creates a new image bed service
func NewImageBedService(plugin *ImageBedPlugin, app *application.App) *ImageBedService {
	return &ImageBedService{
		plugin: plugin,
		app:    app,
	}
}

// ServiceStartup is called when the application starts
func (s *ImageBedService) ServiceStartup(app *application.App) error {
	return s.plugin.ServiceStartup(app)
}

// ServiceShutdown is called when the application shuts down
func (s *ImageBedService) ServiceShutdown(app *application.App) error {
	return s.plugin.ServiceShutdown(app)
}

// GetConfig returns the current configuration
func (s *ImageBedService) GetConfig() *ImageBedConfig {
	return s.plugin.GetConfig()
}

// SetConfig updates the configuration
func (s *ImageBedService) SetConfig(config *ImageBedConfig) error {
	return s.plugin.SetConfig(config)
}

// ValidateConfig validates the current configuration
func (s *ImageBedService) ValidateConfig() (*ConfigValidationResult, error) {
	return s.plugin.ValidateConfig()
}

// UploadImage uploads an image from base64 content
func (s *ImageBedService) UploadImage(fileName string, base64Content string) (*UploadResult, error) {
	return s.plugin.UploadImage(fileName, base64Content)
}

// GetUploadHistory returns all upload records
func (s *ImageBedService) GetUploadHistory() []UploadRecord {
	return s.plugin.GetRecords()
}

// DeleteImage deletes an image from GitHub and history
func (s *ImageBedService) DeleteImage(id string) (*UploadResult, error) {
	return s.plugin.DeleteImage(id)
}

// GetUploadRecord returns a single record by ID
func (s *ImageBedService) GetUploadRecord(id string) *UploadRecord {
	return s.plugin.GetRecordByID(id)
}

// GetFormattedLink returns the image URL in the specified format
func (s *ImageBedService) GetFormattedLink(id string, format string) string {
	record := s.plugin.GetRecordByID(id)
	if record == nil {
		return ""
	}

	linkFormat := LinkFormat(format)
	if linkFormat == "" {
		linkFormat = LinkFormatRaw
	}

	return FormatLink(record.CDNURL, record.FileName, linkFormat)
}

// ClearHistory clears all upload history (without deleting from GitHub)
func (s *ImageBedService) ClearHistory() error {
	s.plugin.history.Records = []UploadRecord{}
	return s.plugin.SaveHistory(s.plugin.dataDir)
}

// SyncFromRepository syncs images from GitHub repository to local history
func (s *ImageBedService) SyncFromRepository() ([]UploadRecord, error) {
	if s.plugin.uploader == nil {
		s.plugin.uploader = NewUploader(s.plugin.config)
	}

	records, err := s.plugin.uploader.ListRepositoryImages()
	if err != nil {
		return nil, err
	}

	// Merge with existing history
	existingIDs := make(map[string]bool)
	for _, record := range s.plugin.history.Records {
		existingIDs[record.ID] = true
	}

	// Add new records that don't exist locally
	for _, record := range records {
		if !existingIDs[record.ID] {
			s.plugin.history.Records = append(s.plugin.history.Records, record)
		}
	}

	// Save updated history
	if err := s.plugin.SaveHistory(s.plugin.dataDir); err != nil {
		return nil, err
	}

	return s.plugin.GetRecords(), nil
}

// RenameImage renames an image file in GitHub repository
func (s *ImageBedService) RenameImage(id string, newFileName string) (*UploadResult, error) {
	return s.plugin.RenameImage(id, newFileName)
}
