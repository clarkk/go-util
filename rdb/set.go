package rdb

import (
	"time"
	"context"
)

//	Store members in set
func Sadd(ctx context.Context, key string, values []any, expires int) error {
	if err := client.SAdd(ctx, key, values...).Err(); err != nil {
		return err
	}
	client.Expire(ctx, key, time.Duration(expires) * time.Second)
	return nil
}