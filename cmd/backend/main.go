package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Moukhtar-youssef/URL_Shortner.git/internal/middlewares"
	"github.com/Moukhtar-youssef/URL_Shortner.git/internal/routes"
	Storage "github.com/Moukhtar-youssef/URL_Shortner.git/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

var (
	DB                 *Storage.URLDB
	GetURLRateLimit    *middlewares.Ratelimiter
	CreateURLRateLimit *middlewares.Ratelimiter
)

func Setup() {
	user := os.Getenv("POSTGRES_USER")
	password := os.Getenv("POSTGRES_PASSWORD")
	host := os.Getenv("POSTGRES_HOST")
	pport := os.Getenv("POSTGRES_PORT")
	dbname := os.Getenv("POSTGRES_DB")
	Redis_host := os.Getenv("REDIS_HOST")
	Redis_port := os.Getenv("REDIS_PORT")
	if Redis_port == "" {
		Redis_port = "6379"
	}
	sslmode := "disable"

	connURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", user, password, host, pport, dbname, sslmode)
	Redis_URL := fmt.Sprintf("%s:%s", Redis_host, Redis_port)
	var err error
	DB, err = Storage.ConnectToDB(connURL, Redis_URL)
	if err != nil {
		log.Fatal(err)
	}
	err = DB.CreateTables()
	if err != nil {
		log.Fatal(err)
	}
	GetURLRateLimit = middlewares.NewRateLimiter(Redis_URL, 1000000000000000000, time.Minute)
	CreateURLRateLimit = middlewares.NewRateLimiter(Redis_URL, 1000000000000000000, time.Minute)
}

func main() {
	middlewares.StartAsyncStreamLogger(1000)

	Setup()

	defer func() {
		err := DB.Close()
		if err != nil {
			log.Fatal(err)
		}
		middlewares.StopAsyncStreamLogger()
	}()

	router := chi.NewRouter()

	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.CleanPath)
	router.Use(middlewares.FileLoggingMiddleware)
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))
	router.Use(middleware.Recoverer)

	router.Mount("/api", routes.SetupRoutes(DB, CreateURLRateLimit, GetURLRateLimit))

	server := &http.Server{
		Addr:         ":8081",
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Println("Server running on http://localhost:8081")
	server.ListenAndServe()
}
