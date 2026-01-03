package rdb

import "context"

//	Store members in set
func Sadd(ctx context.Context, key string, values []any, expire int) error {
	pipe := client.Pipeline()
	pipe.SAdd(ctx, key, values...)
	pipe.Expire(ctx, key, time_expire(expire))
	_, err := pipe.Exec(ctx)
	return err
}