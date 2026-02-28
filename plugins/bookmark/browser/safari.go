package browser

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"howett.net/plist"
)

// SafariParser Safari 书签解析器
type SafariParser struct {
	bookmarksPath string
}

// NewSafariParser 创建 Safari 解析器
func NewSafariParser() *SafariParser {
	return &SafariParser{}
}

// Name 返回浏览器名称
func (p *SafariParser) Name() string {
	return "Safari"
}

// GetBookmarksPath 获取书签文件路径
func (p *SafariParser) GetBookmarksPath() (string, error) {
	if p.bookmarksPath != "" {
		return p.bookmarksPath, nil
	}

	// Safari 仅在 macOS 上可用
	if runtime.GOOS != "darwin" {
		return "", fmt.Errorf("Safari is only available on macOS")
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	p.bookmarksPath = filepath.Join(homeDir, "Library", "Safari", "Bookmarks.plist")
	return p.bookmarksPath, nil
}

// IsAvailable 检查 Safari 是否可用
func (p *SafariParser) IsAvailable() bool {
	path, err := p.GetBookmarksPath()
	if err != nil {
		return false
	}

	_, err = os.Stat(path)
	return err == nil
}

// safariBookmarkItem Safari 书签 plist 结构
type safariBookmarkItem struct {
	Title           string                 `plist:"Title"`
	URLString       string                 `plist:"URLString"`
	BookmarkType    string                 `plist:"BookmarkType"` // "WebBookmarkTypeList" or "WebBookmarkTypeLeaf"
	Children        []safariBookmarkItem   `plist:"Children"`
	WebBookmarkUUID string                 `plist:"WebBookmarkUUID"`
	DateAdded       *time.Time             `plist:"DateAdded"`
	DescendantCount int                    `plist:"DescendantCount"`
}

// safariBookmarksRoot Safari 书签根结构
type safariBookmarksRoot struct {
	Children []safariBookmarkItem `plist:"Children"`
}

// Parse 解析 Safari 书签
func (p *SafariParser) Parse() ([]Bookmark, error) {
	path, err := p.GetBookmarksPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read bookmarks file: %w", err)
	}

	var root safariBookmarksRoot
	if _, err := plist.Unmarshal(data, &root); err != nil {
		return nil, fmt.Errorf("failed to parse bookmarks plist: %w", err)
	}

	var bookmarks []Bookmark

	// 遍历根节点的子节点
	for _, child := range root.Children {
		// 跳过非书签类型的根节点（如 Reading List）
		if child.Title == "com.apple.ReadingList" {
			continue
		}

		bookmarks = append(bookmarks, p.parseItem(&child, "")...)
	}

	return bookmarks, nil
}

// parseItem 递归解析书签项
func (p *SafariParser) parseItem(item *safariBookmarkItem, folderPath string) []Bookmark {
	var bookmarks []Bookmark

	// 如果是书签（叶子节点）
	if item.BookmarkType == "WebBookmarkTypeLeaf" && item.URLString != "" {
		addedAt := time.Now()
		if item.DateAdded != nil {
			addedAt = *item.DateAdded
		}

		bookmarks = append(bookmarks, Bookmark{
			ID:      generateID(item.URLString),
			Title:   item.Title,
			URL:     item.URLString,
			Folder:  folderPath,
			Browser: "safari",
			AddedAt: addedAt,
		})
	}

	// 如果是文件夹（列表节点）
	if item.BookmarkType == "WebBookmarkTypeList" && len(item.Children) > 0 {
		// 特殊处理根文件夹名称
		newPath := folderPath
		if item.Title != "" && item.Title != "BookmarksBar" && item.Title != "BookmarksMenu" {
			if folderPath == "" {
				newPath = item.Title
			} else {
				newPath = folderPath + "/" + item.Title
			}
		} else if item.Title == "BookmarksBar" {
			newPath = "书签栏"
		} else if item.Title == "BookmarksMenu" {
			newPath = "书签菜单"
		}

		for i := range item.Children {
			bookmarks = append(bookmarks, p.parseItem(&item.Children[i], newPath)...)
		}
	}

	return bookmarks
}
