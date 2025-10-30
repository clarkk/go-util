package cache

import (
	"time"
	"sync"
)

type (
	Cache[K comparable, V any] struct {
		lock		sync.RWMutex
		items		map[K]cache_item[V]
	}
	
	cache_item[V any] struct {
		value		V
		expires		int64
	}
)

//	Create new cache
func NewCache[K comparable, V any](purge_interval int) *Cache[K, V] {
	c := &Cache[K, V]{
		items: map[K]cache_item[V]{},
	}
	//	Purge expired values from cache with time interval
	ticker := time.NewTicker(time.Duration(purge_interval) * time.Second)
	go func(){
		for range ticker.C {
			go c.purge_expired()
		}
	}()
	return c
}

//	Get cached value
func (c *Cache[K, V]) Get(key K) (V, bool){
	c.lock.RLock()
	defer c.lock.RUnlock()
	i, found := c.items[key]
	if !found {
		return i.value, false
	}
	//	Check if value has expired
	if i.expires != 0 && time_unix() > i.expires {
		return i.value, false
	}
	return i.value, true
}

//	Set value in cache
func (c *Cache[K, V]) Set(key K, value V, ttl int){
	c.lock.Lock()
	defer c.lock.Unlock()
	c.items[key] = cache_item[V]{
		value:		value,
		expires:	time_expires(ttl),
	}
}

func (c *Cache[K, V]) purge_expired(){
	c.lock.Lock()
	defer c.lock.Unlock()
	time_unix := time_unix()
	for key, i := range c.items {
		if i.expires != 0 && time_unix > i.expires {
			delete(c.items, key)
		}
	}
}