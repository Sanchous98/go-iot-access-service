package cache

import (
	"sync"
	"time"
)

const defaultTtl = time.Hour

type cacheItem[T any] struct {
	item     T
	expireAt time.Time
}

type Cache[T any] struct {
	storage []cacheItem[T]
	mu      sync.RWMutex
}

func (c *Cache[T]) Get(condition func(T) bool) (item T, hit bool) {
	for index, value := range c.storage {
		if time.Now().After(value.expireAt) {
			c.mu.Lock()
			c.storage = append(c.storage[:index], c.storage[index:]...)
			c.mu.Unlock()
			continue
		}

		if condition(value.item) {
			item = value.item
			hit = true
			return
		}
	}

	return
}

func (c *Cache[T]) Put(items ...T) {
	c.mu.Lock()
	for _, item := range items {
		c.storage = append(c.storage, cacheItem[T]{
			item:     item,
			expireAt: time.Now().Add(defaultTtl),
		})
	}
	c.mu.Unlock()
}
