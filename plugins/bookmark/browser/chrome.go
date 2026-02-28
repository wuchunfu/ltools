package browser

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

// ChromeParser Chrome 书签解析器
type ChromeParser struct {
	bookmarksPath string
}

// NewChromeParser 创建 Chrome 解析器
func NewChromeParser() *ChromeParser {
	return &ChromeParser{}
}

// Name 返回浏览器名称
func (p *ChromeParser) Name() string {
	return "Chrome"
}

// GetBookmarksPath 获取书签文件路径
func (p *ChromeParser) GetBookmarksPath() (string, error) {
	if p.bookmarksPath != "" {
		return p.bookmarksPath, nil
	}

	var basePath string
	switch runtime.GOOS {
	case "darwin":
		basePath = filepath.Join(os.Getenv("HOME"), "Library", "Application Support", "Google", "Chrome", "Default")
	case "windows":
		basePath = filepath.Join(os.Getenv("LOCALAPPDATA"), "Google", "Chrome", "User Data", "Default")
	case "linux":
		basePath = filepath.Join(os.Getenv("HOME"), ".config", "google-chrome", "Default")
	default:
		return "", fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	p.bookmarksPath = filepath.Join(basePath, "Bookmarks")
	return p.bookmarksPath, nil
}

// IsAvailable 检查 Chrome 是否可用
func (p *ChromeParser) IsAvailable() bool {
	path, err := p.GetBookmarksPath()
	if err != nil {
		return false
	}

	_, err = os.Stat(path)
	return err == nil
}

// chromeBookmark Chrome 书签 JSON 结构
type chromeBookmark struct {
	Checksum string `json:"checksum"`
	Roots    struct {
		BookmarkBar chromeBookmarkNode `json:"bookmark_bar"`
		Other       chromeBookmarkNode `json:"other"`
		Synced      chromeBookmarkNode `json:"synced"`
	} `json:"roots"`
	Version int `json:"version"`
}

// chromeBookmarkNode Chrome 书签节点
type chromeBookmarkNode struct {
	ID           string               `json:"id"`
	Name         string               `json:"name"`
	Type         string               `json:"type"` // "url" or "folder"
	URL          string               `json:"url,omitempty"`
	DateAdded    string               `json:"date_added,omitempty"`
	Children     []chromeBookmarkNode `json:"children,omitempty"`
}

// Parse 解析 Chrome 书签
func (p *ChromeParser) Parse() ([]Bookmark, error) {
	path, err := p.GetBookmarksPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read bookmarks file: %w", err)
	}

	var chromeData chromeBookmark
	if err := json.Unmarshal(data, &chromeData); err != nil {
		return nil, fmt.Errorf("failed to parse bookmarks JSON: %w", err)
	}

	var bookmarks []Bookmark

	// 解析书签栏
	bookmarks = append(bookmarks, p.parseNode(&chromeData.Roots.BookmarkBar, "书签栏")...)

	// 解析其他书签
	bookmarks = append(bookmarks, p.parseNode(&chromeData.Roots.Other, "其他书签")...)

	// 解析同步书签
	bookmarks = append(bookmarks, p.parseNode(&chromeData.Roots.Synced, "移动设备书签")...)

	return bookmarks, nil
}

// parseNode 递归解析书签节点
func (p *ChromeParser) parseNode(node *chromeBookmarkNode, folderPath string) []Bookmark {
	var bookmarks []Bookmark

	if node.Type == "url" && node.URL != "" {
		// 解析 Chrome 时间格式（WebKit timestamp: microseconds since 1601-01-01）
		addedAt := time.Now()
		if node.DateAdded != "" {
			if timestamp, err := parseChromeTimestamp(node.DateAdded); err == nil {
				addedAt = timestamp
			}
		}

		bookmarks = append(bookmarks, Bookmark{
			ID:      generateID(node.URL),
			Title:   node.Name,
			URL:     node.URL,
			Folder:  folderPath,
			Browser: "chrome",
			AddedAt: addedAt,
		})
	} else if node.Type == "folder" && len(node.Children) > 0 {
		// 递归处理子节点
		for i := range node.Children {
			newPath := folderPath + "/" + node.Name
			bookmarks = append(bookmarks, p.parseNode(&node.Children[i], newPath)...)
		}
	}

	return bookmarks
}

// parseChromeTimestamp 解析 Chrome 时间戳
func parseChromeTimestamp(timestamp string) (time.Time, error) {
	// Chrome 使用 WebKit timestamp: microseconds since 1601-01-01 00:00:00 UTC
	var microseconds int64
	if _, err := fmt.Sscanf(timestamp, "%d", &microseconds); err != nil {
		return time.Time{}, err
	}

	// 转换为 Unix timestamp
	seconds := microseconds / 1000000
	nanos := (microseconds % 1000000) * 1000

	// WebKit epoch: 1601-01-01
	webkitEpoch := time.Date(1601, 1, 1, 0, 0, 0, 0, time.UTC)
	return webkitEpoch.Add(time.Duration(seconds)*time.Second + time.Duration(nanos)*time.Nanosecond), nil
}

// generateID 生成书签 ID（URL hash）
func generateID(url string) string {
	hash := sha256.Sum256([]byte(url))
	return hex.EncodeToString(hash[:])[:16]
}
