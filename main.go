package main

import (
	"fmt"
	"log"
	"time"

	"github.com/Moukhtar-youssef/URL_Shortner.git/handlers"
	"github.com/Moukhtar-youssef/URL_Shortner.git/middleware"
	"github.com/Moukhtar-youssef/URL_Shortner.git/routes"
	Storage "github.com/Moukhtar-youssef/URL_Shortner.git/storage"
	"github.com/labstack/echo/v4"
)

func main() {
	DB, err := Storage.ConnectToDB("./test.db", "localhost:6379")
	if err != nil {
		log.Fatal(err)
	}
	err = DB.CreateTable()
	if err != nil {
		log.Fatal(err)
	}
	shortstring, err := handlers.CreateShortURL(DB, "http://www.youtube.com")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(shortstring)
	e := echo.New()
	e.Use(middleware.AllowRequests(10, 1*time.Minute, 20, 1*time.Minute))
	routes.SetupRoutes(DB, e)
	if err := e.Start(":8080"); err != nil {
		e.Logger.Fatal(err)
	}
}
