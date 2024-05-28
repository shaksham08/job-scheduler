package mq

import (
	"runtime"
	"sync"

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
	for i := 0; i < s.concurrency; i++ {
		s.wg.Add(1)
		go s.worker(mux)
	}
	s.wg.Wait()
	return nil
}

func (s *Server) worker(mux *ServeMux) {
	defer s.wg.Done()
	for {
		task, err := s.broker.Dequeue(s.queues...)
		if err != nil {
			return
		}
		if task == nil {
			continue
		}
		if f, ok := mux.mp[task.Name]; ok {
			if err := f(task); err != nil {
				return
			}
		}
	}
}
