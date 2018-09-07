package utils

import (
	"container/list"
	"sync"
)

type Cache struct {
	size                   int
	evictList              *list.List
	items                  map[interface{}]*list.Element
	lock                   sync.RWMutex
	name                   string
	defaultExpiry          int64
	invalidateClusterEvent string
	currentGeneration      int64
	len                    int
}


func NewLru(size int) *Cache {
	return &Cache{
		size:      size,
		evictList: list.New(),
		items:     make(map[interface{}]*list.Element, size),
	}
}