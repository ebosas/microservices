package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/ebosas/microservices/internal/cache"
	"github.com/go-redis/redis/v8"
)

// handleAPICache handles an API call for cached messages
func handleAPICache(cr *redis.Client) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		data, err := cache.GetCacheJSON(cr)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("500 - Something happened"))

			log.Printf("get cache json: %s", err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, data)
	}
}
