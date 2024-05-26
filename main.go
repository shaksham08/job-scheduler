package main

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/redis/go-redis/v9"
)

var redisClient *redis.Client
var taskFunctions = make(map[string]func())

type Client struct {
}

type Task struct {
	ID      string
	Name    string
	Payload []byte
}

func (c *Client) Enqueue(t Task, ctx context.Context) {
	// store to redis
	task_json, err := json.Marshal(t)
	if err != nil {
		panic(err)
	}
	redisClient.LPush(ctx, "task_example", task_json)

}

func NewClient() *Client {
	return &Client{}
}

func NewTask(id string, name string, payload []byte) *Task {
	return &Task{
		ID:      id,
		Name:    name,
		Payload: payload,
	}
}

func Worker(wg *sync.WaitGroup, ctx context.Context) {
	defer wg.Done()
	defer fmt.Println("worker done")

	fmt.Println("Worker started")

	task_json, err := redisClient.BRPop(ctx, 0, "task_example").Result()
	if err != nil {
		panic(err)
	}
	var task Task
	err = json.Unmarshal([]byte(task_json[1]), &task)
	if err != nil {
		panic(err)
	}
	taskFunctions[task.Name]()
}

func NewServer(concurrency int, ctx context.Context) {
	var wg sync.WaitGroup
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go Worker(&wg, ctx)
	}
	wg.Wait()
}

func RegisterFunction(name string, f func()) {
	taskFunctions[name] = f
}

func main() {

	ctx := context.Background()

	redisClient = redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "",
		DB:       0,
	})

	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		panic(err)
	}

	// client
	client := NewClient()
	task := NewTask("1", "task1", []byte("payload"))
	client.Enqueue(*task, ctx)

	// worker
	RegisterFunction("task1", func() {
		fmt.Println("Task 1 executed")
	})

	fmt.Println("task functions", taskFunctions)

	// Server
	NewServer(5, ctx)

}
