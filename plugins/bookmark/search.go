package bookmark

import (
	"sort"
	"strings"
)

// SearchEngine 搜索引擎
type SearchEngine struct {
	bookmarks []Bookmark
	config    *BookmarkConfig
}

// NewSearchEngine 创建搜索引擎
func NewSearchEngine(config *BookmarkConfig) *SearchEngine {
	return &SearchEngine{
		config: config,
	}
}

// SetBookmarks 设置书签数据
func (e *SearchEngine) SetBookmarks(bookmarks []Bookmark) {
	e.bookmarks = bookmarks
}

// Search 搜索书签
func (e *SearchEngine) Search(query string) []SearchResult {
	if query == "" || len(e.bookmarks) == 0 {
		return []SearchResult{}
	}

	query = strings.ToLower(strings.TrimSpace(query))
	var results []SearchResult

	for i := range e.bookmarks {
		bm := &e.bookmarks[i]
		score, matchType := e.calculateScore(bm, query)

		if score > 0 {
			results = append(results, SearchResult{
				Bookmark:  *bm,
				Score:     score,
				MatchType: matchType,
			})
		}
	}

	// 按分数排序
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	// 限制结果数量
	if len(results) > e.config.MaxResults {
		results = results[:e.config.MaxResults]
	}

	return results
}

// calculateScore 计算匹配分数
func (e *SearchEngine) calculateScore(bm *Bookmark, query string) (int, string) {
	titleLower := strings.ToLower(bm.Title)
	urlLower := strings.ToLower(bm.URL)
	folderLower := strings.ToLower(bm.Folder)

	// 1. 标题前缀匹配（最高优先级）
	if strings.HasPrefix(titleLower, query) {
		score := 100 + (100-len(bm.Title))/10
		return score, "prefix"
	}

	// 2. 标题完全匹配
	if titleLower == query {
		return 95, "exact"
	}

	// 3. 标题包含匹配
	if idx := strings.Index(titleLower, query); idx >= 0 {
		score := 80 + (100-idx)/10
		return score, "contains"
	}

	// 4. URL 包含匹配
	if strings.Contains(urlLower, query) {
		return 60, "url"
	}

	// 5. 文件夹匹配
	if strings.Contains(folderLower, query) {
		return 40, "folder"
	}

	return 0, ""
}
