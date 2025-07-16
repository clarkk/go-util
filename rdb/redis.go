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
func Get(ctx context.Context, key string) (value string, not_found bool, err error){
	value, err = client.Get(ctx, key).Result()
	not_found = not_found_error(err)
	return
}

//	Store a single value associated with hash
func Set(ctx context.Context, key string, value []byte, expire int) error {
	return client.Set(ctx, key, value, time_expire(expire)).Err()
}

//	Fetch field in hash
func Hget(ctx context.Context, key, field string) (value string, not_found bool, err error){
	value, err = client.HGet(ctx, key, field).Result()
	not_found = not_found_error(err)
	return
}

//	Fetch all fields in hash
func Hgetall(ctx context.Context, key string, ref any) error {
	res := client.HGetAll(ctx, key)
	if err := res.Err(); err != nil {
		return err
	}
	if err := res.Scan(ref); err != nil {
		return err
	}
	return nil
}

//	Store multiple key-value pairs in hash
func Hset(ctx context.Context, key string, values any, expire int) error {
	if err := client.HSet(ctx, key, values).Err(); err != nil {
		return err
	}
	if err := Expire(ctx, key, expire); err != nil {
		return err
	}
	return nil
}

func Expire(ctx context.Context, key string, expire int) error {
	status := client.Expire(ctx, key, time_expire(expire))
    return status.Err()
}

//	Delete hash
func Del(ctx context.Context, key string) error {
	return client.Del(ctx, key).Err()
}

//	Close connection
func Close() error {
	if client != nil {
		return client.Close()
	}
	return nil
}

func time_expire(expire int) time.Duration {
	return time.Duration(expire) * time.Second
}

func not_found_error(err error) bool {
	return err != nil && err == redis.Nil
}