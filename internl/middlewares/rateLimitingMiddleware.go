package middlewares

import (
	"fmt"
	"math"
	"net"
	"net/http"
	"sync"
	"time"
)

type visitor struct {
	tokens     int
	lastRefill time.Time
	mu         sync.Mutex
}

type Ratelimiter struct {
	visitors map[string]*visitor
	rate     int
	window   time.Duration
	mu       sync.Mutex
}

func NewRateLimiter(rate int, window time.Duration) *Ratelimiter {
	rl := &Ratelimiter{
		visitors: make(map[string]*visitor),
		rate:     rate,
		window:   window,
	}
	go rl.cleanup()
	return rl
}

func (rl *Ratelimiter) getVisitor(ip string) *visitor {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, exists := rl.visitors[ip]
	if !exists {
		v = &visitor{
			tokens:     rl.rate,
			lastRefill: time.Now(),
		}
		rl.visitors[ip] = v
	}
	return v
}

func (rl *Ratelimiter) allow(ip string) (bool, time.Duration) {
	v := rl.getVisitor(ip)

	v.mu.Lock()
	defer v.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(v.lastRefill)
	if elapsed >= rl.window {
		v.tokens = rl.rate
		v.lastRefill = now
	}

	if v.tokens <= 0 {
		return false, elapsed
	}

	v.tokens--
	return true, elapsed
}

func (rl *Ratelimiter) cleanup() {
	for {
		time.Sleep(1 * time.Minute)
		rl.mu.Lock()
		now := time.Now()
		for ip, v := range rl.visitors {
			v.mu.Lock()
			if now.Sub(v.lastRefill) > 5*time.Minute {
				delete(rl.visitors, ip)
			}
			v.mu.Unlock()
		}
		rl.mu.Unlock()
	}
}

func getIP(r *http.Request) string {
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		return ip
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func RateLimitMiddleware(rl *Ratelimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := getIP(r)

			allowed, elapsed := rl.allow(ip)
			remaining := rl.window - elapsed
			retry := formateRetryString(remaining)
			if !allowed {
				Error_msg := fmt.Sprintf("Too many Requests please try again after %s", retry)
				http.Error(w, Error_msg, http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func formateRetryString(window time.Duration) string {
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
