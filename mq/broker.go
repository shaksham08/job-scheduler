package mq

import (
	"encoding/json"

	"github.com/go-redis/redis"
)

type Broker interface {
	Enqueue(task Task) error
	Dequeue(queues ...string) (*Task, error)
}

type RedisBroker struct {
	RedisConnection redis.Client
}

func (r *RedisBroker) Enqueue(task Task) error {
	taskBytes, err := json.Marshal(task)
	if err != nil {
		return err
	}

	cmd := r.RedisConnection.LPush(DEFAULT_QUEUE, taskBytes)
	if cmd.Err() != nil {
		return cmd.Err()
	}
	return nil
}

func (r *RedisBroker) Dequeue(queues ...string) (*Task, error) {
	cmd := r.RedisConnection.BRPop(0, queues...)
	if cmd.Err() != nil {
		return nil, cmd.Err()
	}

	var task Task
	err := json.Unmarshal([]byte(cmd.Val()[1]), &task)
	if err != nil {
		return nil, err
	}

	return &task, nil
}
