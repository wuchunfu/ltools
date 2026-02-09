package plugins

import (
	"path/filepath"
	"testing"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// TestNewManager tests creating a new plugin manager
func TestNewManager(t *testing.T) {
	app := application.New(application.Options{
		Name: "Test App",
	})

	tempDir := t.TempDir()
	dataDir := filepath.Join(tempDir, "test-plugins")

	manager, err := NewManager(app, dataDir)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	if manager == nil {
		t.Fatal("Manager is nil")
	}

	if manager.registry == nil {
		t.Error("Registry is nil")
	}
}

// TestRegisterPlugin tests registering a plugin
func TestRegisterPlugin(t *testing.T) {
	app := application.New(application.Options{
		Name: "Test App",
	})

	tempDir := t.TempDir()
	dataDir := filepath.Join(tempDir, "test-plugins")

	manager, err := NewManager(app, dataDir)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Create a test plugin
	metadata := &PluginMetadata{
		ID:          "test.plugin",
		Name:        "Test Plugin",
		Version:     "1.0.0",
		Author:      "Test",
		Description: "A test plugin",
		Type:        PluginTypeBuiltIn,
		State:       PluginStateInstalled,
	}

	plugin := NewBasePlugin(metadata)

	// Register the plugin
	err = manager.Register(plugin)
	if err != nil {
		t.Fatalf("Failed to register plugin: %v", err)
	}

	// Verify plugin is registered
	retrieved, ok := manager.Get("test.plugin")
	if !ok {
		t.Fatal("Plugin not found after registration")
	}

	if retrieved.Metadata().ID != "test.plugin" {
		t.Errorf("Expected plugin ID 'test.plugin', got '%s'", retrieved.Metadata().ID)
	}
}

// TestEnableDisablePlugin tests enabling and disabling plugins
func TestEnableDisablePlugin(t *testing.T) {
	app := application.New(application.Options{
		Name: "Test App",
	})

	tempDir := t.TempDir()
	dataDir := filepath.Join(tempDir, "test-plugins")

	manager, err := NewManager(app, dataDir)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Create and register a test plugin
	metadata := &PluginMetadata{
		ID:          "test.plugin",
		Name:        "Test Plugin",
		Version:     "1.0.0",
		Author:      "Test",
		Description: "A test plugin",
		Type:        PluginTypeBuiltIn,
		State:       PluginStateInstalled,
	}

	plugin := NewBasePlugin(metadata)
	err = manager.Register(plugin)
	if err != nil {
		t.Fatalf("Failed to register plugin: %v", err)
	}

	// Verify initial state - BasePlugin defaults to enabled
	if !plugin.Enabled() {
		t.Error("Plugin should be enabled initially (BasePlugin default)")
	}

	// First disable the plugin
	err = manager.Disable("test.plugin")
	if err != nil {
		t.Fatalf("Failed to disable plugin: %v", err)
	}

	if plugin.Enabled() {
		t.Error("Plugin should be disabled after Disable()")
	}

	// Now test enabling
	err = manager.Enable("test.plugin")
	if err != nil {
		t.Fatalf("Failed to enable plugin: %v", err)
	}

	if !plugin.Enabled() {
		t.Error("Plugin should be enabled after Enable()")
	}

	// Verify the state was updated
	retrievedPlugin, ok := manager.Get("test.plugin")
	if !ok {
		t.Fatal("Plugin not found after enable")
	}

	if !retrievedPlugin.Enabled() {
		t.Error("Retrieved plugin should be enabled")
	}

	// Test disabling
	err = manager.Disable("test.plugin")
	if err != nil {
		t.Fatalf("Failed to disable plugin: %v", err)
	}

	if plugin.Enabled() {
		t.Error("Plugin should be disabled after Disable()")
	}

	// Verify the state was updated
	retrievedPlugin, ok = manager.Get("test.plugin")
	if !ok {
		t.Fatal("Plugin not found after disable")
	}

	if retrievedPlugin.Enabled() {
		t.Error("Retrieved plugin should be disabled")
	}
}

// TestPluginSearch tests searching for plugins
func TestPluginSearch(t *testing.T) {
	app := application.New(application.Options{
		Name: "Test App",
	})

	tempDir := t.TempDir()
	dataDir := filepath.Join(tempDir, "test-plugins")

	manager, err := NewManager(app, dataDir)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Register test plugins
	plugins := []*PluginMetadata{
		{
			ID:          "datetime.plugin",
			Name:        "Date Time",
			Version:     "1.0.0",
			Author:      "Test",
			Description: "A datetime plugin",
			Type:        PluginTypeBuiltIn,
			State:       PluginStateInstalled,
			Keywords:    []string{"time", "date", "clock"},
		},
		{
			ID:          "calculator.plugin",
			Name:        "Calculator",
			Version:     "1.0.0",
			Author:      "Test",
			Description: "A calculator plugin",
			Type:        PluginTypeBuiltIn,
			State:       PluginStateInstalled,
			Keywords:    []string{"math", "calc", "calculator"},
		},
	}

	for _, metadata := range plugins {
		plugin := NewBasePlugin(metadata)
		err = manager.Register(plugin)
		if err != nil {
			t.Fatalf("Failed to register plugin %s: %v", metadata.ID, err)
		}
	}

	// Test searching
	results := manager.Search("time")
	if len(results) != 1 {
		t.Errorf("Expected 1 result for 'time', got %d", len(results))
	}

	results = manager.Search("calc")
	if len(results) != 1 {
		t.Errorf("Expected 1 result for 'calc', got %d", len(results))
	}

	results = manager.Search("test")
	if len(results) != 2 {
		t.Errorf("Expected 2 results for 'test', got %d", len(results))
	}
}

// TestRegistry tests the plugin registry
func TestRegistry(t *testing.T) {
	tempDir := t.TempDir()

	registry, err := NewRegistry(tempDir)
	if err != nil {
		t.Fatalf("Failed to create registry: %v", err)
	}

	// Register a plugin
	metadata := &PluginMetadata{
		ID:          "test.plugin",
		Name:        "Test Plugin",
		Version:     "1.0.0",
		Author:      "Test",
		Description: "A test plugin",
		Type:        PluginTypeBuiltIn,
		State:       PluginStateInstalled,
	}

	err = registry.Register(metadata)
	if err != nil {
		t.Fatalf("Failed to register plugin: %v", err)
	}

	// Retrieve the plugin
	retrieved, ok := registry.Get("test.plugin")
	if !ok {
		t.Fatal("Plugin not found")
	}

	if retrieved.Name != "Test Plugin" {
		t.Errorf("Expected name 'Test Plugin', got '%s'", retrieved.Name)
	}

	// List all plugins
	list := registry.List()
	if len(list) != 1 {
		t.Errorf("Expected 1 plugin, got %d", len(list))
	}

	// Test persistence
	registry2, err := NewRegistry(tempDir)
	if err != nil {
		t.Fatalf("Failed to create second registry: %v", err)
	}

	retrieved2, ok := registry2.Get("test.plugin")
	if !ok {
		t.Fatal("Plugin not found in second registry")
	}

	if retrieved2.Name != "Test Plugin" {
		t.Errorf("Expected name 'Test Plugin', got '%s'", retrieved2.Name)
	}
}
