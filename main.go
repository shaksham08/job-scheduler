package main

import (
	"encoding/json"
	"log"

	mq "github.com/shaksham08/job-scheduler/mq"
)

type DeleteFileTask struct {
	FileName string `json:"file_name"`
}

func main() {

	// Create a new client

	client := mq.NewClient(mq.RedisConfig{
		Address: "localhost:6379",
	})

	// Enqueue a message
	payload, err := json.Marshal(DeleteFileTask{FileName: "file_1"})
	if err != nil {
		log.Fatal(err)
	}
	task := mq.NewTask("delete_file", payload)
	client.Enqueue(task)

}
