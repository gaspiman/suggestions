package datastore

import "github.com/go-redis/redis/v8"

type Cachestore struct {
	Client *redis.Client
}

func NewCachestore() *Cachestore {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	cs := &Cachestore{
		Client: rdb,
	}
	return cs
}
