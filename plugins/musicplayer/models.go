package musicplayer

import (
	"fmt"
	"strings"
	"time"
)

// Song 歌曲信息
type Song struct {
	ID       string   `json:"id"`        // 歌曲 ID
	Name     string   `json:"name"`      // 歌曲名
	Artist   []string `json:"artist"`    // 歌手列表
	Album    string   `json:"album"`     // 专辑名
	PicID    string   `json:"pic_id"`    // 封面 ID
	URLID    string   `json:"url_id"`    // 播放 URL ID
	LyricID  string   `json:"lyric_id"`  // 歌词 ID
	Source   string   `json:"source"`    // 来源平台 (netease/tencent/kugou等)
	Duration int      `json:"duration"`  // 时长（秒）
}

// SongURL 播放地址信息
type SongURL struct {
	URL     string `json:"url"`      // 播放地址
	Size    int64  `json:"size"`     // 文件大小
	Bitrate int    `json:"bitrate"`  // 码率
}

// Config 播放器配置
type Config struct {
	Platform string `json:"platform"`  // 当前平台 (netease/tencent/kugou等)
	Volume   int    `json:"volume"`    // 音量 (0-100)
}

// LikeList 喜欢列表
type LikeList struct {
	Songs     []Song    `json:"songs"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Album 专辑信息
type Album struct {
	ID     string   `json:"id"`
	Name   string   `json:"name"`
	Artist []string `json:"artist"`
	Pic    string   `json:"pic"`
}

// Artist 歌手信息
type Artist struct {
	ID     string   `json:"id"`
	Name   string   `json:"name"`
	Pic    string   `json:"pic"`
	Albums []Album  `json:"albums"`
}

// ===== LX Music 格式转换 =====

// ConvertLXSongToInternal 将 LX Music 格式转换为内部格式
func ConvertLXSongToInternal(lxSong LXSong) Song {
	return Song{
		ID:       lxSong.ID,
		Name:     lxSong.Name,
		Artist:   parseSingers(lxSong.Singer),
		Album:    lxSong.Meta.AlbumName,
		PicID:    lxSong.Meta.PicURL, // 使用封面 URL 而不是 SongID
		URLID:    lxSong.Meta.SongID,
		LyricID:  lxSong.Meta.SongID,
		Source:   mapLXSourceToInternal(lxSong.Source),
		Duration: parseInterval(lxSong.Interval),
	}
}

// parseSingers 解析歌手字符串
func parseSingers(singer string) []string {
	// 常见分隔符：、, &, /
	separators := []string{"、", ",", "&", "/"}

	result := []string{singer}
	for _, sep := range separators {
		var newResult []string
		for _, s := range result {
			parts := strings.Split(s, sep)
			for _, part := range parts {
				trimmed := strings.TrimSpace(part)
				if trimmed != "" {
					newResult = append(newResult, trimmed)
				}
			}
		}
		result = newResult
	}

	return result
}

// parseInterval 解析时长字符串（"03:45" -> 225 秒）
func parseInterval(interval string) int {
	parts := strings.Split(interval, ":")
	if len(parts) != 2 {
		return 0
	}

	minutes := 0
	seconds := 0

	fmt.Sscanf(parts[0], "%d", &minutes)
	fmt.Sscanf(parts[1], "%d", &seconds)

	return minutes*60 + seconds
}

// mapLXSourceToInternal 映射 LX 源标识到内部标识
func mapLXSourceToInternal(lxSource string) string {
	mapping := map[string]string{
		"kw": "kuwo",     // 酷我
		"kg": "kugou",    // 酷狗
		"tx": "tencent",  // QQ音乐
		"wy": "netease",  // 网易云
		"mg": "migu",     // 咪咕
	}

	if internal, ok := mapping[lxSource]; ok {
		return internal
	}

	return lxSource
}

// mapInternalSourceToLX 映射内部标识到 LX 源标识
func mapInternalSourceToLX(internalSource string) string {
	mapping := map[string]string{
		"kuwo":    "kw",
		"kugou":   "kg",
		"tencent": "tx",
		"netease": "wy",
		"migu":    "mg",
	}

	if lx, ok := mapping[internalSource]; ok {
		return lx
	}

	return internalSource
}
