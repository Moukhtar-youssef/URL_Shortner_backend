package middlewares

import (
	"sync"
	"time"
)

type visitor struct {
	tokens     int
	lastRefill time.Time
	mu         sync.Mutex
}

type ratelimiter struct {
	visitors map[string]*visitor
	rate     int
	window   time.Duration
	mu       sync.Mutex
}
