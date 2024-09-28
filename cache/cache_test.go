package cache

/*
	Test
	# go test . -v
*/

import (
	"time"
	"sync"
	"testing"
)

func Test_cache(t *testing.T){
	var wg sync.WaitGroup
	
	purge_interval 	:= 60
	c := New[any](purge_interval)
	
	_, ok := c.Get("TEST1")
	if ok != false {
		t.Fatalf("Cache should return false")
	}
	
	wg.Add(1)
	c.Set("TEST1", "cache1", 5)
	go func(){
		defer wg.Done()
		
		time.Sleep(1 * time.Second)
		val, ok := c.Get("TEST1")
		if ok != true {
			t.Fatalf("Cache should return true")
		}
		if val != "cache1" {
			t.Fatalf("Invalid cache value")
		}
		
		time.Sleep(1 * time.Second)
		val, ok = c.Get("TEST1")
		if ok != true {
			t.Fatalf("Cache should return true")
		}
		if val != "cache1" {
			t.Fatalf("Invalid cache value")
		}
		
		time.Sleep(6 * time.Second)
		val, ok = c.Get("TEST1")
		if ok != false {
			t.Fatalf("Cache should return false")
		}
		
		c.Set("TEST1", "cache1", 5)
		val, ok = c.Get("TEST1")
		if ok != true {
			t.Fatalf("Cache should return true")
		}
		if val != "cache1" {
			t.Fatalf("Invalid cache value")
		}
	}()
	
	wg.Wait()
}