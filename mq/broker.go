package mq

import (
	"encoding/json"

	"github.com/go-redis/redis"
)

type Broker interface {
	Enqueue(task Task)
}

type RedisBroker struct {
	RedisConnection redis.Client
}

func (r *RedisBroker) Enqueue(task Task) {
	taskBytes, err := json.Marshal(task)
	if err != nil {
		panic(err)
	}

	cmd := r.RedisConnection.LPush(DEFAULT_QUEUE, taskBytes)
	if cmd.Err() != nil {
		panic(cmd.Err())
	}
}
