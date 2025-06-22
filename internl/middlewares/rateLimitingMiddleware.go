package middlewares

import (
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

func (rl *Ratelimiter) allow(ip string) bool {
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
		return false
	}

	v.tokens--
	return true
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

			if !rl.allow(ip) {
				http.Error(w, "Too many Requests", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
