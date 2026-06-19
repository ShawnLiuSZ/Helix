package cache

import (
	"testing"
	"time"
)

func TestMemoryCache(t *testing.T) {
	t.Run("NewMemoryCache", func(t *testing.T) {
		cache := NewMemoryCache[string, int](100)
		if cache == nil {
			t.Fatal("expected non-nil cache")
		}
		if cache.Size() != 0 {
			t.Errorf("expected size 0, got %d", cache.Size())
		}
	})

	t.Run("SetAndGet", func(t *testing.T) {
		cache := NewMemoryCache[string, int](100)

		cache.Set("key1", 42, time.Minute)

		value, ok := cache.Get("key1")
		if !ok {
			t.Error("expected to find key1")
		}
		if value != 42 {
			t.Errorf("expected 42, got %d", value)
		}
	})

	t.Run("Get_NotFound", func(t *testing.T) {
		cache := NewMemoryCache[string, int](100)

		_, ok := cache.Get("nonexistent")
		if ok {
			t.Error("expected not to find nonexistent")
		}
	})

	t.Run("Get_Expired", func(t *testing.T) {
		cache := NewMemoryCache[string, int](100)

		cache.Set("key1", 42, 1*time.Millisecond)
		time.Sleep(10 * time.Millisecond)

		_, ok := cache.Get("key1")
		if ok {
			t.Error("expected key1 to be expired")
		}
	})

	t.Run("Delete", func(t *testing.T) {
		cache := NewMemoryCache[string, int](100)

		cache.Set("key1", 42, time.Minute)
		cache.Delete("key1")

		_, ok := cache.Get("key1")
		if ok {
			t.Error("expected key1 to be deleted")
		}
	})

	t.Run("Clear", func(t *testing.T) {
		cache := NewMemoryCache[string, int](100)

		cache.Set("key1", 42, time.Minute)
		cache.Set("key2", 43, time.Minute)
		cache.Clear()

		if cache.Size() != 0 {
			t.Errorf("expected size 0, got %d", cache.Size())
		}
	})

	t.Run("MaxSize", func(t *testing.T) {
		cache := NewMemoryCache[string, int](2)

		cache.Set("key1", 1, time.Minute)
		cache.Set("key2", 2, time.Minute)
		cache.Set("key3", 3, time.Minute) // 应该触发清理

		if cache.Size() > 2 {
			t.Errorf("expected size <= 2, got %d", cache.Size())
		}
	})
}

func TestLRUCache(t *testing.T) {
	t.Run("NewLRUCache", func(t *testing.T) {
		cache := NewLRUCache[string, int](100)
		if cache == nil {
			t.Fatal("expected non-nil cache")
		}
	})

	t.Run("SetAndGet", func(t *testing.T) {
		cache := NewLRUCache[string, int](100)

		cache.Set("key1", 42, time.Minute)

		value, ok := cache.Get("key1")
		if !ok {
			t.Error("expected to find key1")
		}
		if value != 42 {
			t.Errorf("expected 42, got %d", value)
		}
	})

	t.Run("LRU_Eviction", func(t *testing.T) {
		cache := NewLRUCache[string, int](2)

		cache.Set("key1", 1, time.Minute)
		cache.Set("key2", 2, time.Minute)
		cache.Get("key1") // 访问 key1，使其变新
		cache.Set("key3", 3, time.Minute) // 应该淘汰 key2

		_, ok := cache.Get("key2")
		if ok {
			t.Error("expected key2 to be evicted")
		}

		_, ok = cache.Get("key1")
		if !ok {
			t.Error("expected key1 to exist")
		}
	})

	t.Run("Delete", func(t *testing.T) {
		cache := NewLRUCache[string, int](100)

		cache.Set("key1", 42, time.Minute)
		cache.Delete("key1")

		_, ok := cache.Get("key1")
		if ok {
			t.Error("expected key1 to be deleted")
		}
	})

	t.Run("Clear", func(t *testing.T) {
		cache := NewLRUCache[string, int](100)

		cache.Set("key1", 42, time.Minute)
		cache.Set("key2", 43, time.Minute)
		cache.Clear()

		if cache.Size() != 0 {
			t.Errorf("expected size 0, got %d", cache.Size())
		}
	})

	t.Run("Update", func(t *testing.T) {
		cache := NewLRUCache[string, int](100)

		cache.Set("key1", 42, time.Minute)
		cache.Set("key1", 43, time.Minute)

		value, ok := cache.Get("key1")
		if !ok {
			t.Error("expected to find key1")
		}
		if value != 43 {
			t.Errorf("expected 43, got %d", value)
		}
	})
}

func TestStatsCache(t *testing.T) {
	t.Run("NewStatsCache", func(t *testing.T) {
		inner := NewMemoryCache[string, int](100)
		cache := NewStatsCache(inner)
		if cache == nil {
			t.Fatal("expected non-nil cache")
		}
	})

	t.Run("Stats", func(t *testing.T) {
		inner := NewMemoryCache[string, int](100)
		cache := NewStatsCache(inner)

		cache.Set("key1", 42, time.Minute)
		cache.Get("key1")
		cache.Get("key2") // miss

		stats := cache.Stats()
		if stats.Hits != 1 {
			t.Errorf("expected 1 hit, got %d", stats.Hits)
		}
		if stats.Misses != 1 {
			t.Errorf("expected 1 miss, got %d", stats.Misses)
		}
	})

	t.Run("HitRate", func(t *testing.T) {
		inner := NewMemoryCache[string, int](100)
		cache := NewStatsCache(inner)

		cache.Set("key1", 42, time.Minute)
		cache.Get("key1")
		cache.Get("key1")
		cache.Get("key2") // miss

		stats := cache.Stats()
		rate := stats.HitRate()
		if rate < 0.6 || rate > 0.7 {
			t.Errorf("expected hit rate ~0.66, got %f", rate)
		}
	})
}

func TestCacheStats(t *testing.T) {
	stats := &CacheStats{
		Hits:   10,
		Misses: 5,
	}

	rate := stats.HitRate()
	if rate != 0.6666666666666666 {
		t.Errorf("expected 0.666..., got %f", rate)
	}

	// 测试零值
	zeroStats := &CacheStats{}
	if zeroStats.HitRate() != 0 {
		t.Errorf("expected 0, got %f", zeroStats.HitRate())
	}
}

func TestLRUCacheGet_Expired(t *testing.T) {
	cache := NewLRUCache[string, int](100)

	cache.Set("key1", 42, 1*time.Millisecond)
	time.Sleep(10 * time.Millisecond)

	_, ok := cache.Get("key1")
	if ok {
		t.Error("expected key1 to be expired")
	}
}

func TestMemoryCacheConcurrent(t *testing.T) {
	cache := NewMemoryCache[string, int](100)

	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				cache.Set("key", j, time.Minute)
				cache.Get("key")
			}
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}
