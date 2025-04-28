package service

import (
	"LoadBalancer/TimeLimiter/pkg/model"
	"LoadBalancer/TimeLimiter/pkg/repository"
)

type Userservice interface {
	Allow(clientID string) bool
}

type UserserviceImpl struct {
	RLservice *RateLimiterService
	repo      *repository.UserRepoImpl
}

func NewUserserviceImpl(RLservice *RateLimiterService, repo *repository.UserRepoImpl) *UserserviceImpl {
	return &UserserviceImpl{
		RLservice: RLservice,
		repo:      repo,
	}
}

func (us *UserserviceImpl) Allow(clientID string) bool {
	return us.RLservice.Allow(clientID)
}

func (us *UserserviceImpl) AddClient(config model.ClientConfig) error {
	return us.repo.AddClient(config)
}

func (us *UserserviceImpl) DeleteClient(clientID string) error {
	return us.repo.DeleteClient(clientID)
}

func (us *UserserviceImpl) UpdateClient(config model.ClientConfig) error {
	return us.repo.UpdateClient(config)
}

func (us *UserserviceImpl) GetClients() error {
	return us.repo.GetClients()
}
