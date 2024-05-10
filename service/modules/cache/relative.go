package cache

import "sync"

type RelativeCacheKeys[PrimaryKey, SecondStageKey comparable] struct {
	lock sync.RWMutex
	keys map[PrimaryKey][]SecondStageKey
}

func NewRelativeCacheKeys[PrimaryKey, SecondStageKey comparable]() *RelativeCacheKeys[PrimaryKey, SecondStageKey] {
	return &RelativeCacheKeys[PrimaryKey, SecondStageKey]{
		keys: make(map[PrimaryKey][]SecondStageKey),
	}
}

func (c *RelativeCacheKeys[PrimaryKey, SecondStageKey]) Append(key PrimaryKey, second SecondStageKey) {
	c.lock.Lock()
	c.keys[key] = append(c.keys[key], second)
	c.lock.Unlock()
}

func (c *RelativeCacheKeys[PrimaryKey, SecondStageKey]) _RangeLocked(key PrimaryKey, fn func(second SecondStageKey)) {
	for _, s := range c.keys[key] {
		fn(s)
	}
}

func (c *RelativeCacheKeys[PrimaryKey, SecondStageKey]) Delete(key PrimaryKey, fn func(second SecondStageKey)) {
	c.lock.Lock()
	if fn != nil {
		c._RangeLocked(key, fn)
	}
	delete(c.keys, key)
	c.lock.Unlock()
}
