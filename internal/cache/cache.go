package cache

import (
	"sync"
	"time"
)

// Cache 缓存接口
type Cache[K comparable, V any] interface {
	Get(key K) (V, bool)
	Set(key K, value V, ttl time.Duration)
	Delete(key K)
	Clear()
	Size() int
}

// Entry 缓存条目
type Entry[V any] struct {
	Value     V
	ExpiresAt time.Time
}

// IsExpired 检查是否过期
func (e *Entry[V]) IsExpired() bool {
	return !e.ExpiresAt.IsZero() && time.Now().After(e.ExpiresAt)
}

// MemoryCache 内存缓存
type MemoryCache[K comparable, V any] struct {
	mu      sync.RWMutex
	items   map[K]*Entry[V]
 maxSize int
}

// NewMemoryCache 创建内存缓存
func NewMemoryCache[K comparable, V any](maxSize int) *MemoryCache[K, V] {
	if maxSize <= 0 {
		maxSize = 1000
	}
	return &MemoryCache[K, V]{
		items:   make(map[K]*Entry[V]),
		maxSize: maxSize,
	}
}

// Get 获取缓存
func (c *MemoryCache[K, V]) Get(key K) (V, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, ok := c.items[key]
	if !ok || entry.IsExpired() {
		var zero V
		return zero, false
	}
	return entry.Value, true
}

// Set 设置缓存
func (c *MemoryCache[K, V]) Set(key K, value V, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 检查是否需要清理
	if len(c.items) >= c.maxSize {
		c.evict()
	}

	var expiresAt time.Time
	if ttl > 0 {
		expiresAt = time.Now().Add(ttl)
	}

	c.items[key] = &Entry[V]{
		Value:     value,
		ExpiresAt: expiresAt,
	}
}

// Delete 删除缓存
func (c *MemoryCache[K, V]) Delete(key K) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, key)
}

// Clear 清空缓存
func (c *MemoryCache[K, V]) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make(map[K]*Entry[V])
}

// Size 返回缓存大小
func (c *MemoryCache[K, V]) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items)
}

// evict 清理过期条目
func (c *MemoryCache[K, V]) evict() {
	now := time.Now()
	for key, entry := range c.items {
		if entry.IsExpired() || now.After(entry.ExpiresAt) {
			delete(c.items, key)
		}
	}

	// 如果还是满的，删除最旧的
	if len(c.items) >= c.maxSize {
		// 简单策略：删除第一个
		for key := range c.items {
			delete(c.items, key)
			break
		}
	}
}

// LRUEntry LRU 缓存条目
type LRUEntry[V any] struct {
	Value      V
	ExpiresAt time.Time
	Key        string
}

// IsExpired 检查是否过期
func (e *LRUEntry[V]) IsExpired() bool {
	return !e.ExpiresAt.IsZero() && time.Now().After(e.ExpiresAt)
}

// LRUCache LRU 缓存
type LRUCache[K comparable, V any] struct {
	mu        sync.RWMutex
	items     map[K]*LRUEntry[V]
	order     []K
	maxSize   int
}

// NewLRUCache 创建 LRU 缓存
func NewLRUCache[K comparable, V any](maxSize int) *LRUCache[K, V] {
	if maxSize <= 0 {
		maxSize = 100
	}
	return &LRUCache[K, V]{
		items:   make(map[K]*LRUEntry[V]),
		order:   make([]K, 0),
		maxSize: maxSize,
	}
}

// Get 获取缓存
func (c *LRUCache[K, V]) Get(key K) (V, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry, ok := c.items[key]
	if !ok || entry.IsExpired() {
		var zero V
		return zero, false
	}

	// 移到最新
	c.moveToEnd(key)

	return entry.Value, true
}

// Set 设置缓存
func (c *LRUCache[K, V]) Set(key K, value V, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 如果已存在，更新
	if entry, ok := c.items[key]; ok {
		entry.Value = value
		if ttl > 0 {
			entry.ExpiresAt = time.Now().Add(ttl)
		}
		c.moveToEnd(key)
		return
	}

	// 检查容量
	if len(c.items) >= c.maxSize {
		c.evictOldest()
	}

	var expiresAt time.Time
	if ttl > 0 {
		expiresAt = time.Now().Add(ttl)
	}

	c.items[key] = &LRUEntry[V]{
		Value:     value,
		ExpiresAt: expiresAt,
	}
	c.order = append(c.order, key)
}

// Delete 删除缓存
func (c *LRUCache[K, V]) Delete(key K) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.items[key]; ok {
		delete(c.items, key)
		c.removeFromOrder(key)
	}
}

// Clear 清空缓存
func (c *LRUCache[K, V]) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make(map[K]*LRUEntry[V])
	c.order = make([]K, 0)
}

// Size 返回缓存大小
func (c *LRUCache[K, V]) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items)
}

// moveToEnd 移到末尾
func (c *LRUCache[K, V]) moveToEnd(key K) {
	c.removeFromOrder(key)
	c.order = append(c.order, key)
}

// removeFromOrder 从顺序中移除
func (c *LRUCache[K, V]) removeFromOrder(key K) {
	for i, k := range c.order {
		if k == key {
			c.order = append(c.order[:i], c.order[i+1:]...)
			return
		}
	}
}

// evictOldest 淘汰最旧的
func (c *LRUCache[K, V]) evictOldest() {
	if len(c.order) == 0 {
		return
	}
	oldest := c.order[0]
	delete(c.items, oldest)
	c.order = c.order[1:]
}

// CacheStats 缓存统计
type CacheStats struct {
	Hits      int64
	Misses    int64
	Evictions int64
	Size      int
}

// HitRate 命中率
func (s *CacheStats) HitRate() float64 {
	total := s.Hits + s.Misses
	if total == 0 {
		return 0
	}
	return float64(s.Hits) / float64(total)
}

// StatsCache 带统计的缓存
type StatsCache[K comparable, V any] struct {
	cache Cache[K, V]
	stats CacheStats
	mu    sync.RWMutex
}

// NewStatsCache 创建带统计的缓存
func NewStatsCache[K comparable, V any](cache Cache[K, V]) *StatsCache[K, V] {
	return &StatsCache[K, V]{
		cache: cache,
	}
}

// Get 获取缓存（记录统计）
func (c *StatsCache[K, V]) Get(key K) (V, bool) {
	value, ok := c.cache.Get(key)
	c.mu.Lock()
	defer c.mu.Unlock()
	if ok {
		c.stats.Hits++
	} else {
		c.stats.Misses++
	}
	return value, ok
}

// Set 设置缓存
func (c *StatsCache[K, V]) Set(key K, value V, ttl time.Duration) {
	c.cache.Set(key, value, ttl)
}

// Delete 删除缓存
func (c *StatsCache[K, V]) Delete(key K) {
	c.cache.Delete(key)
}

// Clear 清空缓存
func (c *StatsCache[K, V]) Clear() {
	c.cache.Clear()
}

// Size 返回缓存大小
func (c *StatsCache[K, V]) Size() int {
	return c.cache.Size()
}

// Stats 获取统计信息
func (c *StatsCache[K, V]) Stats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()
	stats := c.stats
	stats.Size = c.cache.Size()
	return stats
}
