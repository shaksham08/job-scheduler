package mq

import "github.com/go-redis/redis"

type Client struct {
	broker Broker
}

func NewClient(r RedisConfig) *Client {
	c := r.MakeRedisClient()
	return &Client{broker: &RedisBroker{RedisConnection: *c.(*redis.Client)}}
}

func (c *Client) Enqueue(task Task) {
	c.broker.Enqueue(task)
}
