package bookmark

import (
	"encoding/json"
	"fmt"
	"html"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Exporter 导出器
type Exporter struct {
	plugin *BookmarkPlugin
}

// NewExporter 创建导出器
func NewExporter(plugin *BookmarkPlugin) *Exporter {
	return &Exporter{plugin: plugin}
}

// ExportHTML 导出为 HTML 格式（Netscape Bookmark File Format）
func (e *Exporter) ExportHTML(outputPath string) error {
	cacheData, err := e.plugin.cache.Load()
	if err != nil {
		return fmt.Errorf("failed to load bookmarks: %w", err)
	}

	if cacheData == nil || len(cacheData.Bookmarks) == 0 {
		return fmt.Errorf("no bookmarks to export")
	}

	var sb strings.Builder

	// HTML 头部
	sb.WriteString(`<!DOCTYPE NETSCAPE-Bookmark-file-1>
<!-- This is an automatically generated file.
     It will be read and overwritten.
     DO NOT EDIT! -->
<META HTTP-EQUIV="Content-Type" CONTENT="text/html; charset=UTF-8">
<TITLE>Bookmarks</TITLE>
<H1>Bookmarks</H1>
<DL><p>
`)

	// 按浏览器和文件夹组织书签
	browserBookmarks := e.groupByBrowser(cacheData.Bookmarks)

	for browser, bookmarks := range browserBookmarks {
		sb.WriteString(fmt.Sprintf(`    <DT><H3>%s 书签</H3>
    <DL><p>
`, browser))

		// 按文件夹组织
		folderBookmarks := e.groupByFolder(bookmarks)
		for folder, bms := range folderBookmarks {
			if folder != "" {
				sb.WriteString(fmt.Sprintf(`        <DT><H3>%s</H3>
        <DL><p>
`, html.EscapeString(folder)))
			}

			for _, bm := range bms {
				sb.WriteString(fmt.Sprintf(`            <DT><A HREF="%s" ADD_DATE="%d">%s</A>
`,
					html.EscapeString(bm.URL),
					bm.AddedAt.Unix(),
					html.EscapeString(bm.Title),
				))
			}

			if folder != "" {
				sb.WriteString("        </DL><p>\n")
			}
		}

		sb.WriteString("    </DL><p>\n")
	}

	sb.WriteString("</DL><p>\n")

	// 确保目录存在
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// 写入文件
	if err := os.WriteFile(outputPath, []byte(sb.String()), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// ExportJSON 导出为 JSON 格式
func (e *Exporter) ExportJSON(outputPath string) error {
	cacheData, err := e.plugin.cache.Load()
	if err != nil {
		return fmt.Errorf("failed to load bookmarks: %w", err)
	}

	if cacheData == nil || len(cacheData.Bookmarks) == 0 {
		return fmt.Errorf("no bookmarks to export")
	}

	// 获取浏览器列表
	browsers := make([]string, 0, len(cacheData.BrowserStats))
	for browser := range cacheData.BrowserStats {
		browsers = append(browsers, browser)
	}

	exportData := map[string]interface{}{
		"exported_at":  time.Now().Format(time.RFC3339),
		"total_count":  len(cacheData.Bookmarks),
		"browsers":     browsers,
		"browser_stats": cacheData.BrowserStats,
		"bookmarks":    cacheData.Bookmarks,
	}

	jsonData, err := json.MarshalIndent(exportData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	// 确保目录存在
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// 写入文件
	if err := os.WriteFile(outputPath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// groupByBrowser 按浏览器分组
func (e *Exporter) groupByBrowser(bookmarks []Bookmark) map[string][]Bookmark {
	result := make(map[string][]Bookmark)
	for _, bm := range bookmarks {
		result[bm.Browser] = append(result[bm.Browser], bm)
	}
	return result
}

// groupByFolder 按文件夹分组
func (e *Exporter) groupByFolder(bookmarks []Bookmark) map[string][]Bookmark {
	result := make(map[string][]Bookmark)
	for _, bm := range bookmarks {
		result[bm.Folder] = append(result[bm.Folder], bm)
	}
	return result
}
