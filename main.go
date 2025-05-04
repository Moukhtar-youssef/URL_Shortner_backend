package main

import (
	"fmt"
	"log"

	Storage "github.com/Moukhtar-youssef/URL_Shortner.git/storage"
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
	err = DB.SaveURL("localhost:8080/644", "https://chatgpt.com/c/68179ff0-bc1c-8002-b8b5-80562e6a02e0")
	if err != nil {
		log.Fatal(err)
	}
	LongURL, err := DB.GetURL("localhost:8080/644")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(LongURL)
	err = DB.DeleteURL("localhost:8080/644")
	if err != nil {
		log.Fatal(err)
	}
}
