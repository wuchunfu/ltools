package sticky

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// StickyService exposes sticky note functionality to frontend
type StickyService struct {
	plugin *StickyPlugin
	app    *application.App
}

// NewStickyService creates a new sticky service
func NewStickyService(plugin *StickyPlugin, app *application.App, dataDir string) *StickyService {
	// Set dataDir on plugin if not already set
	if plugin.dataDir == "" {
		plugin.dataDir = dataDir
	}
	return &StickyService{
		plugin: plugin,
		app:    app,
	}
}

// ServiceStartup is called when the application starts
func (s *StickyService) ServiceStartup(app *application.App) error {
	return s.plugin.ServiceStartup(app)
}

// ServiceShutdown is called when the application shuts down
func (s *StickyService) ServiceShutdown(app *application.App) error {
	return s.plugin.ServiceShutdown(app)
}

// ============================================================================
// Sticky Note Management API
// ============================================================================

// CreateNote creates a new sticky note, opens its window, and saves config
func (s *StickyService) CreateNote() (*StickyNote, error) {
	now := time.Now()

	// Calculate default position with cascade effect
	x, y := s.calculateDefaultPosition()

	// Random color
	color := s.randomColor()

	note := &StickyNote{
		ID:        generateID(),
		Content:   "",
		Color:     color,
		X:         x,
		Y:         y,
		Width:     280,
		Height:    320,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Add to config
	s.plugin.config.Notes = append(s.plugin.config.Notes, *note)

	// Save config
	if err := s.plugin.SaveConfig(s.plugin.dataDir); err != nil {
		return nil, fmt.Errorf("保存配置失败: %w", err)
	}

	// Create and show window
	if s.plugin.windowManager != nil {
		_, err := s.plugin.windowManager.CreateWindow(note.ID, note.X, note.Y, note.Width, note.Height, note.Color)
		if err != nil {
			return nil, fmt.Errorf("创建窗口失败: %w", err)
		}
	}

	// Emit event
	s.plugin.emitEvent("created", note.ID)

	return note, nil
}

// GetNote gets a note by ID
func (s *StickyService) GetNote(id string) (*StickyNote, error) {
	for i := range s.plugin.config.Notes {
		if s.plugin.config.Notes[i].ID == id {
			return &s.plugin.config.Notes[i], nil
		}
	}
	return nil, fmt.Errorf("便签未找到: %s", id)
}

// UpdateNote updates a note's content and/or position
func (s *StickyService) UpdateNote(note StickyNote) error {
	log.Printf("[StickyService] UpdateNote called: id=%s, content length=%d", note.ID, len(note.Content))
	for i := range s.plugin.config.Notes {
		if s.plugin.config.Notes[i].ID == note.ID {
			// Preserve created time
			note.CreatedAt = s.plugin.config.Notes[i].CreatedAt
			note.UpdatedAt = time.Now()

			s.plugin.config.Notes[i] = note

			// Save config
			if err := s.plugin.SaveConfig(s.plugin.dataDir); err != nil {
				log.Printf("[StickyService] SaveConfig failed: %v", err)
				return fmt.Errorf("保存配置失败: %w", err)
			}

			log.Printf("[StickyService] Note saved successfully: id=%s, content length=%d", note.ID, len(note.Content))
			// Emit event
			s.plugin.emitEvent("updated", note.ID)
			return nil
		}
	}
	return fmt.Errorf("便签未找到: %s", note.ID)
}

// DeleteNote deletes a note by ID, closes its window if open, and saves config
func (s *StickyService) DeleteNote(id string) error {
	for i, note := range s.plugin.config.Notes {
		if note.ID == id {
			// Close window if open
			if s.plugin.windowManager != nil {
				if err := s.plugin.windowManager.CloseWindow(id); err != nil {
					// Window might not be open, that's ok
				}
			}

			// Remove from config
			s.plugin.config.Notes = append(
				s.plugin.config.Notes[:i],
				s.plugin.config.Notes[i+1:]...,
			)

			// Save config
			if err := s.plugin.SaveConfig(s.plugin.dataDir); err != nil {
				return fmt.Errorf("保存配置失败: %w", err)
			}

			// Emit event
			s.plugin.emitEvent("deleted", id)
			return nil
		}
	}
	return fmt.Errorf("便签未找到: %s", id)
}

// ListNotes returns all sticky notes
func (s *StickyService) ListNotes() ([]StickyNote, error) {
	if s.plugin.config == nil {
		return []StickyNote{}, nil
	}
	return s.plugin.config.Notes, nil
}

// OpenNoteWindow opens a window for an existing note
func (s *StickyService) OpenNoteWindow(id string) error {
	// Find the note
	var note *StickyNote
	for i := range s.plugin.config.Notes {
		if s.plugin.config.Notes[i].ID == id {
			note = &s.plugin.config.Notes[i]
			break
		}
	}
	if note == nil {
		return fmt.Errorf("便签未找到: %s", id)
	}

	// Check if window already exists
	if s.plugin.windowManager != nil {
		if s.plugin.windowManager.HasWindow(id) {
			// Focus existing window
			return s.plugin.windowManager.FocusWindow(id)
		}

		// Create new window
		_, err := s.plugin.windowManager.CreateWindow(note.ID, note.X, note.Y, note.Width, note.Height, note.Color)
		if err != nil {
			return fmt.Errorf("创建窗口失败: %w", err)
		}
	}

	return nil
}

// CloseNoteWindow closes a specific note window
func (s *StickyService) CloseNoteWindow(id string) error {
	if s.plugin.windowManager == nil {
		return fmt.Errorf("窗口管理器未初始化")
	}
	return s.plugin.windowManager.CloseWindow(id)
}

// CloseAllWindows closes all sticky note windows
func (s *StickyService) CloseAllWindows() error {
	if s.plugin.windowManager == nil {
		return fmt.Errorf("窗口管理器未初始化")
	}
	s.plugin.windowManager.CloseAll()
	return nil
}

// GetAvailableColors returns the available color options
func (s *StickyService) GetAvailableColors() map[string]string {
	return StickyColors
}

// HandleShortcut handles global shortcut for creating a new note
func (s *StickyService) HandleShortcut(pluginID string) {
	// Create a new note when shortcut is triggered
	if pluginID == PluginID {
		_, err := s.CreateNote()
		if err != nil {
			// Log error but don't return it (shortcut handlers typically don't return errors)
			fmt.Printf("[Sticky] Failed to create note from shortcut: %v\n", err)
		}
	}
}

// ============================================================================
// Helper Functions
// ============================================================================

// calculateDefaultPosition calculates a default position for new notes with cascade effect
func (s *StickyService) calculateDefaultPosition() (int, int) {
	// 使用屏幕左上角作为基准点
	baseX, baseY := 100, 100
	offset := 150 // 每个窗口偏移150px，非常明显

	// Count existing notes to calculate cascade offset
	count := len(s.plugin.config.Notes)

	// 计算偏移，每5个窗口重置一次
	cascadeIndex := count % 5
	x := baseX + (cascadeIndex * offset)
	y := baseY + (cascadeIndex * offset)

	log.Printf("[Sticky] Calculated position for note #%d (cascade %d): (%d, %d)", count, cascadeIndex, x, y)

	return x, y
}

// randomColor returns a random color from the available colors
func (s *StickyService) randomColor() string {
	colors := []string{"yellow", "pink", "green", "blue", "purple"}
	return colors[rand.Intn(len(colors))]
}
