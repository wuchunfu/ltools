package musicplayer

import (
	"context"
	"fmt"
	"log"
	"sort"
	"sync"
	"time"
)

// SourceManager 音源管理器
type SourceManager struct {
	sources []*MusicSource
	stats   map[string]*SourceStats
	mutex   sync.RWMutex

	// 优先级顺序
	priorityOrder []string
}

// MusicSource 音源
type MusicSource struct {
	ID      string
	Name    string
	Enabled bool
	Priority int
}

// SourceStats 音源统计
type SourceStats struct {
	TotalRequests   int64
	SuccessRequests int64
	FailedRequests  int64
	AvgLatency      time.Duration
	LastFailTime    time.Time
	LastSuccessTime time.Time
}

// NewSourceManager 创建源管理器
func NewSourceManager() *SourceManager {
	// 默认音源（按优先级排序）
	// 注意：ID 必须与 LX Music 服务中实际加载的插件 platform 字段匹配
	sources := []*MusicSource{
		{ID: "元力KW", Name: "酷我音乐（元力）", Enabled: true, Priority: 1},
		{ID: "开心汽水", Name: "开心汽水", Enabled: true, Priority: 2},
	}

	stats := make(map[string]*SourceStats)
	priorityOrder := make([]string, len(sources))

	for i, source := range sources {
		stats[source.ID] = &SourceStats{}
		priorityOrder[i] = source.ID
	}

	return &SourceManager{
		sources:       sources,
		stats:         stats,
		priorityOrder: priorityOrder,
	}
}

// GetSources 获取所有音源
func (sm *SourceManager) GetSources() []*MusicSource {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	result := make([]*MusicSource, len(sm.sources))
	copy(result, sm.sources)
	return result
}

// GetEnabledSources 获取启用的音源
func (sm *SourceManager) GetEnabledSources() []string {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	result := make([]string, 0)
	for _, id := range sm.priorityOrder {
		for _, source := range sm.sources {
			if source.ID == id && source.Enabled {
				result = append(result, id)
				break
			}
		}
	}
	return result
}

// GetSourceStats 获取音源统计
func (sm *SourceManager) GetSourceStats(sourceID string) *SourceStats {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	return sm.stats[sourceID]
}

// RecordSuccess 记录成功
func (sm *SourceManager) RecordSuccess(sourceID string, latency time.Duration) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	if stats, ok := sm.stats[sourceID]; ok {
		stats.TotalRequests++
		stats.SuccessRequests++
		stats.AvgLatency = (stats.AvgLatency*time.Duration(stats.TotalRequests-1) + latency) / time.Duration(stats.TotalRequests)
		stats.LastSuccessTime = time.Now()

		// 动态调整优先级
		sm.updatePriority()
	}
}

// RecordFailure 记录失败
func (sm *SourceManager) RecordFailure(sourceID string) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	if stats, ok := sm.stats[sourceID]; ok {
		stats.TotalRequests++
		stats.FailedRequests++
		stats.LastFailTime = time.Now()

		// 动态调整优先级
		sm.updatePriority()
	}
}

// GetSuccessRate 获取成功率
func (sm *SourceManager) GetSuccessRate(sourceID string) float64 {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	stats := sm.stats[sourceID]
	if stats.TotalRequests == 0 {
		return 1.0 // 默认 100%
	}

	return float64(stats.SuccessRequests) / float64(stats.TotalRequests)
}

// updatePriority 动态调整优先级（根据成功率）
func (sm *SourceManager) updatePriority() {
	// 按成功率排序
	sort.Slice(sm.priorityOrder, func(i, j int) bool {
		rateI := sm.GetSuccessRate(sm.priorityOrder[i])
		rateJ := sm.GetSuccessRate(sm.priorityOrder[j])
		return rateI > rateJ // 降序
	})

	log.Printf("[SourceManager] Priority updated: %v", sm.priorityOrder)
}

// GetMusicURLWithFallback 获取播放 URL（多源回退）
func (sm *SourceManager) GetMusicURLWithFallback(
	ctx context.Context,
	lxClient *LXClient,
	cacheManager *CacheManager,
	songID, songName, singer string,
	duration int,
	quality string,
) ([]SongURLOption, error) {

	// 获取启用的音源
	sources := sm.GetEnabledSources()
	if len(sources) == 0 {
		return nil, fmt.Errorf("no enabled sources")
	}

	// 生成缓存键
	cacheKey := GenerateCacheKey("music_url", songID, quality)

	// 查询缓存
	if cached, ok := cacheManager.Get(cacheKey); ok {
		if urls, ok := cached.([]SongURLOption); ok {
			log.Printf("[SourceManager] Cache hit for song %s", songID)
			return urls, nil
		}
	}

	// 批量获取 URL（并发请求多个源）
	result, err := lxClient.GetMusicURLBatch(ctx, songID, songName, singer, duration, sources, quality)
	if err != nil {
		return nil, err
	}

	if len(result.URLs) == 0 {
		return nil, fmt.Errorf("no available URL from any source")
	}

	// 记录成功
	for _, url := range result.URLs {
		sm.RecordSuccess(url.Source, 0) // latency 由 LXClient 内部记录
	}

	// 缓存结果
	cacheManager.Set(cacheKey, result.URLs, 30*time.Minute)

	return result.URLs, nil
}

// GetBestSource 获取最佳音源（根据成功率）
func (sm *SourceManager) GetBestSource() string {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	if len(sm.priorityOrder) > 0 {
		return sm.priorityOrder[0]
	}
	return ""
}

// EnableSource 启用音源
func (sm *SourceManager) EnableSource(sourceID string) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	for _, source := range sm.sources {
		if source.ID == sourceID {
			source.Enabled = true
			break
		}
	}
}

// DisableSource 禁用音源
func (sm *SourceManager) DisableSource(sourceID string) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	for _, source := range sm.sources {
		if source.ID == sourceID {
			source.Enabled = false
			break
		}
	}
}

// GetStatsSummary 获取统计摘要
func (sm *SourceManager) GetStatsSummary() map[string]interface{} {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	summary := make(map[string]interface{})
	for id, stats := range sm.stats {
		summary[id] = map[string]interface{}{
			"total_requests":   stats.TotalRequests,
			"success_requests": stats.SuccessRequests,
			"failed_requests":  stats.FailedRequests,
			"success_rate":     sm.GetSuccessRate(id),
			"avg_latency":      stats.AvgLatency.String(),
		}
	}
	return summary
}
