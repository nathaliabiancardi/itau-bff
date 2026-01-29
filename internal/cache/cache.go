package cache

import (
	"sync"
	"time"
)

type Item[T any] struct {
	Value      T
	Expiration time.Time
}

type Cache[T any] struct {
	mu    sync.RWMutex
	items map[string]Item[T]
	ttl   time.Duration
}

func NewCache[T any](ttl time.Duration) *Cache[T] {
	return &Cache[T]{
		items: make(map[string]Item[T]),
		ttl:   ttl,
	}
}

func (c *Cache[T]) Get(key string) (T, bool) {
	c.mu.RLock()
	item, found := c.items[key]
	c.mu.RUnlock()

	var zero T
	if !found {
		return zero, false
	}

	if time.Now().After(item.Expiration) {
		c.mu.Lock()
		delete(c.items, key)
		c.mu.Unlock()
		return zero, false
	}

	return item.Value, true
}

func (c *Cache[T]) Set(key string, value T) {
	c.mu.Lock()
	c.items[key] = Item[T]{
		Value:      value,
		Expiration: time.Now().Add(c.ttl),
	}
	c.mu.Unlock()
}
