package cache

import "log"

type CacheService interface {
	Connect(logger *log.Logger)
	SetItem(key string, item interface{}) error
	GetItems() ([]string, error)
}
