package hosts

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// HostsService exposes hosts management functionality to frontend
type HostsService struct {
	plugin  *HostsPlugin
	app     *application.App
	dataDir string
}

// NewHostsService creates a new hosts service
func NewHostsService(plugin *HostsPlugin, app *application.App, dataDir string) *HostsService {
	return &HostsService{
		plugin:  plugin,
		app:     app,
		dataDir: dataDir,
	}
}

// SetApp sets the application instance
func (s *HostsService) SetApp(app *application.App) {
	s.app = app
}

// SetDataDir sets the data directory
func (s *HostsService) SetDataDir(dataDir string) {
	s.dataDir = dataDir
}

// ============================================================================
// Scenario Management API
// ============================================================================

// GetScenarios returns all scenarios
func (s *HostsService) GetScenarios() []Scenario {
	if s.plugin.config == nil {
		return []Scenario{}
	}
	return s.plugin.config.Scenarios
}

// CreateScenarioResult represents the result of creating a scenario
type CreateScenarioResult struct {
	Scenario *Scenario `json:"scenario"`
	Error    string    `json:"error,omitempty"`
}

// CreateScenario creates a new scenario
func (s *HostsService) CreateScenario(name, description string) (*CreateScenarioResult, error) {
	if name == "" {
		return &CreateScenarioResult{Error: "场景名称不能为空"}, nil
	}

	// Generate unique ID
	id := generateID(name)

	// Check for duplicate ID
	for _, sc := range s.plugin.config.Scenarios {
		if sc.ID == id {
			return &CreateScenarioResult{Error: "场景 ID 冲突，请使用不同的名称"}, nil
		}
	}

	scenario := &Scenario{
		ID:          id,
		Name:        name,
		Description: description,
		Entries:     []HostEntry{},
		IsActive:    false,
		CreatedAt:   time.Now(),
	}

	s.plugin.config.Scenarios = append(s.plugin.config.Scenarios, *scenario)

	if err := s.plugin.SaveConfig(s.dataDir); err != nil {
		return &CreateScenarioResult{Error: fmt.Sprintf("保存配置失败: %v", err)}, nil
	}

	s.plugin.emitEvent("scenario:created", id)

	return &CreateScenarioResult{Scenario: scenario}, nil
}

// UpdateScenario updates an existing scenario
func (s *HostsService) UpdateScenario(id string, scenario Scenario) error {
	for i := range s.plugin.config.Scenarios {
		if s.plugin.config.Scenarios[i].ID == id {
			// Preserve the original ID and created time
			scenario.ID = id
			scenario.CreatedAt = s.plugin.config.Scenarios[i].CreatedAt
			// Preserve active state
			scenario.IsActive = s.plugin.config.Scenarios[i].IsActive

			s.plugin.config.Scenarios[i] = scenario

			if err := s.plugin.SaveConfig(s.dataDir); err != nil {
				return fmt.Errorf("保存配置失败: %w", err)
			}

			s.plugin.emitEvent("scenario:updated", id)
			return nil
		}
	}
	return fmt.Errorf("场景未找到: %s", id)
}

// DeleteScenario deletes a scenario
func (s *HostsService) DeleteScenario(id string) error {
	for i, sc := range s.plugin.config.Scenarios {
		if sc.ID == id {
			// Cannot delete active scenario
			if sc.IsActive {
				return fmt.Errorf("无法删除当前激活的场景")
			}

			s.plugin.config.Scenarios = append(
				s.plugin.config.Scenarios[:i],
				s.plugin.config.Scenarios[i+1:]...,
			)

			if err := s.plugin.SaveConfig(s.dataDir); err != nil {
				return fmt.Errorf("保存配置失败: %w", err)
			}

			s.plugin.emitEvent("scenario:deleted", id)
			return nil
		}
	}
	return fmt.Errorf("场景未找到: %s", id)
}

// SwitchScenario switches to a different scenario
func (s *HostsService) SwitchScenario(id string) error {
	// Find scenario
	var scenario *Scenario
	for i := range s.plugin.config.Scenarios {
		if s.plugin.config.Scenarios[i].ID == id {
			scenario = &s.plugin.config.Scenarios[i]
			break
		}
	}
	if scenario == nil {
		return fmt.Errorf("场景未找到: %s", id)
	}

	// Create backup before switching
	if _, err := s.plugin.CreateBackup(s.dataDir, id); err != nil {
		s.plugin.emitEvent("error", fmt.Sprintf("备份失败: %v", err))
		// Continue anyway, backup failure is not critical
	}

	// Read current hosts content
	content, err := ReadHostsFile()
	if err != nil {
		return fmt.Errorf("读取 hosts 文件失败: %w", err)
	}

	// Parse and extract system entries
	systemEntries, _, _, _ := ParseHostsFile(content)

	// Format new hosts content with scenario
	newContent := FormatHostsContent(systemEntries, scenario.Name, scenario.Entries)

	// Write to hosts file
	if err := WriteHostsFile(newContent); err != nil {
		return fmt.Errorf("写入 hosts 文件失败: %w", err)
	}

	// Update active states
	for i := range s.plugin.config.Scenarios {
		s.plugin.config.Scenarios[i].IsActive = (s.plugin.config.Scenarios[i].ID == id)
	}
	s.plugin.config.CurrentScenario = id

	if err := s.plugin.SaveConfig(s.dataDir); err != nil {
		return fmt.Errorf("保存配置失败: %w", err)
	}

	s.plugin.emitEvent("scenario:switched", id)

	return nil
}

// ============================================================================
// Entries Management API
// ============================================================================

// AddEntry adds a host entry to a scenario
func (s *HostsService) AddEntry(scenarioID string, entry HostEntry) error {
	for i := range s.plugin.config.Scenarios {
		if s.plugin.config.Scenarios[i].ID == scenarioID {
			s.plugin.config.Scenarios[i].Entries = append(
				s.plugin.config.Scenarios[i].Entries,
				entry,
			)

			if err := s.plugin.SaveConfig(s.dataDir); err != nil {
				return fmt.Errorf("保存配置失败: %w", err)
			}

			s.plugin.emitEvent("entry:added", scenarioID)
			return nil
		}
	}
	return fmt.Errorf("场景未找到: %s", scenarioID)
}

// UpdateEntry updates a host entry in a scenario
func (s *HostsService) UpdateEntry(scenarioID string, index int, entry HostEntry) error {
	for i := range s.plugin.config.Scenarios {
		if s.plugin.config.Scenarios[i].ID == scenarioID {
			if index < 0 || index >= len(s.plugin.config.Scenarios[i].Entries) {
				return fmt.Errorf("条目索引超出范围: %d", index)
			}

			s.plugin.config.Scenarios[i].Entries[index] = entry

			if err := s.plugin.SaveConfig(s.dataDir); err != nil {
				return fmt.Errorf("保存配置失败: %w", err)
			}

			s.plugin.emitEvent("entry:updated", scenarioID)
			return nil
		}
	}
	return fmt.Errorf("场景未找到: %s", scenarioID)
}

// RemoveEntry removes a host entry from a scenario
func (s *HostsService) RemoveEntry(scenarioID string, index int) error {
	for i := range s.plugin.config.Scenarios {
		if s.plugin.config.Scenarios[i].ID == scenarioID {
			if index < 0 || index >= len(s.plugin.config.Scenarios[i].Entries) {
				return fmt.Errorf("条目索引超出范围: %d", index)
			}

			s.plugin.config.Scenarios[i].Entries = append(
				s.plugin.config.Scenarios[i].Entries[:index],
				s.plugin.config.Scenarios[i].Entries[index+1:]...,
			)

			if err := s.plugin.SaveConfig(s.dataDir); err != nil {
				return fmt.Errorf("保存配置失败: %w", err)
			}

			s.plugin.emitEvent("entry:removed", scenarioID)
			return nil
		}
	}
	return fmt.Errorf("场景未找到: %s", scenarioID)
}

// ============================================================================
// Backup Management API
// ============================================================================

// ListBackups returns all backups
func (s *HostsService) ListBackups() []Backup {
	return s.plugin.ListBackups()
}

// RestoreBackup restores a backup
func (s *HostsService) RestoreBackup(id string) error {
	return s.plugin.RestoreBackup(s.dataDir, id)
}

// DeleteBackup deletes a backup
func (s *HostsService) DeleteBackup(id string) error {
	return s.plugin.DeleteBackup(s.dataDir, id)
}

// CleanOldBackups removes old backups
func (s *HostsService) CleanOldBackups(keepCount int) int {
	return s.plugin.CleanOldBackups(s.dataDir, keepCount)
}

// ============================================================================
// Validation & Preview API
// ============================================================================

// ValidateEntriesResult represents validation result
type ValidateEntriesResult struct {
	Valid bool     `json:"valid"`
	Errors []string `json:"errors"`
}

// ValidateEntries validates hosts entries
func (s *HostsService) ValidateEntries(entries []HostEntry) *ValidateEntriesResult {
	errors := ValidateEntries(entries)

	result := &ValidateEntriesResult{
		Valid:  len(errors) == 0,
		Errors: make([]string, len(errors)),
	}

	for i, err := range errors {
		result.Errors[i] = err.Error()
	}

	return result
}

// PreviewScenarioContent returns the hosts content for a scenario
func (s *HostsService) PreviewScenarioContent(id string) (string, error) {
	var scenario *Scenario
	for i := range s.plugin.config.Scenarios {
		if s.plugin.config.Scenarios[i].ID == id {
			scenario = &s.plugin.config.Scenarios[i]
			break
		}
	}
	if scenario == nil {
		return "", fmt.Errorf("场景未找到: %s", id)
	}

	return FormatHostsContent([]HostEntry{}, scenario.Name, scenario.Entries), nil
}

// GetCurrentHostsContent returns the current system hosts content
func (s *HostsService) GetCurrentHostsContent() (string, error) {
	return ReadHostsFile()
}

// ============================================================================
// System Info API
// ============================================================================

// GetSystemInfo returns system information
func (s *HostsService) GetSystemInfo() *SystemInfo {
	hasPriv := CheckPrivileges()
	currentScenario := ""
	if s.plugin.config != nil {
		currentScenario = s.plugin.config.CurrentScenario
	}

	return &SystemInfo{
		HostsPath:       GetHostsPath(),
		CurrentScenario: currentScenario,
		HasPrivileges:   hasPriv,
	}
}

// CheckPrivileges returns the current privilege status
func (s *HostsService) CheckPrivileges() bool {
	return CheckPrivileges()
}

// ============================================================================
// Helper Functions
// ============================================================================

// generateID generates a unique ID from a name
func generateID(name string) string {
	// Normalize name: lowercase, remove spaces, replace with hyphens
	normalized := strings.ToLower(strings.TrimSpace(name))
	normalized = strings.ReplaceAll(normalized, " ", "-")
	normalized = strings.ReplaceAll(normalized, "_", "-")

	// Remove any non-alphanumeric characters except hyphens
	result := strings.Builder{}
	for _, r := range normalized {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result.WriteRune(r)
		}
	}

	id := result.String()
	if id == "" {
		id = "scenario"
	}

	// Add hash suffix for uniqueness
	hash := md5.Sum([]byte(name + time.Now().Format(time.RFC3339Nano)))
	hashSuffix := hex.EncodeToString(hash[:])[:8]

	return fmt.Sprintf("%s-%s", id, hashSuffix)
}

// ============================================================================
// Hosts File Parsing & Formatting
// ============================================================================

const (
	hostsHeaderPrefix = "# === LTools Scenario: "
	hostsFooterPrefix = "# === End LTools Scenario"
)

// ParseHostsFile parses the hosts file content
// Returns: system entries, LTools entries, scenario name, error
func ParseHostsFile(content string) ([]HostEntry, []HostEntry, string, error) {
	lines := strings.Split(content, "\n")

	var systemEntries []HostEntry
	var ltoolsEntries []HostEntry
	var scenarioName string
	inLToolsSection := false

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip empty lines
		if line == "" {
			continue
		}

		// Check for LTools section markers
		if strings.HasPrefix(line, hostsHeaderPrefix) {
			inLToolsSection = true
			scenarioName = strings.TrimPrefix(line, hostsHeaderPrefix)
			scenarioName = strings.TrimSuffix(scenarioName, " ===")
			continue
		}

		if strings.HasPrefix(line, hostsFooterPrefix) {
			inLToolsSection = false
			continue
		}

		// Parse as host entry
		entry, isEntry := parseHostEntry(line)
		if isEntry {
			if inLToolsSection {
				ltoolsEntries = append(ltoolsEntries, entry)
			} else {
				systemEntries = append(systemEntries, entry)
			}
		}
	}

	return systemEntries, ltoolsEntries, scenarioName, nil
}

// parseHostEntry parses a single line as a host entry
func parseHostEntry(line string) (HostEntry, bool) {
	// Skip comments
	if strings.HasPrefix(line, "#") {
		return HostEntry{}, false
	}

	// Split by whitespace
	fields := strings.Fields(line)
	if len(fields) < 2 {
		return HostEntry{}, false
	}

	// First field is IP, rest are hostnames
	ip := fields[0]
	hostname := fields[1]

	// Check if there's an inline comment
	comment := ""
	commentIdx := strings.Index(line, "#")
	if commentIdx > 0 {
		comment = strings.TrimSpace(line[commentIdx+1:])
	}

	return HostEntry{
		IP:       ip,
		Hostname: hostname,
		Comment:  comment,
		Enabled:  !strings.HasPrefix(strings.TrimSpace(line), "#"),
	}, true
}

// FormatHostsContent formats hosts content with system and LTools entries
func FormatHostsContent(systemEntries []HostEntry, scenarioName string, ltoolsEntries []HostEntry) string {
	var builder strings.Builder

	// Write header
	builder.WriteString("#\n")
	builder.WriteString("# Hosts File - Managed by LTools\n")
	builder.WriteString("# Generated: " + time.Now().Format(time.RFC1123) + "\n")
	builder.WriteString("#\n\n")

	// Write system entries
	for _, entry := range systemEntries {
		formatHostEntry(&builder, &entry)
	}

	// Write LTools section if there are entries
	if len(ltoolsEntries) > 0 {
		builder.WriteString("\n")
		builder.WriteString(fmt.Sprintf("%s%s ===\n", hostsHeaderPrefix, scenarioName))
		builder.WriteString("# Do not edit this section manually\n")
		for _, entry := range ltoolsEntries {
			formatHostEntry(&builder, &entry)
		}
		builder.WriteString(fmt.Sprintf("%s\n", hostsFooterPrefix))
	}

	return builder.String()
}

// formatHostEntry formats a single host entry
func formatHostEntry(builder *strings.Builder, entry *HostEntry) {
	if !entry.Enabled {
		builder.WriteString("# ")
	}

	builder.WriteString(fmt.Sprintf("%s\t%s", entry.IP, entry.Hostname))

	if entry.Comment != "" {
		builder.WriteString(fmt.Sprintf("\t# %s", entry.Comment))
	}

	builder.WriteString("\n")
}

// ValidateEntries validates an array of host entries
func ValidateEntries(entries []HostEntry) []error {
	var errors []error

	for i, entry := range entries {
		// Validate IP
		if entry.IP == "" {
			errors = append(errors, fmt.Errorf("条目 %d: IP 地址不能为空", i+1))
			continue
		}

		// Validate hostname
		if entry.Hostname == "" {
			errors = append(errors, fmt.Errorf("条目 %d: 主机名不能为空", i+1))
			continue
		}

		// Validate IP format (basic check)
		if !isValidIP(entry.IP) {
			errors = append(errors, fmt.Errorf("条目 %d: IP 地址格式无效: %s", i+1, entry.IP))
		}

		// Validate hostname format (basic check)
		if !isValidHostname(entry.Hostname) {
			errors = append(errors, fmt.Errorf("条目 %d: 主机名格式无效: %s", i+1, entry.Hostname))
		}
	}

	return errors
}

// isValidIP performs basic IP validation
func isValidIP(ip string) bool {
	// Basic IPv4 check
	parts := strings.Split(ip, ".")
	if len(parts) != 4 {
		// Could be IPv6 or localhost
		return ip == "::1" || ip == "localhost" || strings.Contains(ip, ":")
	}

	for _, part := range parts {
		if part == "" {
			return false
		}
		// Check if numeric
		num := 0
		for _, c := range part {
			if c < '0' || c > '9' {
				return false
			}
			num = num*10 + int(c-'0')
		}
		if num > 255 {
			return false
		}
	}

	return true
}

// isValidHostname performs basic hostname validation
func isValidHostname(hostname string) bool {
	if hostname == "" {
		return false
	}

	// Allow wildcards
	if strings.HasPrefix(hostname, "*.") {
		hostname = strings.TrimPrefix(hostname, "*.")
	}

	// Basic hostname validation
	for i, c := range hostname {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
			(c >= '0' && c <= '9') || c == '-' || c == '.') {
			return false
		}
		// Cannot start or end with hyphen or dot
		if (i == 0 || i == len(hostname)-1) && (c == '-' || c == '.') {
			return false
		}
	}

	return true
}
