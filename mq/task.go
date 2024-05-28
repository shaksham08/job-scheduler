package mq

import "github.com/google/uuid"

type TaskMeta struct {
	Payload        []byte `json:"payload"`
	MaxRetries     int    `json:"max_retries"`
	CurrentRetries int    `json:"current_retries"`
}

type Task struct {
	Id   string   `json:"id"`
	Name string   `json:"name"`
	Meta TaskMeta `json:"info"`
}

func NewTask(name string, meta TaskMeta) Task {
	id := uuid.New().String()
	return Task{Id: id, Name: name, Meta: meta}
}
