package mq

import (
	"encoding/json"
	"log/slog"
	"os"
	"time"

	"github.com/go-redis/redis"
)

var logger *slog.Logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
	Level: slog.LevelDebug,
}))

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

	queue := DEFAULT_QUEUE
	if task.Meta.CronExpr != "" {
		queue = CRON_QUEUE
	}

	cmd := r.RedisConnection.LPush(queue, taskBytes)
	if cmd.Err() != nil {
		return cmd.Err()
	}
	return nil
}

func (r *RedisBroker) Dequeue(queues ...string) (*Task, error) {
	// BRPop is a blocking operation, so it will wait for a task to be enqueued, adding a timeout of 10 seconds
	cmd := r.RedisConnection.BRPop(10*time.Second, queues...)
	if cmd.Err() != nil {
		logger.Error("Error in BRPop", slog.String("error:", cmd.Err().Error()))
		return nil, cmd.Err()
	}

	var task Task
	err := json.Unmarshal([]byte(cmd.Val()[1]), &task)
	if err != nil {
		return nil, err
	}

	return &task, nil
}
