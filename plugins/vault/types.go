package vault

import "time"

// VaultEntry 表示一个密码条目
type VaultEntry struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Website   string    `json:"website,omitempty"`
	Username  string    `json:"username"`
	Password  string    `json:"password"` // 加密存储
	Notes     string    `json:"notes,omitempty"`
	Category  string    `json:"category,omitempty"`
	Tags      []string  `json:"tags,omitempty"`
	Favorite  bool      `json:"favorite"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// VaultConfig 表示保险库配置
type VaultConfig struct {
	Version          int          `json:"version"`
	VerificationHash string       `json:"verificationHash"` // 用于验证主密码
	Salt             string       `json:"salt"`
	Entries          []VaultEntry `json:"entries"`
	Categories       []string     `json:"categories"`
}

// CreateEntryRequest 创建条目请求
type CreateEntryRequest struct {
	Title    string   `json:"title"`
	Website  string   `json:"website,omitempty"`
	Username string   `json:"username"`
	Password string   `json:"password"`
	Notes    string   `json:"notes,omitempty"`
	Category string   `json:"category,omitempty"`
	Tags     []string `json:"tags,omitempty"`
	Favorite bool     `json:"favorite"`
}

// UpdateEntryRequest 更新条目请求
type UpdateEntryRequest struct {
	ID       string   `json:"id"`
	Title    string   `json:"title,omitempty"`
	Website  string   `json:"website,omitempty"`
	Username string   `json:"username,omitempty"`
	Password string   `json:"password,omitempty"`
	Notes    string   `json:"notes,omitempty"`
	Category string   `json:"category,omitempty"`
	Tags     []string `json:"tags,omitempty"`
	Favorite *bool    `json:"favorite,omitempty"`
}

// SetupRequest 设置主密码请求
type SetupRequest struct {
	MasterPassword string `json:"masterPassword"`
}

// UnlockRequest 解锁请求
type UnlockRequest struct {
	MasterPassword string `json:"masterPassword"`
}

// VaultStatus 保险库状态
type VaultStatus struct {
	Initialized bool `json:"initialized"`
	Locked      bool `json:"locked"`
}

// SearchResult 搜索结果
type SearchResult struct {
	Entries []VaultEntry `json:"entries"`
	Total   int          `json:"total"`
}
