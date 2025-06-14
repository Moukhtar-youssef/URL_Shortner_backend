package handlers

import (
	"fmt"
	"math/rand/v2"
	"net/url"
	"os"
	"strings"
	"time"

	customerrors "github.com/Moukhtar-youssef/URL_Shortner.git/internl/custom_errors"
	Storage "github.com/Moukhtar-youssef/URL_Shortner.git/internl/storage"
)

const (
	Alphabet      = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	NumberOfChrs  = 7
	maximum_tries = 3
	Try_delay     = 2 * time.Second
)

var baseURL = os.Getenv("BASE_URL")

var AlphabetRunes = []rune(Alphabet)

func urlValidator(longurl string) error {
	u, err := url.Parse(longurl)
	if err != nil {
		return customerrors.ErrInvalidLongURL
	}
	path := strings.Trim(u.Path, "/")
	if len(path) == 0 {
		return customerrors.ErrInvalidLongURL
	}
	if len(path) > 0 && len(path) <= 8 {
		return customerrors.ErrAlreadyShortened
	}
	return nil
}

func Shortner(longurl string) (string, error) {
	err := urlValidator(longurl)
	if err != nil {
		return "", err
	}
	var ShortURLCode strings.Builder
	for range NumberOfChrs {
		ShortURLCode.WriteRune(AlphabetRunes[rand.IntN(len(AlphabetRunes))])
	}
	return fmt.Sprintf("%v/%v", baseURL, ShortURLCode.String()), nil
}

func CreateShortURL(DB *Storage.URLDB, longurl string) (string, error) {
	const maxGenerateAttmept = 10
	for range maxGenerateAttmept {
		ShortURL, err := Shortner(longurl)
		if err != nil {
			return "", err
		}
		exists, err := DB.CheckShortURLExists(ShortURL)
		if err != nil {
			return "", err
		}
		if exists {
			continue
		}
		for range maximum_tries {
			err := DB.SaveURL(ShortURL, longurl)
			if err == nil {
				return ShortURL, nil
			}
			time.Sleep(Try_delay)
		}

	}
	return "", fmt.Errorf("failed to create short URL after %d attempts", maxGenerateAttmept)
}

func DeleteShortURL(DB *Storage.URLDB, shorturl string) error {
	exists, err := DB.CheckShortURLExists(shorturl)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("The provided url isn't found")
	}
	type result struct {
		err error
	}
	results := make(chan result, maximum_tries)
	for range maximum_tries {
		DB.Wg.Add(1)
		go func() {
			defer DB.Wg.Done()
			err := DB.DeleteURL(shorturl)
			results <- result{err: err}
		}()
		time.Sleep(Try_delay)
	}
	DB.Wg.Wait()
	close(results)

	for res := range results {
		if res.err == nil {
			return nil // deletion successful
		}
	}

	return fmt.Errorf("Failed to delte shorturl after %d tries", maximum_tries)
}

func GetLongURL(DB *Storage.URLDB, shorturl string) (string, error) {
	exists, err := DB.CheckShortURLExists(shorturl)
	if err != nil {
		return "", err
	}
	if !exists {
		return "", fmt.Errorf("there is no url associated with this long url")
	}
	longURL, err := DB.GetURL(shorturl)
	if err != nil {
		return "", fmt.Errorf("there is no url associated with this short url")
	}
	return longURL, nil
}

func EditLongURL(DB *Storage.URLDB, shorturl string, newlong string) (string, error) {
	exists, err := DB.CheckShortURLExists(shorturl)
	if err != nil {
		return "", err
	}
	if !exists {
		return "", fmt.Errorf("there is no url associated with this long url")
	}
	err = DB.EditURL(shorturl, newlong)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Edited the long url associated with : %s", shorturl), nil
}
