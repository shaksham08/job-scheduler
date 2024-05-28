package mq

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"

	"github.com/go-redis/redis"
)

type Server struct {
	broker      Broker
	concurrency int
	queues      []string
	wg          sync.WaitGroup
	scheduler   *Scheduler
}

type ServeMux struct {
	mp map[string]func(*Task) error
}

func NewServer(r RedisConfig, n int) *Server {
	c := r.MakeRedisClient()
	if n < 1 {
		n = runtime.NumCPU()
	}

	return &Server{broker: &RedisBroker{RedisConnection: *c.(*redis.Client)}, concurrency: n, queues: []string{DEFAULT_QUEUE, CRON_QUEUE}, scheduler: NewScheduler()}
}

func NewServeMux() *ServeMux {
	mp := make(map[string]func(*Task) error)
	return &ServeMux{mp: mp}
}

func (sm *ServeMux) HandleFunc(name string, f func(*Task) error) {
	sm.mp[name] = f
}

func (s *Server) Run(mux *ServeMux) error {
	s.scheduler.Start()
	defer s.scheduler.Stop()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	for i := 0; i < s.concurrency; i++ {
		s.wg.Add(1)
		go s.worker(mux, ctx)
	}
	<-stop
	cancel()
	logger.Info("Stopping server...")
	s.wg.Wait()
	return nil
}

func (s *Server) worker(mux *ServeMux, ctx context.Context) {
	defer s.wg.Done()
	for {
		select {
		case <-ctx.Done():
			logger.Info("Worker shutting down...")
			return
		default:
			task, err := s.broker.Dequeue(s.queues...)
			if err != nil {
				time.Sleep(1 * time.Second) // Sleep to prevent tight loop on error
				continue
			}
			if task == nil {
				time.Sleep(1 * time.Second) // Sleep to prevent tight loop on nil task
				continue
			}

			//  Check if task is in mux
			handler, ok := mux.mp[task.Name]
			if !ok {
				logger.Error("No handler for task", slog.String("task_id:", task.Id))
				continue
			}

			// handle cron tasks
			if task.Meta.CronExpr != "" {
				_, err := s.scheduler.ScheduleTask(task, handler)
				if err != nil {
					logger.Error("Error scheduling task", slog.String("task_id:", task.Id))
				}
				continue
			}

			// handle non-cron tasks
			if err := handler(task); err != nil {
				// retry logic
				logger.Info("Retrying task ", slog.String("task_id:", task.Id))
				task.Meta.CurrentRetries++
				if task.Meta.CurrentRetries > task.Meta.MaxRetries {
					logger.Error("Max retries reached, dropping task ", slog.String("task_id:", task.Id))
					continue
				}
				time.Sleep(RETRY_DELAY * time.Second)
				err := s.broker.Enqueue(*task)
				if err != nil {
					logger.Error("Error re-enqueuing task", slog.String("task_id:", task.Id))
				}
			}

		}
	}
}
