package main

import (
	"errors"
	"log"

	"github.com/Moukhtar-youssef/URL_Shortner.git/routes"
	Storage "github.com/Moukhtar-youssef/URL_Shortner.git/storage"
)

var (
	DB                  *Storage.URLDB
	ErrAlreadyShortened = errors.New("this URL is already shortened")
)

func init() {
	var err error
	DB, err = Storage.ConnectToDB("./Test.db", "localhost:6379")
	if err != nil {
		log.Fatal(err)
	}
	err = DB.CreateTable()
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	defer func() {
		err := DB.Close()
		if err != nil {
			log.Fatal(err)
		}
	}()
	echo_routes := routes.SetupRoutes(DB)
	if err := echo_routes.Start(":8080"); err != nil {
		echo_routes.Logger.Fatal(err)
	}
}
