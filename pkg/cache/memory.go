package cache

import (
	"sync"
	"time"
)

type Cache interface {
	Get(key string) (interface{}, bool)
	Set(key string, value interface{})
	SetWithTTL(key string, value interface{}, ttl time.Duration)
	Delete(key string)
	Clear()
	Size() int
}

type MemoryCache struct {
	data       map[string]cacheItem
	mu         sync.RWMutex
	defaultTTL time.Duration
}

type cacheItem struct {
	value     interface{}
	expiresAt time.Time
}

func NewMemoryCache(defaultTTL time.Duration) *MemoryCache {
	cache := &MemoryCache{
		data:       make(map[string]cacheItem),
		defaultTTL: defaultTTL,
	}
	go cache.cleanup()
	return cache
}

func (c *MemoryCache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, exists := c.data[key]
	if !exists || time.Now().After(item.expiresAt) {
		return nil, false
	}
	return item.value, true
}

func (c *MemoryCache) Set(key string, value interface{}) {
	c.SetWithTTL(key, value, c.defaultTTL)
}

func (c *MemoryCache) SetWithTTL(key string, value interface{}, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data[key] = cacheItem{
		value:     value,
		expiresAt: time.Now().Add(ttl),
	}
}

func (c *MemoryCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.data, key)
}

func (c *MemoryCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data = make(map[string]cacheItem)
}

func (c *MemoryCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.data)
}

func (c *MemoryCache) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()
		now := time.Now()
		for key, item := range c.data {
			if now.After(item.expiresAt) {
				delete(c.data, key)
			}
		}
		c.mu.Unlock()
	}
}

func (c *MemoryCache) GetKeys() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var keys []string
	now := time.Now()
	for key, item := range c.data {
		if !now.After(item.expiresAt) {
			keys = append(keys, key)
		}
	}
	return keys
}

func (c *MemoryCache) GetStats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	now := time.Now()
	expired := 0
	for _, item := range c.data {
		if now.After(item.expiresAt) {
			expired++
		}
	}

	return CacheStats{
		TotalItems:   len(c.data),
		ExpiredItems: expired,
		ActiveItems:  len(c.data) - expired,
	}
}

type CacheStats struct {
	TotalItems   int
	ExpiredItems int
	ActiveItems  int
}
