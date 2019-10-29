package cache

import (
	"fmt"
	"log"

	cache "github.com/patrickmn/go-cache"
)

type MemoryCache struct {
	cache  *cache.Cache
	logger *log.Logger
}

func (m *MemoryCache) Connect(logger *log.Logger) {
	m.cache = cache.New(-1, -1)
	m.logger = logger
}

func (m *MemoryCache) SetItem(key string, item interface{}) error {
	m.cache.Set(key, item, -1)
	m.saveLastKeys(key)
	// return error, if Set method of some library return error...
	return nil
}

func (m *MemoryCache) saveLastKeys(key string) {
	if _, ok := m.cache.Get("first"); !ok {
		m.cache.Set("first", key, -1)
	} else {
		if oldSecondKey, ok := m.cache.Get("second"); ok {
			m.cache.Set("first", oldSecondKey, -1)
		}
		m.cache.Set("second", key, -1)
	}
}

func (m *MemoryCache) getLastKeys() []string {
	var keys []string
	key, ok := m.cache.Get("first")
	if ok {
		keys = append(keys, key.(string))
	}
	key, ok = m.cache.Get("second")
	if ok {
		keys = append(keys, key.(string))
	}
	return keys
}

func (m *MemoryCache) GetItems() ([]string, error) {
	if m.cache.ItemCount() >= 2 {
		if keys := m.getLastKeys(); keys != nil && len(keys) == 2 {
			var values []string
			value, ok := m.cache.Get(keys[0])
			if ok {
				values = append(values, value.(string))
			}
			value, ok = m.cache.Get(keys[1])
			if ok {
				values = append(values, value.(string))
			}
			return values, nil
		}
		return nil, fmt.Errorf("not found 'first' or 'second' key")
	}
	return nil, fmt.Errorf("random strings list is empty or small")
}
