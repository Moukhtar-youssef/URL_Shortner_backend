package main

import (
	"fmt"
	"log"
	"os"

	"github.com/Moukhtar-youssef/URL_Shortner.git/routes"
	Storage "github.com/Moukhtar-youssef/URL_Shortner.git/storage"
)

var DB *Storage.URLDB

func init() {
	user := os.Getenv("POSTGRES_USER")
	password := os.Getenv("POSTGRES_PASSWORD")
	host := os.Getenv("POSTGRES_HOST")
	pport := os.Getenv("POSTGRES_PORT")
	dbname := os.Getenv("POSTGRES_DB")
	Redis_host := os.Getenv("REDIS_HOST")
	// Usually sslmode=disable for local dev; adjust as needed
	sslmode := "disable"

	connURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", user, password, host, pport, dbname, sslmode)
	Redis_URL := fmt.Sprintf("%s:6379", Redis_host)
	fmt.Println(connURL)
	var err error
	DB, err = Storage.ConnectToDB(connURL, Redis_URL)
	if err != nil {
		log.Fatal(err)
	}
	err = DB.CreateTables()
	if err != nil {
		log.Fatal(err)
	}
	DB.DB.Exec(DB.Ctx, "INSERT INTO urls (short,long) VALUES ('http://localhost:8080/try','https://www.youtube.com') ON CONFLICT (short) DO NOTHING")
	eixts, _ := DB.CheckShortURLExists("http://localhost:8080/try")
	fmt.Println(eixts)
}

func main() {
	defer func() {
		err := DB.Close()
		if err != nil {
			log.Fatal(err)
		}
	}()
	echo_routes := routes.SetupRoutes(DB)
	if err := echo_routes.Start(":8081"); err != nil {
		echo_routes.Logger.Fatal(err)
	}
}
