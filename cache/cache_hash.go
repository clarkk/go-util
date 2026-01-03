package cache

import (
	"fmt"
	"log"
	"time"
	"sync"
	"bytes"
	"encoding/gob"
	"github.com/clarkk/go-util/hash"
)

var buffer_pool = sync.Pool{
    New: func() any {
        return bytes.NewBuffer(make([]byte, 0, 512))
    },
}

type (
	Cache_hash[K comparable, V any] struct {
		lock		sync.RWMutex
		items		map[K]cache_hash_item[V]
		ttl			int
		verify		func(key K, hash *string) (bool, error)
		refresh		func(key K) (V, *string, error)
	}
	
	cache_hash_item[V any] struct {
		value		V
		hash		*string
		expires		int64
	}
)

//	Create new hash cache
func NewCache_hash[K comparable, V any](
	ttl int,
	verify func(key K, hash *string) (bool, error),
	refresh func(key K) (V, *string, error),
	purge_interval int,
) *Cache_hash[K, V] {
	c := &Cache_hash[K, V]{
		items:		map[K]cache_hash_item[V]{},
		ttl:		ttl,
		verify:		verify,
		refresh:	refresh,
	}
	//	Purge expired values from cache with time interval
	ticker := time.NewTicker(time.Duration(purge_interval) * time.Second)
	go func(){
		for range ticker.C {
			func(){
				defer func(){
					if r := recover(); r != nil {
						log.Printf("purge_expired panic: %v", r)
					}
				}()
				c.purge_expired()
			}()
		}
	}()
	return c
}

//	Generate hash of value
func Hash(v any) (*string, error){
	if v == nil {
		return nil, nil
	}
	
	buf := buffer_pool.Get().(*bytes.Buffer)
	buf.Reset()
	defer buffer_pool.Put(buf)
	
	enc := gob.NewEncoder(buf)
	if err := enc.Encode(v); err != nil {
		return nil, fmt.Errorf("Binary serialization: %w", err)
	}
	s := hash.SHA256_hex(buf.Bytes())
	return &s, nil
}

//	Get cached value
func (c *Cache_hash[K, V]) Get(key K) (V, error){
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
func (c *Cache_hash[K, V]) Refresh(key K) (V, error){
	c.lock.Lock()
	defer c.lock.Unlock()
	
	value, hash, err := c.refresh(key)
	if err != nil {
		var zero V
		return zero, err
	}
	c.items[key] = cache_hash_item[V]{
		value:		value,
		hash:		hash,
		expires:	time_expires(c.ttl),
	}
	return value, nil
}

//	Delete value in cache
func (c *Cache_hash[K, V]) Delete(key K){
	c.lock.Lock()
	defer c.lock.Unlock()
	if _, ok := c.items[key]; ok {
		delete(c.items, key)
	}
}

func (c *Cache_hash[K, V]) purge_expired(){
	if ok := c.lock.TryLock(); !ok {
		return
	}
	defer c.lock.Unlock()
	time_unix := time_unix()
	for key, i := range c.items {
		if i.expires != 0 && time_unix > i.expires {
			delete(c.items, key)
		}
	}
}