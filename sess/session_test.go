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
	
	t.Run("mutex race conditions", func(t *testing.T){
		//	Create session and close it
		sid, s := test_create_session()
		s.close()
		
		//	Fetch the session with concurrency and update session
		var wg sync.WaitGroup
		for i := range 100 {
			fmt.Printf("open: %s (%d)\n", sid, i)
			wg.Add(1)
			go func(i int){
				test_update_session_random_sleep(t, sid, i)
				wg.Done()
			}(i)
		}
		wg.Wait()
	})
	
	t.Run("integrity", func(t *testing.T){
		var (
			want	string
			got		string
		)
		
		sid, s1 := test_create_session()
		
		//	Update session 1
		s1.Write(session_data{
			"test": "a",
		})
		s1.close()
		if !s1.Closed() {
			t.Fatalf("Session not properly closed")
		}
		fmt.Printf("updated session 1: %s\n", sid)
		
		want 	= "a"
		got 	= s1.Data()["test"].(string)
		if got != want {
			t.Fatalf("Session data want [%s] but got [%s]",
				want,
				got,
			)
		}
		
		//	Update session 2
		s2 := test_fetch_session(t, sid)
		s2.Write(session_data{
			"test": "b",
		})
		s2.close()
		if !s2.Closed() {
			t.Fatalf("Session not properly closed")
		}
		fmt.Printf("updated session 2: %s\n", sid)
		
		want 	= "a"
		got 	= s1.Data()["test"].(string)
		if got != want {
			t.Fatalf("Session data want [%s] but got [%s]",
				want,
				got,
			)
		}
	})
	
	t.Run("integrity closed", func(t *testing.T){
		var (
			wg		sync.WaitGroup
			want	string
			got		string
			
			sid		string
			s1		*Session
			s2		*Session
			
			s2_csrf	string
		)
		
		sid, s1 = test_create_session()
		
		//	Update session 1
		s1.Write(session_data{
			"user_id": "123",
		})
		s1.close()
		if !s1.Closed() {
			t.Fatalf("Session not properly closed")
		}
		fmt.Printf("updated session 1: %s\n", sid)
		
		wg.Add(1)
		go func(i int){
			s2 = test_fetch_session(t, sid)
			s2.Write(session_data{
				"user_id": "456",
			})
			s2.generate_CSRF()
			s2_csrf = s2.csrf_token()
			time.Sleep(100 * time.Millisecond)
			s2.close()
			if !s2.Closed() {
				t.Fatalf("Session not properly closed")
			}
			fmt.Printf("updated session %d: %s\n", i, sid)
			wg.Done()
		}(2)
		
		wg.Wait()
		
		want 	= "123"
		got 	= s1.Data()["user_id"].(string)
		if got != want {
			t.Fatalf("Session data want [%s] but got [%s]",
				want,
				got,
			)
		}
		
		want 	= ""
		got 	= s1.csrf_token()
		if got != want {
			t.Fatalf("Session CSRF want [%s] but got [%s]",
				want,
				got,
			)
		}
		
		want 	= "456"
		got 	= s2.Data()["user_id"].(string)
		if got != want {
			t.Fatalf("Session data want [%s] but got [%s]",
				want,
				got,
			)
		}
		
		want 	= s2_csrf
		got 	= s2.csrf_token()
		if got != want {
			t.Fatalf("Session CSRF want [%s] but got [%s]",
				want,
				got,
			)
		}
	})
}

func test_update_session_random_sleep(t *testing.T, sid string, i int){
	s := test_fetch_session(t, sid)
	s.Write(session_data{
		"test": "test123",
	})
	time.Sleep(rand.N(100 * time.Millisecond))
	s.close()
	if !s.Closed() {
		t.Fatalf("Session not properly closed")
	}
	fmt.Printf("updated: %s (%d)\n", sid, i)
}

func test_fetch_session(t *testing.T, sid string) *Session {
	sess, _ := fetch_session(ctx, sid)
	if sess == nil {
		t.Fatalf("Unable to fetch session")
	}
	sess.reset()
	return wrap_session(sess)
}

func test_create_session() (string, *Session){
	sid := uuid_string()
	fmt.Println("created: "+sid)
	return sid, wrap_session(create_session(sid))
}