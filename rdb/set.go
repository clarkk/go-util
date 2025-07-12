package rdb

import (
	"time"
	"context"
	//"github.com/redis/go-redis/v9"
)

//	Store members in set
func Sadd(ctx context.Context, key string, values []any, expires int) error {
	if err := client.Sadd(ctx, key, values...).Err(); err != nil {
		return err
	}
	client.Expire(ctx, key, time.Duration(expires) * time.Second)
	return nil
}