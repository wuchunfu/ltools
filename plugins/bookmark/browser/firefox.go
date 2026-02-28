package browser

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// FirefoxParser Firefox 书签解析器
type FirefoxParser struct {
	bookmarksPath string
}

// NewFirefoxParser 创建 Firefox 解析器
func NewFirefoxParser() *FirefoxParser {
	return &FirefoxParser{}
}

// Name 返回浏览器名称
func (p *FirefoxParser) Name() string {
	return "Firefox"
}

// GetBookmarksPath 获取书签数据库路径
func (p *FirefoxParser) GetBookmarksPath() (string, error) {
	if p.bookmarksPath != "" {
		return p.bookmarksPath, nil
	}

	var profilesDir string
	switch runtime.GOOS {
	case "darwin":
		profilesDir = filepath.Join(os.Getenv("HOME"), "Library", "Application Support", "Firefox", "Profiles")
	case "windows":
		profilesDir = filepath.Join(os.Getenv("APPDATA"), "Mozilla", "Firefox", "Profiles")
	case "linux":
		profilesDir = filepath.Join(os.Getenv("HOME"), ".mozilla", "firefox")
	default:
		return "", fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	// 查找默认配置文件（以 .default 或 .default-release 结尾）
	entries, err := os.ReadDir(profilesDir)
	if err != nil {
		return "", fmt.Errorf("failed to read profiles directory: %w", err)
	}

	var defaultProfile string
	for _, entry := range entries {
		if entry.IsDir() {
			name := entry.Name()
			// 优先选择 .default-release，然后是 .default
			if filepath.Ext(name) == ".default-release" || filepath.Ext(name) == ".default" {
				if defaultProfile == "" || filepath.Ext(name) == ".default-release" {
					defaultProfile = filepath.Join(profilesDir, name)
				}
			}
		}
	}

	if defaultProfile == "" {
		return "", fmt.Errorf("no Firefox profile found")
	}

	p.bookmarksPath = filepath.Join(defaultProfile, "places.sqlite")
	return p.bookmarksPath, nil
}

// IsAvailable 检查 Firefox 是否可用
func (p *FirefoxParser) IsAvailable() bool {
	path, err := p.GetBookmarksPath()
	if err != nil {
		return false
	}

	_, err = os.Stat(path)
	return err == nil
}

// Parse 解析 Firefox 书签
func (p *FirefoxParser) Parse() ([]Bookmark, error) {
	path, err := p.GetBookmarksPath()
	if err != nil {
		return nil, err
	}

	// SQLite 需要使用绝对路径
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	// 打开数据库（只读模式）
	db, err := sql.Open("sqlite3", "file:"+absPath+"?mode=ro")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	// 查询书签
	// Firefox 书签结构：
	// - moz_bookmarks: 存储书签和文件夹
	// - moz_places: 存储 URL 信息
	query := `
		SELECT
			b.id,
			COALESCE(b.title, p.url) as title,
			p.url,
			b.dateAdded,
			b.parent,
			b.type
		FROM moz_bookmarks b
		LEFT JOIN moz_places p ON b.fk = p.id
		WHERE b.type IN (1, 2)
		ORDER BY b.parent, b.position
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query bookmarks: %w", err)
	}
	defer rows.Close()

	// 构建文件夹 ID 到名称的映射
	folderNames, err := p.buildFolderMap(db)
	if err != nil {
		return nil, err
	}

	var bookmarks []Bookmark
	for rows.Next() {
		var id int
		var title, url sql.NullString
		var dateAdded int64
		var parent, bookmarkType int

		if err := rows.Scan(&id, &title, &url, &dateAdded, &parent, &bookmarkType); err != nil {
			continue
		}

		// type 1 = 书签, type 2 = 文件夹
		if bookmarkType == 1 && url.Valid && url.String != "" {
			// 转换 Firefox 时间戳（微秒）
			addedAt := time.Unix(dateAdded/1000000, (dateAdded%1000000)*1000)

			// 构建文件夹路径
			folderPath := p.buildFolderPath(parent, folderNames)

			bookmarks = append(bookmarks, Bookmark{
				ID:      generateID(url.String),
				Title:   title.String,
				URL:     url.String,
				Folder:  folderPath,
				Browser: "firefox",
				AddedAt: addedAt,
			})
		}
	}

	return bookmarks, nil
}

// buildFolderMap 构建文件夹 ID 到名称的映射
func (p *FirefoxParser) buildFolderMap(db *sql.DB) (map[int]folderInfo, error) {
	query := `
		SELECT id, parent, title, type
		FROM moz_bookmarks
		WHERE type = 2
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query folders: %w", err)
	}
	defer rows.Close()

	folders := make(map[int]folderInfo)
	for rows.Next() {
		var id, parent int
		var title sql.NullString

		if err := rows.Scan(&id, &parent, &title); err != nil {
			continue
		}

		folders[id] = folderInfo{
			Title:  title.String,
			Parent: parent,
		}
	}

	return folders, nil
}

type folderInfo struct {
	Title  string
	Parent int
}

// buildFolderPath 构建文件夹路径
func (p *FirefoxParser) buildFolderPath(folderID int, folders map[int]folderInfo) string {
	if folderID == 0 || folderID == 1 {
		return ""
	}

	var path []string
	currentID := folderID
	maxDepth := 20 // 防止无限循环

	for i := 0; i < maxDepth && currentID > 1; i++ {
		info, exists := folders[currentID]
		if !exists {
			break
		}

		if info.Title != "" && info.Title != "menu" && info.Title != "toolbar" && info.Title != "unfiled" {
			path = append([]string{info.Title}, path...)
		} else if info.Title == "menu" {
			path = append([]string{"书签菜单"}, path...)
		} else if info.Title == "toolbar" {
			path = append([]string{"书签工具栏"}, path...)
		} else if info.Title == "unfiled" {
			path = append([]string{"其他书签"}, path...)
		}

		currentID = info.Parent
	}

	result := ""
	for i, part := range path {
		if i > 0 {
			result += "/"
		}
		result += part
	}

	return result
}
