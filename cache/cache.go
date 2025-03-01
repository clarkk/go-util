package cache

import (
	"time"
	"sync"
)

type (
	Cache[V any] struct {
		items		map[string]item[V]
		lock		sync.RWMutex
	}
	
	item[V any] struct {
		value		V
		expires		int64
	}
)

//	Create new cache
func New[V any](purge_interval int) *Cache[V] {
	c := &Cache[V]{
		items: map[string]item[V]{},
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
func (c *Cache[V]) Get(key string) (V, bool){
	c.lock.RLock()
	defer c.lock.RUnlock()
	i, ok := c.items[key]
	if !ok {
		return i.value, false
	}
	//	Check if value has expired
	if time_unix() > i.expires {
		return i.value, false
	}
	return i.value, true
}

//	Set value in cache
func (c *Cache[V]) Set(key string, value V, expires int){
	c.lock.Lock()
	defer c.lock.Unlock()
	c.items[key] = item[V]{
		value:		value,
		expires:	time_unix() + int64(expires),
	}
}

func (c *Cache[V]) purge_expired(){
	c.lock.Lock()
	defer c.lock.Unlock()
	time_unix := time_unix()
	for key, i := range c.items {
		if time_unix > i.expires {
			delete(c.items, key)
		}
	}
}

func time_unix() int64 {
	return time.Now().Unix()
}