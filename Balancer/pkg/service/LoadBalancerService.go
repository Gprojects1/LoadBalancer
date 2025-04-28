package service

import (
	"LoadBalancer/Balancer/pkg/utils"
	"log"
	"net/url"
	"sync"
	"sync/atomic"
)

type LoadBlancerService interface {
	NextIndex() int
	GetNextServer() *Backend
	HealthCheck()
	AddBackend(backend *Backend)
}

type ServerPool struct {
	backends []*Backend
	current  uint64
}

func (s *ServerPool) AddBackend(backend *Backend) {
	s.backends = append(s.backends, backend)
}

func (s *ServerPool) NextIndex() int {
	return int(atomic.AddUint64(&s.current, uint64(1)) % uint64(len(s.backends)))
}

func (s *ServerPool) MarkBackendStatus(backendUrl *url.URL, alive bool) {
	for _, b := range s.backends {
		if b.URL.String() == backendUrl.String() {
			b.SetAlive(alive)
			break
		}
	}
}

func (s *ServerPool) GetNextServer() *Backend {
	next := s.NextIndex()
	l := len(s.backends) + next
	for i := next; i < l; i++ {
		idx := i % len(s.backends)
		if s.backends[idx].IsAlive() {
			if i != next {
				atomic.StoreUint64(&s.current, uint64(idx))
			}
			return s.backends[idx]
		}
	}
	return nil
}

func (s *ServerPool) HealthCheck() {
	var wg sync.WaitGroup
	for _, b := range s.backends {
		wg.Add(1)
		go func(backend *Backend) {
			defer wg.Done()
			status := "up"
			alive := utils.TryToConnect(backend.URL)
			backend.SetAlive(alive)
			if !alive {
				status = "down"
			}
			log.Printf("%s [%s]\n", backend.URL, status)
		}(b)
	}
	wg.Wait()
}
