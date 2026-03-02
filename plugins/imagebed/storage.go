package imagebed

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

const (
	historyFileName = "history.json"
	configVersion   = 1
)

// LoadHistory loads the upload history from the data directory
func (p *ImageBedPlugin) LoadHistory(dataDir string) error {
	// Ensure imagebed directory exists
	imagebedDir := filepath.Join(dataDir, "imagebed")
	if err := os.MkdirAll(imagebedDir, 0755); err != nil {
		return err
	}

	historyPath := filepath.Join(imagebedDir, historyFileName)

	// Read existing history or create default
	if data, err := os.ReadFile(historyPath); err == nil {
		if err := json.Unmarshal(data, p.history); err != nil {
			return err
		}
		return nil
	}

	// Create default empty history
	p.history = &UploadHistory{
		Version: configVersion,
		Records: []UploadRecord{},
	}
	return p.SaveHistory(dataDir)
}

// SaveHistory saves the upload history to the data directory
func (p *ImageBedPlugin) SaveHistory(dataDir string) error {
	imagebedDir := filepath.Join(dataDir, "imagebed")
	if err := os.MkdirAll(imagebedDir, 0755); err != nil {
		return err
	}

	historyPath := filepath.Join(imagebedDir, historyFileName)
	data, err := json.MarshalIndent(p.history, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(historyPath, data, 0644)
}

// AddRecord adds a new upload record to history
func (p *ImageBedPlugin) AddRecord(record *UploadRecord, dataDir string) error {
	record.ID = generateID()
	record.UploadTime = time.Now()
	p.history.Records = append([]UploadRecord{*record}, p.history.Records...)

	// Limit history to 100 records
	if len(p.history.Records) > 100 {
		p.history.Records = p.history.Records[:100]
	}

	return p.SaveHistory(dataDir)
}

// DeleteRecord removes a record from history by ID
func (p *ImageBedPlugin) DeleteRecord(id string, dataDir string) error {
	for i, record := range p.history.Records {
		if record.ID == id {
			p.history.Records = append(p.history.Records[:i], p.history.Records[i+1:]...)
			return p.SaveHistory(dataDir)
		}
	}
	return nil
}

// GetRecords returns all upload records
func (p *ImageBedPlugin) GetRecords() []UploadRecord {
	return p.history.Records
}

// GetRecordByID returns a single record by ID
func (p *ImageBedPlugin) GetRecordByID(id string) *UploadRecord {
	for _, record := range p.history.Records {
		if record.ID == id {
			return &record
		}
	}
	return nil
}

// generateID generates a unique ID for upload records
func generateID() string {
	return time.Now().Format("20060102150405") + "-" + randomString(6)
}

// randomString generates a random string of specified length
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[time.Now().UnixNano()%int64(len(charset))]
		// Add small delay to ensure different values
		time.Sleep(time.Nanosecond)
	}
	return string(result)
}
