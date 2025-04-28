package controller

import (
	"LoadBalancer/Balancer/pkg/service"
	"LoadBalancer/Balancer/pkg/utils"
	"context"
	"log"
	"net/http"
	"sync"
	"time"
)

type LoadBalancerController interface {
	BalanceRequest(w http.ResponseWriter, r *http.Request)
	AddNewBackend(backend *service.Backend)
}

type LoadBalancerImpl struct {
	service service.LoadBlancerService
}

func NewLoadBlancerController(service service.LoadBlancerService) LoadBalancerController {
	return &LoadBalancerImpl{
		service: service,
	}
}

func (lb *LoadBalancerImpl) AddNewBackend(backend *service.Backend) {
	lb.service.AddBackend(backend)
}

func (lb LoadBalancerImpl) BalanceRequest(w http.ResponseWriter, r *http.Request) {
	attempts := utils.GetAttemptsFromContext(r)
	if attempts > 3 {
		log.Printf("%s(%s) Max attempts reached, terminating\n", r.RemoteAddr, r.URL.Path)
		http.Error(w, "Service not available", http.StatusServiceUnavailable)
		return
	}

	peer := lb.service.GetNextServer()
	if peer == nil {
		http.Error(w, "Service not available", http.StatusServiceUnavailable)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	r = r.WithContext(ctx)

	var wg sync.WaitGroup
	wg.Add(1)

	errChan := make(chan error, 1)

	go func() {
		defer wg.Done()
		peer.ReverseProxy.ServeHTTP(w, r)
	}()

	go func() {
		wg.Wait()
		close(errChan)
	}()

	select {
	case <-ctx.Done():
		log.Printf("Request canceled: %v", ctx.Err())
		return
	case err := <-errChan:
		if err != nil {
			log.Printf("Proxy error: %v", err)
		}
	}
}
