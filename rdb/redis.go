package rdb

import (
	"fmt"
	"time"
	"context"
	"github.com/redis/go-redis/v9"
)

var (
	client 		*redis.Client
	host 		= "localhost"
	port 		= 6379
	
	connected 	= false
)

func Connect(auth string){
	if connected {
		panic("Redis is already connected")
	}
	
	client = redis.NewClient(&redis.Options{
		Addr:		fmt.Sprintf("%s:%d", host, port),
		Password:	auth,
		DB:			0,
	})
	if _, err := client.Ping(context.Background()).Result(); err != nil {
		panic(err)
	}
	
	connected = true
}

func Connected() bool {
	return connected
}

func Get(ctx context.Context, key string) (string, bool) {
	value, err := client.Get(ctx, key).Result()
	if err != nil && err != redis.Nil {
		panic(err)
	}
	found := err != redis.Nil
	return value, found
}

func Hgetall(ctx context.Context, key string, ref interface{}) bool {
	res := client.HGetAll(ctx, key)
	err := res.Err()
	if err != nil && err != redis.Nil {
		panic(err)
	}
	found := err != redis.Nil
	if found {
		if err := res.Scan(ref); err != nil {
			panic(err)
		}
	}
	return found
}

func Set(ctx context.Context, key string, value []byte, expires int) error {
	return client.Set(ctx, key, value, time.Duration(expires) * time.Second).Err()
}

func Hset(ctx context.Context, key string, values interface{}, expires int) error {
	err := client.HSet(ctx, key, values).Err()
	client.Expire(ctx, key, time.Duration(expires) * time.Second)
	return err
}

func Delete(ctx context.Context, key string) error {
	return client.Del(ctx, key).Err()
}