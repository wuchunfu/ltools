package password

import (
	"github.com/wailsapp/wails/v3/pkg/application"
)

// PasswordService exposes password plugin functionality to frontend
// Note: Most password generation logic is handled on the frontend using crypto.getRandomValues()
// This service exists for plugin system integration
type PasswordService struct {
	app    *application.App
	plugin  *PasswordPlugin
}

// NewPasswordService creates a new password service
func NewPasswordService(plugin *PasswordPlugin, app *application.App) *PasswordService {
	return &PasswordService{
		app:   app,
		plugin: plugin,
	}
}

// ServiceStartup is called when application starts
func (s *PasswordService) ServiceStartup(app *application.App) error {
	return nil
}

// ServiceShutdown is called when application shuts down
func (s *PasswordService) ServiceShutdown(app *application.App) error {
	return nil
}
