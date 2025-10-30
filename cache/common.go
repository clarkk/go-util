package cache

import "time"

func time_unix() int64 {
	return time.Now().Unix()
}

func time_expires(ttl int) int64 {
	if ttl == 0 {
		return 0
	}
	return time_unix() + int64(ttl)
}