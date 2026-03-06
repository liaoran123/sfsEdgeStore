package server

import (
	"net/http"
	"sync"
	"time"
)

type RateLimiter struct {
	requests map[string]*clientRequest
	mu       sync.Mutex
	limit    int
	window   time.Duration
}

type clientRequest struct {
	count     int
	resetTime time.Time
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		requests: make(map[string]*clientRequest),
		limit:    limit,
		window:   window,
	}
}

func (rl *RateLimiter) Allow(clientID string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	req, exists := rl.requests[clientID]

	if !exists || now.After(req.resetTime) {
		rl.requests[clientID] = &clientRequest{
			count:     1,
			resetTime: now.Add(rl.window),
		}
		return true
	}

	if req.count >= rl.limit {
		return false
	}

	req.count++
	return true
}

func (rl *RateLimiter) Middleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clientID := getClientID(r)
		if !rl.Allow(clientID) {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Retry-After", "60")
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"error": "Rate limit exceeded"}`))
			return
		}
		next(w, r)
	}
}

func getClientID(r *http.Request) string {
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		return forwarded
	}
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}
	return r.RemoteAddr
}
