package datetime

import (
	"fmt"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"
	"ltools/internal/plugins"
)

const (
	PluginID      = "datetime.builtin"
	PluginName    = "日期时间"
	PluginVersion = "1.0.0"
)

// DateTimePlugin provides current date and time functionality
type DateTimePlugin struct {
	*plugins.BasePlugin
	app *application.App
}

// NewDateTimePlugin creates a new DateTime plugin
func NewDateTimePlugin() *DateTimePlugin {
	metadata := &plugins.PluginMetadata{
		ID:          PluginID,
		Name:        PluginName,
		Version:     PluginVersion,
		Author:      "LTools",
		Description: "显示当前日期和时间",
		Icon:        "clock",
		Type:        plugins.PluginTypeBuiltIn,
		State:       plugins.PluginStateInstalled,
		Keywords:    []string{"时间", "日期", "时钟", "time", "date", "clock"},
	}

	base := plugins.NewBasePlugin(metadata)
	return &DateTimePlugin{
		BasePlugin: base,
	}
}

// Metadata returns the plugin metadata
func (p *DateTimePlugin) Metadata() *plugins.PluginMetadata {
	return p.BasePlugin.Metadata()
}

// Init initializes the plugin
func (p *DateTimePlugin) Init(app *application.App) error {
	if err := p.BasePlugin.Init(app); err != nil {
		return err
	}
	p.app = app
	return nil
}

// ServiceStartup is called when the application starts
func (p *DateTimePlugin) ServiceStartup(app *application.App) error {
	if err := p.BasePlugin.ServiceStartup(app); err != nil {
		return err
	}
	// Start a goroutine to emit time updates
	go p.emitTimeUpdates()
	return nil
}

// ServiceShutdown is called when the application shuts down
func (p *DateTimePlugin) ServiceShutdown(app *application.App) error {
	return p.BasePlugin.ServiceShutdown(app)
}

// Enabled returns true if the plugin is enabled
func (p *DateTimePlugin) Enabled() bool {
	return p.BasePlugin.Enabled()
}

// SetEnabled enables or disables the plugin
func (p *DateTimePlugin) SetEnabled(enabled bool) error {
	return p.BasePlugin.SetEnabled(enabled)
}


// emitTimeUpdates emits time update events every second
func (p *DateTimePlugin) emitTimeUpdates() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if p.Enabled() {
			now := time.Now()
			p.emitTimeEvent(now)
		}
	}
}

// emitTimeEvent emits a time event with the current time
func (p *DateTimePlugin) emitTimeEvent(now time.Time) {
	// Emit current time
	p.app.Event.Emit("datetime:current", now.Format(time.RFC3339))

	// Emit formatted time components
	p.app.Event.Emit("datetime:time", now.Format("15:04:05"))
	p.app.Event.Emit("datetime:date", now.Format("2006-01-02"))
	p.app.Event.Emit("datetime:datetime", now.Format("2006-01-02 15:04:05"))
	p.app.Event.Emit("datetime:weekday", now.Weekday().String())
	p.app.Event.Emit("datetime:year", now.Year())
	p.app.Event.Emit("datetime:month", int(now.Month()))
	p.app.Event.Emit("datetime:day", now.Day())
	p.app.Event.Emit("datetime:hour", now.Hour())
	p.app.Event.Emit("datetime:minute", now.Minute())
	p.app.Event.Emit("datetime:second", now.Second())
}

// GetCurrentTime returns the current time
func (p *DateTimePlugin) GetCurrentTime() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

// GetCurrentDate returns the current date
func (p *DateTimePlugin) GetCurrentDate() string {
	return time.Now().Format("2006-01-02")
}

// FormatTime formats a time string with the given format
func (p *DateTimePlugin) FormatTime(format, timeStr string) (string, error) {
	t, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		return "", err
	}
	return t.Format(format), nil
}

// GetTimestamp returns the current Unix timestamp
func (p *DateTimePlugin) GetTimestamp() int64 {
	return time.Now().Unix()
}

// TimestampToDateTime converts a Unix timestamp to datetime string
func (p *DateTimePlugin) TimestampToDateTime(timestamp int64) string {
	return time.Unix(timestamp, 0).Format("2006-01-02 15:04:05")
}

// DateTimeToTimestamp converts a datetime string to Unix timestamp
func (p *DateTimePlugin) DateTimeToTimestamp(datetimeStr string) (int64, error) {
	t, err := time.Parse("2006-01-02 15:04:05", datetimeStr)
	if err != nil {
		return 0, fmt.Errorf("invalid datetime format: %w", err)
	}
	return t.Unix(), nil
}

// GetWeekday returns the weekday name for the current date
func (p *DateTimePlugin) GetWeekday() string {
	weekdays := map[time.Weekday]string{
		time.Sunday:    "周日",
		time.Monday:    "周一",
		time.Tuesday:   "周二",
		time.Wednesday: "周三",
		time.Thursday:  "周四",
		time.Friday:    "周五",
		time.Saturday:  "周六",
	}
	return weekdays[time.Now().Weekday()]
}

// IsWeekend returns true if today is a weekend
func (p *DateTimePlugin) IsWeekend() bool {
	weekday := time.Now().Weekday()
	return weekday == time.Saturday || weekday == time.Sunday
}

// GetTimezone returns the current timezone
func (p *DateTimePlugin) GetTimezone() string {
	return time.Now().Location().String()
}
