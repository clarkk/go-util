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

//	Fetch hash
func Get(ctx context.Context, key string) (string, bool){
	value, err := client.Get(ctx, key).Result()
	empty := err == redis.Nil
	if err != nil && !empty {
		panic("Redis get: "+err.Error())
	}
	return value, !empty
}

//	Store a single value associated with hash
func Set(ctx context.Context, key string, value []byte, expire int) error {
	return client.Set(ctx, key, value, time_expire(expire)).Err()
}

//	Fetch field in hash
func Hget(ctx context.Context, key, field string) (string, bool){
	value, err := client.HGet(ctx, key, field).Result()
	empty := err == redis.Nil
	if err != nil && !empty {
		panic("Redis hget: "+err.Error())
	}
	return value, !empty
}

//	Fetch all fields in hash
func Hgetall(ctx context.Context, key string, ref any){
	res := client.HGetAll(ctx, key)
	if err := res.Err(); err != nil {
		panic("Redis hgetall: "+err.Error())
	}
	if err := res.Scan(ref); err != nil {
		panic("Redis hgetall scan: "+err.Error())
	}
}

//	Store multiple key-value pairs in hash
func Hset(ctx context.Context, key string, values any, expire int) error {
	if err := client.HSet(ctx, key, values).Err(); err != nil {
		return err
	}
	Expire(ctx, key, expire)
	return nil
}

func Expire(ctx context.Context, key string, expire int){
	client.Expire(ctx, key, time_expire(expire))
}

//	Delete hash
func Del(ctx context.Context, key string) error {
	return client.Del(ctx, key).Err()
}

func time_expire(expire int) time.Time {
	return time.Duration(expire) * time.Second
}