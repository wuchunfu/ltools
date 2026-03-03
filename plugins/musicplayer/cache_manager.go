package musicplayer

import (
	"container/list"
	"encoding/json"
	"log"
	"sync"
	"time"
)

// CacheManager 缓存管理器（内存 LRU + 磁盘持久化）
type CacheManager struct {
	// 内存缓存（LRU）
	memoryCache *LRUCache
	memoryMutex sync.RWMutex

	// 磁盘缓存（简化版本，使用 JSON 文件）
	// 后续可以升级为 BoltDB
	diskCache   map[string]*DiskCacheItem
	diskMutex   sync.RWMutex
	diskPath    string

	// 统计
	hits   int64
	misses int64
}

// CacheItem 缓存项
type CacheItem struct {
	Key       string
	Value     interface{}
	ExpiresAt time.Time
}

// DiskCacheItem 磁盘缓存项
type DiskCacheItem struct {
	Key       string          `json:"key"`
	Value     json.RawMessage `json:"value"`
	ExpiresAt time.Time       `json:"expires_at"`
}

// LRUCache LRU 缓存
type LRUCache struct {
	capacity int
	cache    map[string]*list.Element
	lruList  *list.List
}

// lruEntry LRU 列表条目
type lruEntry struct {
	key   string
	value *CacheItem
}

// CacheConfig 缓存配置
type CacheConfig struct {
	MemoryCapacity int           // 内存缓存容量
	DiskPath       string        // 磁盘缓存路径
	DefaultTTL     time.Duration // 默认 TTL
}

// NewCacheManager 创建缓存管理器
func NewCacheManager(config *CacheConfig) (*CacheManager, error) {
	// 默认配置
	if config.MemoryCapacity == 0 {
		config.MemoryCapacity = 1000
	}
	if config.DiskPath == "" {
		config.DiskPath = "/tmp/ltools/music-cache"
	}
	if config.DefaultTTL == 0 {
		config.DefaultTTL = 30 * time.Minute
	}

	cm := &CacheManager{
		memoryCache: NewLRUCache(config.MemoryCapacity),
		diskCache:   make(map[string]*DiskCacheItem),
		diskPath:    config.DiskPath,
	}

	// 加载磁盘缓存
	if err := cm.loadDiskCache(); err != nil {
		log.Printf("[CacheManager] Failed to load disk cache: %v", err)
	}

	return cm, nil
}

// Get 获取缓存
func (cm *CacheManager) Get(key string) (interface{}, bool) {
	// 先查内存缓存
	cm.memoryMutex.RLock()
	if item, ok := cm.memoryCache.Get(key); ok {
		cm.memoryMutex.RUnlock()
		cm.hits++
		return item, true
	}
	cm.memoryMutex.RUnlock()

	// 再查磁盘缓存
	cm.diskMutex.RLock()
	if item, ok := cm.diskCache[key]; ok {
		// 检查是否过期
		if time.Now().Before(item.ExpiresAt) {
			cm.diskMutex.RUnlock()

			// 反序列化值
			var value interface{}
			if err := json.Unmarshal(item.Value, &value); err == nil {
				// 回填到内存缓存
				cm.memoryMutex.Lock()
				cm.memoryCache.Set(key, value, time.Until(item.ExpiresAt))
				cm.memoryMutex.Unlock()

				cm.hits++
				return value, true
			}
		}
	}
	cm.diskMutex.RUnlock()

	cm.misses++
	return nil, false
}

// Set 设置缓存
func (cm *CacheManager) Set(key string, value interface{}, ttl time.Duration) {
	if ttl == 0 {
		ttl = 30 * time.Minute
	}

	// 设置内存缓存
	cm.memoryMutex.Lock()
	cm.memoryCache.Set(key, value, ttl)
	cm.memoryMutex.Unlock()

	// 设置磁盘缓存
	valueBytes, err := json.Marshal(value)
	if err == nil {
		cm.diskMutex.Lock()
		cm.diskCache[key] = &DiskCacheItem{
			Key:       key,
			Value:     valueBytes,
			ExpiresAt: time.Now().Add(ttl),
		}
		cm.diskMutex.Unlock()
	}
}

// Delete 删除缓存
func (cm *CacheManager) Delete(key string) {
	cm.memoryMutex.Lock()
	cm.memoryCache.Delete(key)
	cm.memoryMutex.Unlock()

	cm.diskMutex.Lock()
	delete(cm.diskCache, key)
	cm.diskMutex.Unlock()
}

// Clear 清空缓存
func (cm *CacheManager) Clear() {
	cm.memoryMutex.Lock()
	cm.memoryCache.Clear()
	cm.memoryMutex.Unlock()

	cm.diskMutex.Lock()
	cm.diskCache = make(map[string]*DiskCacheItem)
	cm.diskMutex.Unlock()
}

// GetHitRate 获取缓存命中率
func (cm *CacheManager) GetHitRate() float64 {
	total := cm.hits + cm.misses
	if total == 0 {
		return 0
	}
	return float64(cm.hits) / float64(total)
}

// GetStats 获取统计信息
func (cm *CacheManager) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"hits":       cm.hits,
		"misses":     cm.misses,
		"hit_rate":   cm.GetHitRate(),
		"memory_size": cm.memoryCache.Size(),
		"disk_size":   len(cm.diskCache),
	}
}

// loadDiskCache 加载磁盘缓存
func (cm *CacheManager) loadDiskCache() error {
	// 简化版本：从 JSON 文件加载
	// TODO: 实现实际的文件加载逻辑
	return nil
}

// saveDiskCache 保存磁盘缓存
func (cm *CacheManager) saveDiskCache() error {
	// 简化版本：保存到 JSON 文件
	// TODO: 实现实际的文件保存逻辑
	return nil
}

// ===== LRU 缓存实现 =====

// NewLRUCache 创建 LRU 缓存
func NewLRUCache(capacity int) *LRUCache {
	return &LRUCache{
		capacity: capacity,
		cache:    make(map[string]*list.Element),
		lruList:  list.New(),
	}
}

// Get 获取缓存
func (lru *LRUCache) Get(key string) (interface{}, bool) {
	if elem, ok := lru.cache[key]; ok {
		entry := elem.Value.(*lruEntry)
		// 检查是否过期
		if time.Now().Before(entry.value.ExpiresAt) {
			// 移动到前面（最近使用）
			lru.lruList.MoveToFront(elem)
			return entry.value.Value, true
		}
		// 已过期，删除
		lru.deleteElement(elem)
	}
	return nil, false
}

// Set 设置缓存
func (lru *LRUCache) Set(key string, value interface{}, ttl time.Duration) {
	// 如果已存在，更新
	if elem, ok := lru.cache[key]; ok {
		entry := elem.Value.(*lruEntry)
		entry.value.Value = value
		entry.value.ExpiresAt = time.Now().Add(ttl)
		lru.lruList.MoveToFront(elem)
		return
	}

	// 创建新条目
	entry := &lruEntry{
		key: key,
		value: &CacheItem{
			Key:       key,
			Value:     value,
			ExpiresAt: time.Now().Add(ttl),
		},
	}

	// 添加到列表前面
	elem := lru.lruList.PushFront(entry)
	lru.cache[key] = elem

	// 检查容量，淘汰旧条目
	for lru.lruList.Len() > lru.capacity {
		// 删除最旧的（列表末尾）
		oldest := lru.lruList.Back()
		if oldest != nil {
			lru.deleteElement(oldest)
		}
	}
}

// Delete 删除缓存
func (lru *LRUCache) Delete(key string) {
	if elem, ok := lru.cache[key]; ok {
		lru.deleteElement(elem)
	}
}

// Clear 清空缓存
func (lru *LRUCache) Clear() {
	lru.cache = make(map[string]*list.Element)
	lru.lruList.Init()
}

// Size 获取缓存大小
func (lru *LRUCache) Size() int {
	return lru.lruList.Len()
}

// deleteElement 删除元素
func (lru *LRUCache) deleteElement(elem *list.Element) {
	entry := elem.Value.(*lruEntry)
	delete(lru.cache, entry.key)
	lru.lruList.Remove(elem)
}

// ===== 辅助函数 =====

// GenerateCacheKey 生成缓存键
func GenerateCacheKey(prefix string, parts ...string) string {
	key := prefix
	for _, part := range parts {
		key += ":" + part
	}
	return key
}
