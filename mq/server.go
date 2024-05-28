package mq

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"

	"github.com/go-redis/redis"
)

type Server struct {
	broker      Broker
	concurrency int
	queues      []string
	wg          sync.WaitGroup
}

type ServeMux struct {
	mp map[string]func(*Task) error
}

func NewServer(r RedisConfig, n int) *Server {
	c := r.MakeRedisClient()
	if n < 1 {
		n = runtime.NumCPU()
	}

	return &Server{broker: &RedisBroker{RedisConnection: *c.(*redis.Client)}, concurrency: n, queues: []string{DEFAULT_QUEUE}}
}

func NewServeMux() *ServeMux {
	mp := make(map[string]func(*Task) error)
	return &ServeMux{mp: mp}
}

func (sm *ServeMux) HandleFunc(name string, f func(*Task) error) {
	sm.mp[name] = f
}

func (s *Server) Run(mux *ServeMux) error {
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
				continue
			}
			if task == nil {
				continue
			}
			if f, ok := mux.mp[task.Name]; ok {
				if err := f(task); err != nil {
					logger.Error(fmt.Sprintf("Error processing task: %s", task.Name), slog.String("error", err.Error()))
					continue
				}
			}
		}
	}
}
