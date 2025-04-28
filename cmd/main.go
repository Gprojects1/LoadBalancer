package main

import (
	"LoadBalancer/TimeLimiter/config"
	"LoadBalancer/TimeLimiter/pkg/controller"
	"LoadBalancer/TimeLimiter/pkg/database"
	"LoadBalancer/TimeLimiter/pkg/repository"
	"LoadBalancer/TimeLimiter/pkg/service"
	"time"

	lbCon "LoadBalancer/Balancer/pkg/controller"
	"LoadBalancer/Balancer/pkg/health"
	lbServ "LoadBalancer/Balancer/pkg/service"
	"LoadBalancer/Balancer/pkg/utils"
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/gorilla/mux"
)

func main() {
	var serverList string
	var port int
	flag.StringVar(&serverList, "backends", "", "Load balanced backends, use commas to separate")
	flag.IntVar(&port, "port", 3030, "Port to serve")
	flag.Parse()

	if len(serverList) == 0 {
		log.Fatal("Please provide one or more backends to load balance")
	}

	Backends := strings.Split(serverList, ",")
	serverPool := &lbServ.ServerPool{}
	con := lbCon.NewLoadBlancerController(serverPool)
	for _, backend := range Backends {
		backendUrl, err := url.Parse(backend)
		if err != nil {
			log.Fatalf("Failed to parse backend URL %s: %v", backend, err)
		}

		proxy := httputil.NewSingleHostReverseProxy(backendUrl)
		proxy.ErrorHandler = func(writer http.ResponseWriter, request *http.Request, e error) {
			log.Printf("[%s] %s\n", backendUrl.Host, e.Error())
			retries := utils.GetRetryFromContext(request)
			if retries < 3 {
				select {
				case <-time.After(10 * time.Millisecond):
					ctx := context.WithValue(request.Context(), utils.Retry, retries+1)
					proxy.ServeHTTP(writer, request.WithContext(ctx))
				}
				return
			}

			serverPool.MarkBackendStatus(backendUrl, false)

			attempts := utils.GetAttemptsFromContext(request)
			log.Printf("%s(%s) Attempting retry %d\n", request.RemoteAddr, request.URL.Path, attempts)
			ctx := context.WithValue(request.Context(), utils.Attempts, attempts+1)
			con.BalanceRequest(writer, request.WithContext(ctx))
		}

		con.AddNewBackend(&lbServ.Backend{
			URL:          backendUrl,
			Alive:        true,
			ReverseProxy: proxy,
		})
	}

	lbService := serverPool
	lbController := lbCon.NewLoadBlancerController(lbService)

	healthChecker := health.NewLHealthChecker(lbService)
	go healthChecker.HealthCheck()

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}
	defer db.Close()

	err = database.Migrate(db)
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	userRepo := repository.NewUserRepoImpl(db)
	tlRepo := repository.NewRlRepoImpl(userRepo)

	rl, err := service.NewRateLimiter(tlRepo)
	if err != nil {
		log.Fatalf("Failed to make RateLimiter: %v", err)
	}
	defer rl.StopRefill()
	userService := service.NewUserserviceImpl(rl, userRepo)
	router := mux.NewRouter()
	handler := controller.NewUserControllerImpl(userService, lbController)

	router.HandleFunc("/clients", handler.AddClient).Methods("POST")
	router.HandleFunc("/clients/{client_id}", handler.DeleteClient).Methods("DELETE")
	router.HandleFunc("/clients", handler.UpdateClient).Methods("PUT")
	router.HandleFunc("/", handler.CheckRateLimit).Methods("GET")

	serverAddr := fmt.Sprintf(":%d", cfg.Port)
	log.Printf("Starting server on %s\n", serverAddr)

	server := &http.Server{
		Addr:         serverAddr,
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}

	log.Println("Server stopped.")
}
