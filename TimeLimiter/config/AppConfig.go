package config

import (
	"log"
	"os"
	"strconv"
)

type Config struct {
	DatabaseURL string
	Port        int
}

func LoadConfig() (*Config, error) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:password@postgres:5432/timelimiter?sslmode=disable"
		log.Println("Warning: DATABASE_URL not set. Using default:", dbURL)
	}

	portStr := os.Getenv("PORT")
	if portStr == "" {
		portStr = "3030"
		log.Println("Warning: PORT not set. Using default:", portStr)
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, err
	}

	return &Config{
		DatabaseURL: dbURL,
		Port:        port,
	}, nil
}
