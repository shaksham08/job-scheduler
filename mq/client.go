package mq

import (
	"github.com/go-redis/redis"
)

type Client struct {
	broker Broker
}

func NewClient(r RedisConfig) *Client {
	c := r.MakeRedisClient()
	return &Client{broker: &RedisBroker{RedisConnection: *c.(*redis.Client)}}
}

func (c *Client) Enqueue(task Task) error {
	return c.broker.Enqueue(task)
}

func GetAllScheduledTasks() (map[string]*ScheduledTaskEntry, error) {
	return scheduledTasksRegistry, nil
}
