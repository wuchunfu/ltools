package processmanager

import (
	"github.com/wailsapp/wails/v3/pkg/application"
	"ltools/internal/plugins"
)

// ProcessManagerService exposes ProcessManager functionality to the frontend
type ProcessManagerService struct {
	app    *application.App
	plugin plugins.Plugin
}

// NewProcessManagerService creates a new process manager service
func NewProcessManagerService(plugin *ProcessManagerPlugin, app *application.App) *ProcessManagerService {
	return &ProcessManagerService{
		app:    app,
		plugin: plugin,
	}
}

// getPlugin returns the plugin as ProcessManagerPlugin
func (s *ProcessManagerService) getPlugin() *ProcessManagerPlugin {
	return s.plugin.(*ProcessManagerPlugin)
}

// ServiceStartup is called when the application starts
func (s *ProcessManagerService) ServiceStartup(app *application.App) error {
	return s.getPlugin().ServiceStartup(app)
}

// ServiceShutdown is called when the application shuts down
func (s *ProcessManagerService) ServiceShutdown(app *application.App) error {
	return s.getPlugin().ServiceShutdown(app)
}

// GetProcesses returns the list of processes based on the given options
func (s *ProcessManagerService) GetProcesses(options ProcessListOptions) ([]*ProcessInfo, int, error) {
	return s.getPlugin().GetProcesses(options)
}

// GetProcessDetail returns detailed information about a single process
func (s *ProcessManagerService) GetProcessDetail(pid int) (*ProcessInfo, error) {
	return s.getPlugin().GetProcessDetail(pid)
}

// KillProcess terminates a process gracefully
func (s *ProcessManagerService) KillProcess(pid int) error {
	return s.getPlugin().KillProcess(pid)
}

// ForceKillProcess forcefully terminates a process
func (s *ProcessManagerService) ForceKillProcess(pid int) error {
	return s.getPlugin().ForceKillProcess(pid)
}

// ForceRefresh forces an immediate refresh of the process list
func (s *ProcessManagerService) ForceRefresh() {
	s.getPlugin().ForceRefresh()
}

// GetSystemInfo returns system information for process management
func (s *ProcessManagerService) GetSystemInfo() map[string]interface{} {
	return s.getPlugin().GetSystemInfo()
}

// EnterView notifies the plugin that the user has entered its view
func (s *ProcessManagerService) EnterView() error {
	plugin := s.plugin
	if viewPlugin, ok := plugin.(plugins.ViewLifecycle); ok {
		return viewPlugin.OnViewEnter(s.app)
	}
	return nil
}

// LeaveView notifies the plugin that the user has left its view
func (s *ProcessManagerService) LeaveView() error {
	plugin := s.plugin
	if viewPlugin, ok := plugin.(plugins.ViewLifecycle); ok {
		return viewPlugin.OnViewLeave(s.app)
	}
	return nil
}
