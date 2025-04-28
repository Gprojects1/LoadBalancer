package model

import (
	"sync"
	"time"
)

type TokenBucket struct {
	capacity    int
	ratePerSec  float64
	tokens      float64
	lastRefill  time.Time
	client_id   string
	updateMutex sync.Mutex
}

func (tb *TokenBucket) Available() float64 {
	tb.updateMutex.Lock()
	defer tb.updateMutex.Unlock()

	now := time.Now()
	tb.refillTokens(now)
	return tb.tokens
}

func (tb *TokenBucket) Take(n float64) bool {
	tb.updateMutex.Lock()
	defer tb.updateMutex.Unlock()

	now := time.Now()
	tb.refillTokens(now)

	if tb.tokens >= n {
		tb.tokens -= n
		return true
	}
	return false
}

func (tb *TokenBucket) refillTokens(now time.Time) {
	elapsed := now.Sub(tb.lastRefill)
	tb.lastRefill = now

	refillAmount := tb.ratePerSec * elapsed.Seconds()
	tb.tokens += refillAmount

	if tb.tokens > float64(tb.capacity) {
		tb.tokens = float64(tb.capacity)
	}
}

func NewTokenBucket(clientID string, capacity int, ratePerSec float64) *TokenBucket {
	return &TokenBucket{
		capacity:   capacity,
		ratePerSec: ratePerSec,
		tokens:     float64(capacity),
		lastRefill: time.Now(),
		client_id:  clientID,
	}
}

func (tb *TokenBucket) SetCapacity(capacity int) {
	tb.updateMutex.Lock()
	defer tb.updateMutex.Unlock()
	tb.capacity = capacity
	if tb.tokens > float64(capacity) {
		tb.tokens = float64(capacity)
	}
}

func (tb *TokenBucket) SetRate(ratePerSec float64) {
	tb.updateMutex.Lock()
	defer tb.updateMutex.Unlock()
	tb.ratePerSec = ratePerSec
}
