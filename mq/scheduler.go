package mq

import (
	"errors"
	"log/slog"

	"github.com/robfig/cron/v3"
)

type Scheduler struct {
	cr *cron.Cron
}

func NewScheduler() *Scheduler {
	return &Scheduler{
		cr: cron.New(cron.WithSeconds()), // Initializes a cron scheduler with second precision
	}
}

func (s *Scheduler) Start() {
	s.cr.Start()
}

func (s *Scheduler) Stop() {
	logger.Info("stopping scheduler")
	s.cr.Stop()
}

func (s *Scheduler) ScheduleTask(task *Task, execute func(task *Task) error) (cron.EntryID, error) {
	if task.Meta.CronExpr == "" {
		return 0, errors.New("no cron expression provided")
	}

	logger.Info("scheduling task", slog.String("task_id", task.Id), slog.String("cron_expr", task.Meta.CronExpr))

	id, err := s.cr.AddFunc(task.Meta.CronExpr, func() {
		err := execute(task)
		if err != nil {
			logger.Error("Error executing task", slog.String("error:", err.Error()))
		}
	})

	return id, err
}
