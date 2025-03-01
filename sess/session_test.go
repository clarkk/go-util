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
	t.Run("mutex race conditions", func(t *testing.T){
		expires 		:= 60 * 5
		purge_interval 	:= 60
		Init(expires, "", "", purge_interval)
		
		//	Create session and close it
		sid, s := test_create_session()
		s.close()
		
		//	Fetch the session with concurrency and update session
		var wg sync.WaitGroup
		for i := range 100 {
			fmt.Printf("open: %s (%d)\n", sid, i)
			wg.Add(1)
			go func(i int){
				test_update_session(t, sid, i)
				wg.Done()
			}(i)
		}
		wg.Wait()
	})
	
	t.Run("data session closed", func(t *testing.T){
		sid, s1 := test_create_session()
		
		s1.Write(session_data{
			"test": "a",
		})
		s1.close()
		fmt.Printf("updated: %s\n", sid)
		
		if !s1.Closed() {
			t.Fatalf("Session not properly closed")
		}
		
		want 	:= "a"
		got 	:= s1.Data()["test"].(string)
		if got != want {
			t.Fatalf("Session data want [%s] but got [%s]",
				want,
				got,
			)
		}
		
		s2 := test_fetch_session(t, sid)
		s2.Write(session_data{
			"test": "b",
		})
		s2.close()
		fmt.Printf("updated: %s\n", sid)
		
		if !s2.Closed() {
			t.Fatalf("Session not properly closed")
		}
		
		want 	= "a"
		got 	= s1.Data()["test"].(string)
		if got != want {
			t.Fatalf("Session data want [%s] but got [%s]",
				want,
				got,
			)
		}
	})
}

func test_update_session(t *testing.T, sid string, i int){
	s := test_fetch_session(t, sid)
	s.Write(session_data{
		"test": "test123",
	})
	time.Sleep(rand.N(100 * time.Millisecond))
	s.close()
	fmt.Printf("updated: %s (%d)\n", sid, i)
}

func test_fetch_session(t *testing.T, sid string) *Session {
	sess := fetch_session(ctx, sid)
	if sess == nil {
		t.Fatalf("Unable to fetch session")
	}
	sess.reset()
	return wrap_session(sess)
}

func test_create_session() (string, *Session){
	sid := uuid_string()
	fmt.Println("create: "+sid)
	return sid, wrap_session(create_session(sid))
}