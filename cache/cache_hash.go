package cache

import (
	"sync"
	"time"
)

type (
	Refresher interface {
		Hash() string
	}
	
	Hash[K comparable, V Refresher] struct {
		lock		sync.RWMutex
		items		map[K]hash_item[V]
		ttl			int
		verify		func(key K, hash string) (bool, error)
		refresh		func(key K) (V, string, error)
	}
	
	hash_item[V Refresher] struct {
		value		V
		hash		string
		expires		int64
	}
)

//	Create new hash cache
func NewHash[K comparable, V Refresher](
	ttl int,
	verify func(key K, hash string) (bool, error),
	refresh func(key K) (V, string, error),
	purge_interval int,
) *Hash[K, V] {
	c := &Hash[K, V]{
		items:		map[K]hash_item[V]{},
		ttl:		ttl,
		verify:		verify,
		refresh:	refresh,
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
func (c *Hash[K, V]) Get(key K) (V, error){
	c.lock.RLock()
	item, found := c.items[key]
	//	Cache hit
	if found && time_unix() <= item.expires {
		//	Verify hash
		verified, err := c.verify(key, item.hash)
		if err != nil {
			c.lock.RUnlock()
			var zero V
			return zero, err
		}
		if verified {
			c.lock.RUnlock()
			return item.value, nil
		}
	}
	c.lock.RUnlock()
	
	return c.Refresh(key)
}

//	Refresh value in cache
func (c *Hash[K, V]) Refresh(key K) (V, error){
	c.lock.Lock()
	defer c.lock.Unlock()
	
	value, err := c.refresh(key)
	if err != nil {
		var zero V
		return zero, err
	}
	c.items[key] = hash_item[V]{
		value:		value,
		hash:		value.Hash(),
		expires:	time_expires(c.ttl),
	}
	return value, nil
}

func (c *Hash[K, V]) purge_expired(){
	c.lock.Lock()
	defer c.lock.Unlock()
	time_unix := time_unix()
	for key, i := range c.items {
		if i.expires != 0 && time_unix > i.expires {
			delete(c.items, key)
		}
	}
}