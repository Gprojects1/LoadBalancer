package service

import (
	"LoadBalancer/TimeLimiter/pkg/repository"
	"fmt"
	"sync"
	"time"
)

type RateLimiterService struct {
	mu         sync.RWMutex
	ticker     *time.Ticker
	tickerStop chan bool
	repo       repository.RateLimiterRepo
}

func NewRateLimiter(repo repository.RateLimiterRepo) (*RateLimiterService, error) {
	rl := &RateLimiterService{
		repo:       repo,
		tickerStop: make(chan bool),
	}
	if err := rl.repo.GetClients(); err != nil {
		return nil, fmt.Errorf("failed to load clients from DB: %w", err)
	}

	rl.ticker = time.NewTicker(time.Second)
	go rl.startTokenRefill()
	return rl, nil
}

func (rl *RateLimiterService) startTokenRefill() {
	for {
		select {
		case <-rl.ticker.C:
			rl.refillAllBuckets()
		case <-rl.tickerStop:
			rl.ticker.Stop()
			return
		}
	}
}

func (rl *RateLimiterService) StopRefill() {
	rl.tickerStop <- true
}

func (rl *RateLimiterService) refillAllBuckets() {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	for _, bucket := range rl.repo.GetBuckets() {
		bucket.Available()
	}
}

func (rl *RateLimiterService) Allow(clientID string) bool {
	rl.mu.RLock()
	bucket, ok := rl.repo.GetBuckets()[clientID]
	rl.mu.RUnlock()

	if !ok {
		return false
	}

	if bucket.Take(1) {
		return true
	}
	return false
}
