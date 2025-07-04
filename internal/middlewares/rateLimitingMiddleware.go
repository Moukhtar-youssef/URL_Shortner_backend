package middlewares

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type Ratelimiter struct {
	redisClient *redis.Client
	rate        int
	window      time.Duration
}

func NewRateLimiter(redisAddr string, rate int, window time.Duration) *Ratelimiter {
	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr,
		DB:   1,
	})

	return &Ratelimiter{
		redisClient: rdb,
		rate:        rate,
		window:      window,
	}
}

func (rl *Ratelimiter) allow(ip string) (bool, time.Duration) {
	ctx := context.Background()
	key := "rate:" + ip

	// Check if key exists
	tokensStr, err := rl.redisClient.Get(ctx, key).Result()
	var tokens int
	if err == redis.Nil {
		// Not found, initialize
		tokens = rl.rate - 1
		rl.redisClient.Set(ctx, key, tokens, rl.window)
		return true, 0
	} else if err != nil {
		// Redis error
		return false, rl.window
	}

	tokens, _ = strconv.Atoi(tokensStr)
	if tokens <= 0 {
		ttl, _ := rl.redisClient.TTL(ctx, key).Result()
		return false, ttl
	}

	// Decrement and update TTL
	rl.redisClient.Decr(ctx, key)

	return true, 0
}

func RateLimitMiddleware(rl *Ratelimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := r.RemoteAddr

			allowed, retryAfter := rl.allow(ip)
			if !allowed {
				Error_msg := fmt.Sprintf("Too many requests. Try again in %s", formatRetryString(retryAfter))
				http.Error(w, Error_msg, http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func formatRetryString(window time.Duration) string {
	seconds := int(window.Seconds())
	if seconds <= 60 {
		return fmt.Sprintf("%ds", seconds)
	} else if seconds <= 3600 {
		return fmt.Sprintf("%dm", int(math.Round(window.Minutes())))
	} else if seconds <= 86400 {
		return fmt.Sprintf("%dh %dm", int(math.Round(window.Hours())), int(math.Round(window.Minutes()))%60)
	}
	return fmt.Sprintf("%dd %dh", int(math.Round(window.Hours()))/24, int(math.Round(window.Hours()))%24)
}
