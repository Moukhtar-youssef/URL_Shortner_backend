package Storage

import (
	"sync"
	"time"
)

type cacheItem struct {
	value  string
	expiry time.Time
}

type cache struct {
	data map[string]cacheItem
	mu   sync.RWMutex
}

func newCache() *cache {
	return &cache{
		data: make(map[string]cacheItem),
	}
}

func (c *cache) set(key string, value string, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data[key] = cacheItem{
		value:  value,
		expiry: time.Now().Add(ttl),
	}
}

func (c *cache) get(key string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, ok := c.data[key]
	if !ok {
		return "", false
	}

	if item.expiry.Before(time.Now()) {
		delete(c.data, key)
		return "", false
	}

	return item.value, true
}

func (c *cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.data, key)
}

func (c *cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data = make(map[string]cacheItem)
}
