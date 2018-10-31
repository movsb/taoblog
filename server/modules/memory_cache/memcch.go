package memory_cache

import (
	"container/list"
	"log"
	"sync"
	"time"
)

type _Item struct {
	tim time.Time
	key string
	val interface{}
}

// MemoryCache is an in-memory LRU cache.
type MemoryCache struct {
	ttl  time.Duration
	lock sync.RWMutex
	keys map[string]*list.Element
	vals *list.List
	c    chan struct{}
}

// NewMemoryCache news a memory cache.
func NewMemoryCache(ttl time.Duration) *MemoryCache {
	m := &MemoryCache{
		ttl:  ttl,
		keys: make(map[string]*list.Element),
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
			log.Println("quit")
			return
		case <-time.After(m.ttl):
			log.Println("before collect")
			m.collect()
		}
	}
}

func (m *MemoryCache) collect() {
	m.lock.Lock()
	defer m.lock.Unlock()
	for m.vals.Len() > 0 {
		elem := m.vals.Back()
		item := elem.Value.(_Item)
		elapsed := time.Now().Sub(item.tim)
		if elapsed.Seconds() > m.ttl.Seconds() {
			m.vals.Remove(elem)
			delete(m.keys, item.key)
			log.Println("collected key: ", item.key, ", len: ", m.vals.Len())
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
		log.Println("move to front: ", key)
	} else {
		elem := m.vals.PushFront(_Item{
			tim: time.Now(),
			key: key,
			val: val,
		})
		m.keys[key] = elem
		log.Println("new key: ", key)
	}
}

// SetIf sets if
func (m *MemoryCache) SetIf(cond bool, key string, val interface{}) {
	if cond {
		m.Set(key, val)
	}
}

// Get gets
func (m *MemoryCache) Get(key string) (interface{}, bool) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	if elem, ok := m.keys[key]; ok {
		log.Println("hit cache: ", key)
		return elem.Value.(_Item).val, true
	}
	return nil, false
}

// Delete deletes
func (m *MemoryCache) Delete(key string) {
	m.lock.Lock()
	defer m.lock.Unlock()
	if elem, ok := m.keys[key]; ok {
		m.vals.Remove(elem)
		delete(m.keys, key)
		log.Println("delete key: ", key)
	}
}
