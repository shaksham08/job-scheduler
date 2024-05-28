package mq

import "github.com/google/uuid"

type Task struct {
	Id      string `json:"id"`
	Name    string `json:"name"`
	Payload []byte `json:"payload"`
}

func NewTask(name string, payload []byte) Task {
	id := uuid.New().String()
	return Task{Id: id, Name: name, Payload: payload}
}
