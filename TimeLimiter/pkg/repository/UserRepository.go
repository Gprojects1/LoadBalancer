package repository

import (
	"LoadBalancer/TimeLimiter/pkg/model"
	"database/sql"
	"fmt"
	"log"
	"sync"
)

type UserRepo interface {
	GetClients() error
	GetBuckets() map[string]*model.TokenBucket
	AddClient(config model.ClientConfig) error
	DeleteClient(clientID string) error
	UpdateClient(config model.ClientConfig) error
}

type UserRepoImpl struct {
	db      *sql.DB
	Mutex   sync.Mutex
	Buckets map[string]*model.TokenBucket
}

func NewUserRepoImpl(db *sql.DB) *UserRepoImpl {
	log.Println("Creating new UserRepoImpl")
	return &UserRepoImpl{
		db:      db,
		Mutex:   sync.Mutex{},
		Buckets: make(map[string]*model.TokenBucket),
	}
}

func (r *UserRepoImpl) AddClient(config model.ClientConfig) error {
	r.Mutex.Lock()

	log.Printf("Attempting to add client with ID %s", config.ClientID)
	if _, ok := r.Buckets[config.ClientID]; ok {
		err := fmt.Errorf("client with ID %s already exists", config.ClientID)
		log.Printf("Client with ID %s already exists: %v", config.ClientID, err)
		return err
	}

	_, err := r.db.Exec(
		"INSERT INTO clients (client_id, capacity, rate_per_sec) VALUES ($1, $2, $3)",
		config.ClientID, config.Capacity, config.RatePerSec,
	)
	if err != nil {
		err = fmt.Errorf("failed to insert client into DB: %w", err)
		log.Printf("Failed to insert client with ID %s into DB: %v", config.ClientID, err)
		return err
	}

	log.Printf("Successfully inserted client with ID %s into DB", config.ClientID)

	r.Buckets[config.ClientID] = model.NewTokenBucket(
		config.ClientID,
		config.Capacity,
		config.RatePerSec,
	)
	r.Mutex.Unlock()
	log.Printf("Client with ID %s added successfully", config.ClientID)
	return nil
}

func (r *UserRepoImpl) DeleteClient(clientID string) error {
	r.Mutex.Lock()

	log.Printf("Attempting to delete client with ID %s", clientID)

	_, err := r.db.Exec("DELETE FROM clients WHERE client_id = $1", clientID)
	if err != nil {
		err = fmt.Errorf("failed to delete client from DB: %w", err)
		log.Printf("Failed to delete client with ID %s from DB: %v", clientID, err)
		return err
	}

	log.Printf("Successfully deleted client with ID %s from DB", clientID)
	delete(r.Buckets, clientID)
	r.Mutex.Unlock()
	log.Printf("Client with ID %s deleted successfully", clientID)
	return nil
}

func (r *UserRepoImpl) UpdateClient(config model.ClientConfig) error {
	r.Mutex.Lock()

	log.Printf("Attempting to update client with ID %s", config.ClientID)

	_, err := r.db.Exec(
		"UPDATE clients SET capacity = $2, rate_per_sec = $3 WHERE client_id = $1",
		config.ClientID, config.Capacity, config.RatePerSec,
	)
	if err != nil {
		err = fmt.Errorf("failed to update client in DB: %w", err)
		log.Printf("Failed to update client with ID %s in DB: %v", config.ClientID, err)
		return err
	}

	log.Printf("Successfully updated client with ID %s in DB", config.ClientID)

	if bucket, exists := r.Buckets[config.ClientID]; exists {
		bucket.SetCapacity(config.Capacity)
		bucket.SetRate(config.RatePerSec)
	} else {
		r.Buckets[config.ClientID] = model.NewTokenBucket(
			config.ClientID,
			config.Capacity,
			config.RatePerSec,
		)
	}
	r.Mutex.Unlock()
	log.Printf("Client with ID %s updated successfully", config.ClientID)
	return nil
}

func (r *UserRepoImpl) GetClients() error {
	r.Mutex.Lock()
	defer r.Mutex.Unlock()

	log.Println("Attempting to retrieve clients from DB")

	rows, err := r.db.Query("SELECT client_id, capacity, rate_per_sec FROM clients")
	if err != nil {
		err = fmt.Errorf("failed to query clients: %w", err)
		log.Printf("Failed to query clients from DB: %v", err)
		return err
	}
	defer rows.Close()

	newBuckets := make(map[string]*model.TokenBucket)
	for rows.Next() {
		var clientID string
		var capacity int
		var ratePerSec float64
		if err := rows.Scan(&clientID, &capacity, &ratePerSec); err != nil {
			err = fmt.Errorf("failed to scan client: %w", err)
			log.Printf("Failed to scan client row: %v", err)
			return err
		}

		newBuckets[clientID] = model.NewTokenBucket(clientID, capacity, ratePerSec)
		log.Printf("Loaded client %s from DB", clientID)
	}

	if err := rows.Err(); err != nil {
		err = fmt.Errorf("error iterating over rows: %w", err)
		log.Printf("Error iterating over client rows: %v", err)
		return err
	}

	r.Buckets = newBuckets
	log.Printf("Successfully retrieved and cached %d clients from DB", len(r.Buckets))
	return nil
}

func (repo *UserRepoImpl) GetBuckets() map[string]*model.TokenBucket {
	return repo.Buckets
}
