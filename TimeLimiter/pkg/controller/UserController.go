package controller

import (
	"LoadBalancer/TimeLimiter/pkg/model"
	"LoadBalancer/TimeLimiter/pkg/service"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"LoadBalancer/Balancer/pkg/controller"

	"github.com/gorilla/mux"
)

type UserController interface {
	CheckRateLimit(w http.ResponseWriter, r *http.Request)
	AddClient(w http.ResponseWriter, r *http.Request)
	DeleteClient(w http.ResponseWriter, r *http.Request)
	UpdateClient(w http.ResponseWriter, r *http.Request)
}

type UserControllerImpl struct {
	userSevice   *service.UserserviceImpl
	LBcontroller controller.LoadBalancerController
}

func NewUserControllerImpl(userSevice *service.UserserviceImpl, LBcontroller controller.LoadBalancerController) *UserControllerImpl {
	return &UserControllerImpl{
		userSevice:   userSevice,
		LBcontroller: LBcontroller,
	}
}

func (con *UserControllerImpl) AddClient(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	log.Printf("AddClient: Request received at %s", startTime.Format(time.RFC3339))

	var config model.ClientConfig
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		log.Printf("AddClient: Error decoding request body: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := con.userSevice.AddClient(config); err != nil {
		log.Printf("AddClient: Error adding client to repository: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintln(w, "Client added successfully")

	log.Printf("AddClient: Client added successfully, status code: %d, duration: %v", http.StatusCreated, time.Since(startTime))
}

func (con *UserControllerImpl) DeleteClient(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	log.Printf("DeleteClient: Request received at %s", startTime.Format(time.RFC3339))

	vars := mux.Vars(r)
	clientID := vars["client_id"]

	if err := con.userSevice.DeleteClient(clientID); err != nil {
		log.Printf("DeleteClient: Error deleting client from repository: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "Client deleted successfully")

	log.Printf("DeleteClient: Client deleted successfully, status code: %d, duration: %v", http.StatusOK, time.Since(startTime))
}

func (con *UserControllerImpl) UpdateClient(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	log.Printf("UpdateClient: Request received at %s", startTime.Format(time.RFC3339))

	var config model.ClientConfig
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		log.Printf("UpdateClient: Error decoding request body: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := con.userSevice.UpdateClient(config); err != nil {
		log.Printf("UpdateClient: Error updating client in repository: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "Client updated successfully")

	log.Printf("UpdateClient: Client updated successfully, status code: %d, duration: %v", http.StatusOK, time.Since(startTime))
}

func (con *UserControllerImpl) CheckRateLimit(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	log.Printf("CheckRateLimit: Request received at %s", startTime.Format(time.RFC3339))

	clientID := r.URL.Query().Get("client_id")

	if clientID == "" {
		log.Println("CheckRateLimit: client_id is missing in the query parameters")
		http.Error(w, "client_id is required", http.StatusBadRequest)
		return
	}

	allowed := con.userSevice.Allow(clientID)
	if allowed {
		log.Printf("CheckRateLimit: Request allowed for client_id: %s", clientID)
		con.LBcontroller.BalanceRequest(w, r)
		log.Printf("CheckRateLimit: Request balanced, status code: %d, duration: %v", http.StatusOK, time.Since(startTime))
	} else {
		log.Printf("CheckRateLimit: Rate limit exceeded for client_id: %s", clientID)
		w.WriteHeader(http.StatusTooManyRequests)
		fmt.Fprintln(w, "Rate limit exceeded")
		log.Printf("CheckRateLimit: Rate limit exceeded, status code: %d, duration: %v", http.StatusTooManyRequests, time.Since(startTime))
	}
}
