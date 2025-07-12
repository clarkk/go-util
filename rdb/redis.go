package rdb

import (
	"fmt"
	"time"
	"context"
	"github.com/redis/go-redis/v9"
)

var (
	client 		*redis.Client
	connected 	= false
)

func Connect(host string, port int, auth string){
	if connected {
		panic("Redis is already connected")
	}
	
	client = redis.NewClient(&redis.Options{
		Addr:		fmt.Sprintf("%s:%d", host, port),
		Password:	auth,
		DB:			0,
	})
	if _, err := client.Ping(context.Background()).Result(); err != nil {
		panic("Unable to connect to Redis: "+err.Error())
	}
	
	connected = true
}

func Connected() bool {
	return connected
}

//	Fetch key
func Get(ctx context.Context, key string) (string, bool){
	value, err := client.Get(ctx, key).Result()
	empty := err == redis.Nil
	if err != nil && !empty {
		panic("Redis get: "+err.Error())
	}
	return value, !empty
}

func Hgetall(ctx context.Context, key string, ref any){
	res := client.HGetAll(ctx, key)
	if err := res.Err(); err != nil {
		panic("Redis hgetall: "+err.Error())
	}
	if err := res.Scan(ref); err != nil {
		panic("Redis hgetall scan: "+err.Error())
	}
}

//	Store a single value associated with a key
func Set(ctx context.Context, key string, value []byte, expires int) error {
	return client.Set(ctx, key, value, time.Duration(expires) * time.Second).Err()
}

//	Store multiple key-value pairs with a key
func Hset(ctx context.Context, key string, values any, expires int) error {
	if err := client.HSet(ctx, key, values).Err(); err != nil {
		return err
	}
	client.Expire(ctx, key, time.Duration(expires) * time.Second)
	return nil
}

//	Delete key
func Del(ctx context.Context, key string) error {
	return client.Del(ctx, key).Err()
}