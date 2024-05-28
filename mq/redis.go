package mq

import "github.com/go-redis/redis"

type RedisConnection interface {
	MakeRedisClient() interface{}
}

// implements the RedisConnection interface
type RedisConfig struct {
	Network string
	Address string
	DB      int
}

func (r RedisConfig) MakeRedisClient() interface{} {
	return redis.NewClient(&redis.Options{
		Addr: r.Address,
	})
}
