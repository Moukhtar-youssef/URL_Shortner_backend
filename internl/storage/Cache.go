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
	data        map[string]cacheItem
	mu          sync.RWMutex
	stopCleanup chan struct{}
}

func newCache() *cache {
	c := &cache{
		data:        make(map[string]cacheItem),
		mu:          sync.RWMutex{},
		stopCleanup: make(chan struct{}),
	}
	go c.startCleanup(30 * time.Second)
	return c
}

func (c *cache) Set(key string, value string, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data[key] = cacheItem{
		value:  value,
		expiry: time.Now().Add(ttl),
	}
}

func (c *cache) Get(key string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, ok := c.data[key]
	if !ok {
		return "", false
	}

	return item.value, true
}

func (c *cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.data, key)
}

func (c *cache) Stop() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data = make(map[string]cacheItem)
	close(c.stopCleanup)
}

func (c *cache) startCleanup(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.mu.Lock()
			now := time.Now()
			for k, v := range c.data {
				if v.expiry.Before(now) {
					delete(c.data, k)
				}
			}
			c.mu.Unlock()

		case <-c.stopCleanup:
			return
		}
	}
}
