package bookmark

import (
	"sort"
	"strings"

	"github.com/mozillazg/go-pinyin"
)

// SearchEngine 搜索引擎
type SearchEngine struct {
	bookmarks []Bookmark
	config    *BookmarkConfig
	pinyinArgs pinyin.Args
}

// NewSearchEngine 创建搜索引擎
func NewSearchEngine(config *BookmarkConfig) *SearchEngine {
	args := pinyin.NewArgs()
	args.Style = pinyin.Normal // 使用普通风格（不带声调）

	return &SearchEngine{
		config:     config,
		pinyinArgs: args,
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

	// 4. 拼音匹配（如果启用）
	if e.config.EnablePinyin {
		if score, matchType := e.matchPinyin(bm.Title, query); score > 0 {
			return score, matchType
		}
	}

	// 5. URL 包含匹配
	if strings.Contains(urlLower, query) {
		return 60, "url"
	}

	// 6. 文件夹匹配
	if strings.Contains(folderLower, query) {
		return 40, "folder"
	}

	return 0, ""
}

// matchPinyin 拼音匹配
func (e *SearchEngine) matchPinyin(title, query string) (int, string) {
	// 将标题转换为拼音
	pinyins := pinyin.Pinyin(title, e.pinyinArgs)
	if len(pinyins) == 0 {
		return 0, ""
	}

	// 构建拼音字符串（首字母和全拼）
	var fullPinyin, firstLetters strings.Builder
	for _, py := range pinyins {
		if len(py) > 0 {
			fullPinyin.WriteString(py[0])
			if len(py[0]) > 0 {
				firstLetters.WriteByte(py[0][0])
			}
		}
	}

	fullPinyinStr := strings.ToLower(fullPinyin.String())
	firstLettersStr := strings.ToLower(firstLetters.String())

	// 拼音前缀匹配
	if strings.HasPrefix(fullPinyinStr, query) {
		return 70, "pinyin"
	}

	// 首字母前缀匹配
	if strings.HasPrefix(firstLettersStr, query) {
		return 65, "pinyin_abbr"
	}

	// 拼音包含匹配
	if strings.Contains(fullPinyinStr, query) {
		return 50, "pinyin"
	}

	return 0, ""
}
