package memory_cache

import (
	"container/list"
	"fmt"
	"log"
	"sync"
	"time"
)

func debug(f string, a ...interface{}) {
	s := fmt.Sprintf(f, a...)
	log.Println("memory_cache:", s)
}

type _Item struct {
	tim time.Time
	key interface{}
	val interface{}
}

// MemoryCache is an in-memory LRU cache.
type MemoryCache struct {
	ttl  time.Duration
	lock sync.RWMutex
	keys map[interface{}]*list.Element
	vals *list.List
	c    chan struct{}
}

// NewMemoryCache news a memory cache.
func NewMemoryCache(ttl time.Duration) *MemoryCache {
	m := &MemoryCache{
		ttl:  ttl,
		keys: make(map[interface{}]*list.Element),
		vals: list.New(),
		c:    make(chan struct{}),
	}
	go m.checkTTL()
	return m
}

func (m *MemoryCache) checkTTL() {
	for {
		select {
		case <-m.c:
			debug("quit")
			return
		case <-time.After(m.ttl):
			debug("before collect")
			m.collect()
			debug("after collect")
		}
	}
}

func (m *MemoryCache) collect() {
	m.lock.Lock()
	defer m.lock.Unlock()
	for m.vals.Len() > 0 {
		elem := m.vals.Back()
		item := elem.Value.(_Item)
		elapsed := time.Since(item.tim)
		if elapsed.Seconds() > m.ttl.Seconds() {
			m.vals.Remove(elem)
			delete(m.keys, item.key)
			debug("collected key: %s, len: %d", item.key, m.vals.Len())
		} else {
			break
		}
	}
}

// Stop stops
func (m *MemoryCache) Stop() {
	m.c <- struct{}{}
	close(m.c)
}

// Set sets
func (m *MemoryCache) Set(key string, val interface{}) {
	m.lock.Lock()
	defer m.lock.Unlock()
	if elem, ok := m.keys[key]; ok {
		elem.Value = _Item{
			tim: time.Now(),
			key: key,
			val: val,
		}
		m.vals.MoveToFront(elem)
		debug("move to front: %s", key)
	} else {
		elem := m.vals.PushFront(_Item{
			tim: time.Now(),
			key: key,
			val: val,
		})
		m.keys[key] = elem
		debug("new key: %s", key)
	}
}

// SetIf sets if
func (m *MemoryCache) SetIf(cond bool, key string, val interface{}) {
	if cond {
		m.Set(key, val)
	}
}

// Get gets
func (m *MemoryCache) Get(key string, loader func(key string) (interface{}, error)) (interface{}, error) {
	m.lock.RLock()
	if elem, ok := m.keys[key]; ok {
		debug("hit cache: %s", key)
		m.lock.RUnlock()
		return elem.Value.(_Item).val, nil
	}
	m.lock.RUnlock()
	m.lock.Lock()
	defer m.lock.Unlock()
	if elem, ok := m.keys[key]; ok {
		debug("hit cache: %s", key)
		return elem.Value.(_Item).val, nil
	}
	val, err := loader(key)
	if err != nil {
		return nil, err
	}
	elem := m.vals.PushFront(_Item{
		tim: time.Now(),
		key: key,
		val: val,
	})
	m.keys[key] = elem
	debug("new key: %s", key)
	return val, nil
}

// Delete deletes
func (m *MemoryCache) Delete(key string) {
	m.lock.Lock()
	defer m.lock.Unlock()
	if elem, ok := m.keys[key]; ok {
		m.vals.Remove(elem)
		delete(m.keys, key)
		debug("delete key: %s", key)
	}
}
