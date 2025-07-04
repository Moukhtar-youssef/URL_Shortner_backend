package routes

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"strings"

	"github.com/Moukhtar-youssef/URL_Shortner.git/internal/handlers"
	"github.com/Moukhtar-youssef/URL_Shortner.git/internal/middlewares"
	Storage "github.com/Moukhtar-youssef/URL_Shortner.git/internal/storage"
	"github.com/go-chi/chi/v5"
)

type Create struct {
	LongURL string `param:"long_url" query:"long_url" header:"long_url" json:"long_url" xml:"long_url" form:"long_url"`
}

func SetupRoutes(DB *Storage.URLDB, postlimiter, getlimiter *middlewares.Ratelimiter) *chi.Mux {
	router := chi.NewRouter()
	// router.With(middlewares.RateLimitMiddleware(getlimiter)).Get("/{id}", func(w http.ResponseWriter, r *http.Request) {
	// 	id := chi.URLParam(r, "id")
	// 	if id == "" {
	// 		http.Error(w, "Missing URL ID", http.StatusBadRequest)
	// 		return
	// 	}
	// 	url, err := handlers.GetLongURL(DB, id)
	// 	if err != nil {
	// 		http.Error(w, "URL not found", http.StatusNotFound)
	// 		return
	// 	}
	// 	w.Header().Set("Content-Type", "application/json")
	// 	w.WriteHeader(http.StatusOK)
	// 	json.NewEncoder(w).Encode(url)
	// })
	router.Get("/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if id == "" {
			http.Error(w, "Missing URL ID", http.StatusBadRequest)
			return
		}
		url, err := handlers.GetLongURL(DB, id)
		if err != nil {
			http.Error(w, "URL not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(url)
	})
	router.Post("/create", func(w http.ResponseWriter, r *http.Request) {
		var input Create

		err := parseRequest(r, &input)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
			return
		}

		if input.LongURL == "" {
			http.Error(w, "Missing long_url parameter", http.StatusBadRequest)
			return
		}

		shorturl, err := handlers.CreateShortURL(DB, input.LongURL)
		if err != nil {
			fmt.Println(err)
			http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(shorturl)
	})
	// router.With(middlewares.RateLimitMiddleware(postlimiter)).Post("/create", func(w http.ResponseWriter, r *http.Request) {
	// 	var input Create
	//
	// 	err := parseRequest(r, &input)
	// 	if err != nil {
	// 		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
	// 		return
	// 	}
	//
	// 	if input.LongURL == "" {
	// 		http.Error(w, "Missing long_url parameter", http.StatusBadRequest)
	// 		return
	// 	}
	//
	// 	shorturl, err := handlers.CreateShortURL(DB, input.LongURL)
	// 	if err != nil {
	// 		fmt.Println(err)
	// 		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
	// 		return
	// 	}
	// 	w.Header().Set("Content-Type", "application/json")
	// 	w.WriteHeader(http.StatusCreated)
	// 	json.NewEncoder(w).Encode(shorturl)
	// })
	return router
}

func parseRequest(r *http.Request, target any) error {
	contentType := r.Header.Get("Content-Type")

	if contentType == "application/json" {
		return json.NewDecoder(r.Body).Decode(target)
	}

	if contentType == "application/xml" || contentType == "text/xml" {
		return xml.NewDecoder(r.Body).Decode(target)
	}

	if contentType == "application/x-www-form-urlencoded" {
		if err := r.ParseForm(); err != nil {
			return err
		}
		return formToStruct(r.Form, target)
	}

	if r.URL.Query().Get("long_url") != "" {
		return formToStruct(r.URL.Query(), target)
	}
	if r.Header.Get("long_url") != "" {
		return headerToStruct(r.Header, target)
	}
	return fmt.Errorf("unsupported content type: %s", contentType)
}

func formToStruct(form url.Values, target any) error {
	val := reflect.ValueOf(target).Elem()
	typ := val.Type()

	for i := range val.NumField() {
		field := typ.Field(i)
		tag := field.Tag.Get("form")
		if tag == "" {
			tag = strings.ToLower(field.Name)
		}

		if values, ok := form[tag]; ok && len(values) > 0 {
			fieldValue := val.Field(i)
			if fieldValue.CanSet() {
				fieldValue.SetString(values[0])
			}
		}
	}
	return nil
}

func headerToStruct(header http.Header, target any) error {
	val := reflect.ValueOf(target).Elem()
	typ := val.Type()

	for i := range val.NumField() {
		field := typ.Field(i)
		tag := field.Tag.Get("header")
		if tag == "" {
			tag = strings.ToLower(field.Name)
		}

		if values, ok := header[tag]; ok && len(values) > 0 {
			fieldValue := val.Field(i)
			if fieldValue.CanSet() {
				fieldValue.SetString(values[0])
			}
		}
	}
	return nil
}
