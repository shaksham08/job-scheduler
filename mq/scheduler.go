package mq

import (
	"errors"
	"log/slog"
	"time"

	"github.com/robfig/cron/v3"
)

type Scheduler struct {
	cr *cron.Cron
}

type ScheduledTaskEntry struct {
	Task    *Task        `json:"task"`
	EntryId cron.EntryID `json:"entry_id"`
	NextRun time.Time    `json:"next_run"`
}

var scheduledTasksRegistry = make(map[string]*ScheduledTaskEntry)

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

func (s *Scheduler) ScheduleTask(task *Task, execute func(task *Task) error) error {
	if task.Meta.CronExpr == "" {
		return errors.New("no cron expression provided")
	}
	logger.Info("scheduling task", slog.String("task_id", task.Id), slog.String("cron_expr", task.Meta.CronExpr))

	entryID, err := s.cr.AddFunc(task.Meta.CronExpr, func() {
		err := execute(task)
		if err != nil {
			logger.Error("Error executing task", slog.String("error:", err.Error()))
		}
		entryID := scheduledTasksRegistry[task.Id].EntryId
		scheduledTasksRegistry[task.Id].NextRun = s.cr.Entry(entryID).Next
	})
	if err != nil {
		return err
	}

	// Initial registry update with the first scheduled run time
	scheduledTasksRegistry[task.Id] = &ScheduledTaskEntry{
		Task:    task,
		NextRun: s.cr.Entry(entryID).Next,
		EntryId: entryID,
	}

	return nil
}
