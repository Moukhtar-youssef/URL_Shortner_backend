package handlers

import (
	"fmt"
	"math/rand/v2"
	"strings"
	"time"

	Storage "github.com/Moukhtar-youssef/URL_Shortner.git/storage"
)

const (
	Alphabet      = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	NumberOfChrs  = 7
	baseURL       = "http://localhost:8080"
	maximum_tries = 3
	Try_delay     = 2 * time.Second
)

var AlphabetRunes = []rune(Alphabet)

func CreateShortURL(DB *Storage.URLDB, longurl string) (string, error) {
	for {
		var ShortURLCode strings.Builder
		for range NumberOfChrs {
			ShortURLCode.WriteRune(AlphabetRunes[rand.IntN(len(AlphabetRunes))])
		}
		ShortURL := fmt.Sprintf("%v/%v", baseURL, ShortURLCode.String())
		exists, err := DB.CheckShortURLExists(ShortURL)

		if err != nil {
			return "", err
		}
		if !exists {
			fmt.Println(ShortURL)
			go func(ShortURL, longurl string) {
				var saveErr error
				for range maximum_tries {
					saveErr = DB.SaveURL(ShortURL, longurl)
					if saveErr == nil {
						return
					}
					time.Sleep(Try_delay)
				}
			}(ShortURL, longurl)
			return ShortURL, nil
		}

	}
}
func DeleteShortURL(DB *Storage.URLDB, shorturl string) error {
	exists, err := DB.CheckShortURLExists(shorturl)
	if err != nil {
		return err
	}
	if exists {
		go func(shorturl string) {
			for range maximum_tries {
				var saveErr error
				saveErr = DB.DeleteURL(shorturl)
				if saveErr == nil {
					return
				}
				time.Sleep(Try_delay)
			}
		}(shorturl)
		return nil
	}
	return fmt.Errorf("The provided url isn't found")
}
func GetLongURL(DB *Storage.URLDB, shorturl string) (string, error) {
	exists, err := DB.CheckShortURLExists(shorturl)
	if err != nil {
		return "", err
	}
	if exists {
		longURL, err := DB.GetURL(shorturl)
		if err != nil {
			return "", err
		}
		return longURL, nil
	}
	return "", fmt.Errorf("there is no url associated with this short url")
}
