package handlers

import (
	"fmt"
	"math/rand/v2"
	"strings"
	"sync"
	"time"

	Storage "github.com/Moukhtar-youssef/URL_Shortner.git/internal/storage"
	"github.com/Moukhtar-youssef/URL_Shortner.git/internal/utils"
)

const (
	Alphabet      = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	NumberOfChrs  = 7
	maximum_tries = 3
	Try_delay     = 2 * time.Second
)

var AlphabetRunes = []rune(Alphabet)

var builderPool = sync.Pool{
	New: func() any {
		return new(strings.Builder)
	},
}

func Shortner(longurl string) (string, error) {
	err := utils.ValidateURL(longurl)
	if err != nil {
		return "", err
	}
	sb := builderPool.Get().(*strings.Builder)
	sb.Reset()
	for range NumberOfChrs {
		sb.WriteRune(AlphabetRunes[rand.IntN(len(AlphabetRunes))])
	}
	result := sb.String()
	builderPool.Put(sb)
	return result, nil
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
		return fmt.Errorf("the provided url isn't found")
	}
	for range 3 {
		err := DB.DeleteURL(shorturl)
		if err == nil {
			return nil
		}
	}

	return fmt.Errorf("failed to delte shorturl after %d tries", maximum_tries)
}

func GetLongURL(DB *Storage.URLDB, shorturl string) (string, error) {
	exists, err := DB.CheckShortURLExists(shorturl)
	if err != nil {
		return "", err
	}
	if !exists {
		return "", fmt.Errorf("there is no url associated with this short url")
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
