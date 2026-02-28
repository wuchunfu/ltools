package browser

// BookmarkParser 浏览器书签解析器接口
type BookmarkParser interface {
	// Name 返回浏览器名称
	Name() string

	// Parse 解析书签
	Parse() ([]Bookmark, error)

	// IsAvailable 检查浏览器是否可用（书签文件是否存在）
	IsAvailable() bool

	// GetBookmarksPath 获取书签文件路径
	GetBookmarksPath() (string, error)
}
