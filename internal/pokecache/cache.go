package pokecache

import (
	"sync"
	"time"
)

type Cache struct {
	entryMap map[string]cacheEntry
	mu       sync.Mutex
	ttl      time.Duration
}

type cacheEntry struct {
	createdAt time.Time
	val       []byte
}

func NewCache(ttl time.Duration, cleanupInterval time.Duration) *Cache {
	cache := &Cache{
		entryMap: make(map[string]cacheEntry),
		ttl:      ttl,
	}

	go cache.reapLoop(cleanupInterval)

	return cache
}

func (c *Cache) Add(key string, val []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entryMap[key] = cacheEntry{
		createdAt: time.Now(),
		val:       val,
	}
}

func (c *Cache) Get(key string) ([]byte, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if entry, exists := c.entryMap[key]; exists {
		return entry.val, true
	} else {
		return nil, false
	}

}

func (c *Cache) reapLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()
		for key, entry := range c.entryMap {
			if time.Since(entry.createdAt) > c.ttl {
				delete(c.entryMap, key)
			}
		}
		c.mu.Unlock()
	}
}
