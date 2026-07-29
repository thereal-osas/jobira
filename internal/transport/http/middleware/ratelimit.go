package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/rodrigueghenda/jobira/internal/transport/http/response"
)

type visitor struct {
	request   int
	lastSeen  time.Time
	resetTime time.Time
}

type RateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*visitor
	limit    int
	window   time.Duration
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		limit:    limit,
		window:   window,
	}

	go rl.cleanup()

	return rl
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr

		rl.mu.Lock()

		v, exists := rl.visitors[ip]
		now := time.Now()

		if !exists || now.After(v.resetTime) {
			rl.visitors[ip] = &visitor{
				request:   1,
				lastSeen:  now,
				resetTime: now.Add(rl.window),
			}

			rl.mu.Unlock()
			next.ServeHTTP(w, r)
			return
		}

		v.request++
		v.lastSeen = now

		if v.request > rl.limit {
			rl.mu.Unlock()
			response.Error(w, http.StatusTooManyRequests, "too many requests")
			return
		}

		rl.mu.Unlock()
		next.ServeHTTP(w, r)
	})
}

func (rl *RateLimiter) cleanup() {
	for {
		time.Sleep(time.Minute)

		rl.mu.Lock()

		for ip, v := range rl.visitors {
			if time.Since(v.lastSeen) > rl.window {
				delete(rl.visitors, ip)
			}
		}

		rl.mu.Unlock()
	}
}
