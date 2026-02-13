package hosts

import "time"

// HostEntry - 单个 hosts 条目
type HostEntry struct {
	IP       string `json:"ip"`
	Hostname string `json:"hostname"`
	Comment  string `json:"comment,omitempty"`
	Enabled  bool   `json:"enabled"`
}

// Scenario - 场景配置
type Scenario struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Entries     []HostEntry `json:"entries"`
	IsActive    bool        `json:"isActive"`
	CreatedAt   time.Time   `json:"createdAt"`
}

// Backup - 备份记录
type Backup struct {
	ID         string    `json:"id"`
	ScenarioID string    `json:"scenarioId"`
	CreatedAt  time.Time `json:"createdAt"`
	Size       int64     `json:"size"`
}

// HostsConfig - 持久化配置
type HostsConfig struct {
	Version         int         `json:"version"`
	Scenarios       []Scenario  `json:"scenarios"`
	Backups         []Backup    `json:"backups"`
	CurrentScenario string       `json:"currentScenario"`
}

// SystemInfo - 系统信息
type SystemInfo struct {
	HostsPath       string `json:"hostsPath"`
	CurrentScenario string `json:"currentScenario"`
	HasPrivileges   bool   `json:"hasPrivileges"`
}
