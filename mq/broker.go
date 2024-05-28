package mq

import (
	"github.com/go-redis/redis"
)

type Broker interface {
	Enqueue(message string)
}

type RedisBroker struct {
	RedisConnection redis.Client
}

func (r *RedisBroker) Enqueue(message string) {
	r.RedisConnection.LPush(DEFAULT_QUEUE, message)
}
