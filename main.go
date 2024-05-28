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
		log.Println(err)
	}
	task := mq.NewTask("delete_file", payload)
	err = client.Enqueue(task)
	if err != nil {
		log.Println(err)
	}

	// Create a server
	server := mq.NewServer(mq.RedisConfig{
		Address: "localhost:6379",
	}, 10)

	mux := mq.NewServeMux()
	mux.HandleFunc("delete_file", func(task *mq.Task) error {
		var t DeleteFileTask
		err := json.Unmarshal(task.Payload, &t)
		if err != nil {
			return err
		}
		log.Println("Deleting file:", t.FileName)
		return nil
	})

	if err := server.Run(mux); err != nil {
		log.Fatal(err)
	}

}
