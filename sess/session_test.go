package sess

/*
	Test
	# go test . -race -v
*/

import (
	"fmt"
	"time"
	"sync"
	"context"
	"testing"
	"math/rand/v2"
)

var ctx = context.Background()

func Test_session(t *testing.T){
	expires 		:= 60 * 5
	purge_interval 	:= 60
	Init(expires, "", "", purge_interval)
	
	//	Create session and close it
	sid := uuid_string()
	fmt.Println("create: "+sid)
	s := create_session(sid)
	s.close()
	
	//	Fetch the session with concurrency and update session
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		fmt.Printf("open: %s (%d)\n", sid, i)
		wg.Add(1)
		go func(i int){
			test_update_session(sid, i)
			wg.Done()
		}(i)
	}
	wg.Wait()
}

func test_update_session(sid string, i int){
	s := fetch_session(ctx, sid)
	if s == nil {
		return
	}
	
	s.reset()
	
	s.Write(session_data{
		"test": "test123",
	})
	time.Sleep(rand.N(100 * time.Millisecond))
	s.close()
	fmt.Printf("updated: %s (%d)\n", sid, i)
}