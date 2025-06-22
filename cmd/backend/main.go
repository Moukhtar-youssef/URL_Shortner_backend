package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Moukhtar-youssef/URL_Shortner.git/internl/middlewares"
	"github.com/Moukhtar-youssef/URL_Shortner.git/internl/routes"
	Storage "github.com/Moukhtar-youssef/URL_Shortner.git/internl/storage"
)

var DB *Storage.URLDB

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
}

func main() {
	Setup()
	defer func() {
		err := DB.Close()
		if err != nil {
			log.Fatal(err)
		}
	}()
	// rl := middlewares.NewRateLimiter(20, 10*time.Minute)

	mux := routes.SetupRoutes(DB)
	handler := middlewares.Chain(
		middlewares.LoggingMiddleware,
		middlewares.EnableCORS,
		// middlewares.RateLimitMiddleware(rl),
	)
	server := &http.Server{
		Addr:         ":8081",
		Handler:      handler(mux),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}
	log.Printf("Starting server on %s", ":8081")
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
