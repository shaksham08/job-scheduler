package main

import (
	"encoding/json"
	"log"
	"log/slog"
	"os"
	"time"

	mq "github.com/shaksham08/job-scheduler/mq"
)

type DeleteFileTask struct {
	FileName string `json:"file_name"`
}

type CreateFileTask struct {
	FileName string `json:"file_name"`
	Location string `json:"location"`
}

var logger *slog.Logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
	Level: slog.LevelDebug,
}))

func main() {

	// Create a new client

	client := mq.NewClient(mq.RedisConfig{
		Address: "localhost:6379",
	})

	var tasks []mq.Task

	// Create Task 1
	payload, err := json.Marshal(DeleteFileTask{FileName: "file_1"})
	if err != nil {
		logger.Error("Error marshalling payload", slog.String("error:", err.Error()))
	}

	task_meta := mq.TaskMeta{
		Payload:        payload,
		MaxRetries:     3,
		CurrentRetries: 0,
	}
	task := mq.NewTask("delete_file", task_meta)

	tasks = append(tasks, task)

	// Create Task 2
	payload, err = json.Marshal(CreateFileTask{FileName: "file_2", Location: "/tmp"})
	if err != nil {
		logger.Error("Error marshalling payload", slog.String("error:", err.Error()))
	}

	task_meta = mq.TaskMeta{
		Payload:        payload,
		MaxRetries:     3,
		CurrentRetries: 0,
		CronExpr:       "*/15 * * * * *",
	}

	task = mq.NewTask("create_file", task_meta)
	tasks = append(tasks, task)

	// Enqueue the tasks
	for _, t := range tasks {
		err := client.Enqueue(t)
		if err != nil {
			logger.Error("Error enqueuing task", slog.String("error:", err.Error()))
		}
	}

	// Create a server
	server := mq.NewServer(mq.RedisConfig{
		Address: "localhost:6379",
	}, 10)

	mux := mq.NewServeMux()
	mux.HandleFunc("delete_file", func(task *mq.Task) error {
		var t DeleteFileTask
		err := json.Unmarshal(task.Meta.Payload, &t)
		if err != nil {
			return err
		}
		logger.Info("Executing task, deleting file...", slog.String("file_name:", t.FileName))
		return nil
	})
	mux.HandleFunc("create_file", func(task *mq.Task) error {
		var t CreateFileTask
		err := json.Unmarshal(task.Meta.Payload, &t)
		if err != nil {
			return err
		}
		logger.Info("Executing task, creating file...", slog.String("file_name:", t.FileName), slog.String("location:", t.Location))
		return nil
	})

	go func() {
		for {
			logger.Info("Starting to fetch scheduled tasks...")
			// fetch scheduled tasks
			tasks, err := mq.GetAllScheduledTasks()
			if err != nil {
				logger.Error("Error fetching scheduled tasks", slog.String("error:", err.Error()))
			}
			for _, t := range tasks {
				logger.Info("Task scheduled: ", slog.String("task_id:", t.Task.Id), slog.String("next_run:", t.NextRun.String()), slog.String("name:", t.Task.Name), slog.String("cron_expr:", t.Task.Meta.CronExpr))
			}
			time.Sleep(10 * time.Second)
		}
	}()

	if err := server.Run(mux); err != nil {
		log.Fatal(err)
	}

}
