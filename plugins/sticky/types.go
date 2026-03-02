package sticky

import "time"

// StickyColor - 便签颜色常量
var StickyColors = map[string]string{
	"yellow": "#FEF3C7",
	"pink":   "#FCE7F3",
	"green":  "#D1FAE5",
	"blue":   "#DBEAFE",
	"purple": "#EDE9FE",
}

// StickyNote - 单个便签
 type StickyNote struct {
	ID        string    `json:"id"`
	Content   string    `json:"content"`
	Color     string    `json:"color"`
	X         int       `json:"x"`
	Y         int       `json:"y"`
	Width     int       `json:"width"`
	Height    int       `json:"height"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// StickyConfig - 持久化配置
type StickyConfig struct {
	Version int          `json:"version"`
	Notes   []StickyNote `json:"notes"`
}
