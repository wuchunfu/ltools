package datetime

import "time"

// DateTimeService exposes DateTime functionality to the frontend
// This is a thin wrapper around the plugin that can be registered as a Wails service
type DateTimeService struct {
	plugin *DateTimePlugin
}

// NewDateTimeService creates a new DateTime service
func NewDateTimeService(plugin *DateTimePlugin) *DateTimeService {
	return &DateTimeService{
		plugin: plugin,
	}
}

// GetCurrentTime returns the current time
func (s *DateTimeService) GetCurrentTime() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

// GetCurrentDate returns the current date
func (s *DateTimeService) GetCurrentDate() string {
	return time.Now().Format("2006-01-02")
}

// FormatTime formats a time string with the given format
func (s *DateTimeService) FormatTime(format, timeStr string) (string, error) {
	t, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		return "", err
	}
	return t.Format(format), nil
}

// GetTimestamp returns the current Unix timestamp
func (s *DateTimeService) GetTimestamp() int64 {
	return time.Now().Unix()
}

// TimestampToDateTime converts a Unix timestamp to datetime string
func (s *DateTimeService) TimestampToDateTime(timestamp int64) string {
	return time.Unix(timestamp, 0).Format("2006-01-02 15:04:05")
}

// DateTimeToTimestamp converts a datetime string to Unix timestamp
func (s *DateTimeService) DateTimeToTimestamp(datetimeStr string) (int64, error) {
	t, err := time.Parse("2006-01-02 15:04:05", datetimeStr)
	if err != nil {
		return 0, err
	}
	return t.Unix(), nil
}

// GetWeekday returns the weekday name for the current date
func (s *DateTimeService) GetWeekday() string {
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
func (s *DateTimeService) IsWeekend() bool {
	weekday := time.Now().Weekday()
	return weekday == time.Saturday || weekday == time.Sunday
}

// GetTimezone returns the current timezone
func (s *DateTimeService) GetTimezone() string {
	return time.Now().Location().String()
}
