package repository

import (
	"LoadBalancer/TimeLimiter/pkg/model"
	"log"
)

type RateLimiterRepo interface {
	GetClients() error
	GetBuckets() map[string]*model.TokenBucket
}

type RateLimiterRepoImpl struct {
	usRepo *UserRepoImpl
}

func NewRlRepoImpl(db *UserRepoImpl) *RateLimiterRepoImpl {
	log.Println("Creating new NewRlRepoImpl")
	return &RateLimiterRepoImpl{
		usRepo: db,
	}
}

func (r *RateLimiterRepoImpl) GetClients() error {
	return r.usRepo.GetClients()
}

func (r *RateLimiterRepoImpl) GetBuckets() map[string]*model.TokenBucket {
	return r.usRepo.GetBuckets()
}
