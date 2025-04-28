package health

import (
	"LoadBalancer/Balancer/pkg/service"
	"log"
	"time"
)

type HealthChecker interface {
	HealthCheck()
}

type HealthCheckerImpl struct {
	service service.LoadBlancerService
}

func NewLHealthChecker(service service.LoadBlancerService) HealthChecker {
	return &HealthCheckerImpl{
		service: service,
	}
}

func (ch *HealthCheckerImpl) HealthCheck() {
	t := time.NewTicker(time.Minute * 2)
	for {
		select {
		case <-t.C:
			log.Println("Starting health check...")
			ch.service.HealthCheck()
			log.Println("Health check completed")
		}
	}
}
