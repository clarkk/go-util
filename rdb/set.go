package rdb

import (
	"context"
)

//	Store members in set
func Sadd(ctx context.Context, key string, values []any, expire int) error {
	if err := client.SAdd(ctx, key, values...).Err(); err != nil {
		return err
	}
	Expire(ctx, key, expire)
	return nil
}