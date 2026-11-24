package routes

import (
	"context"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/akhmadzaki/go-backend/errors"
	"github.com/google/uuid"
	"golang.org/x/time/rate"
)

// ── CORS Middleware ────────────────────────────────────────────────────────────

// corsMiddleware adds CORS headers to allow cross-origin requests from the
// frontend (e.g., localhost:5173 in dev, localhost:3000 in Docker).
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Max-Age", "86400")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// ── Request ID Middleware ──────────────────────────────────────────────────────

type contextKey string

const RequestIDKey contextKey = "request_id"

// requestIDMiddleware injects a unique request ID into the context and response.
func requestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-ID")
		if id == "" {
			id = uuid.New().String()
		}
		w.Header().Set("X-Request-ID", id)
		ctx := context.WithValue(r.Context(), RequestIDKey, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// ── Rate Limiting Middleware ───────────────────────────────────────────────────

// RateLimiter is a token-bucket rate limiter keyed by IP address.
type RateLimiter struct {
	mu       sync.Mutex
	clients  map[string]*rate.Limiter
	rate     rate.Limit
	burst    int
	cleanupT *time.Ticker
}

// NewRateLimiter creates a new RateLimiter with the given rate and burst.
// A cleanup goroutine runs every minute to remove stale entries.
func NewRateLimiter(r rate.Limit, burst int) *RateLimiter {
	rl := &RateLimiter{
		clients:  make(map[string]*rate.Limiter),
		rate:     r,
		burst:    burst,
		cleanupT: time.NewTicker(1 * time.Minute),
	}
	go rl.cleanup()
	return rl
}

func (rl *RateLimiter) cleanup() {
	for range rl.cleanupT.C {
		rl.mu.Lock()
		// Re-initialize the map to drop all stale entries
		rl.clients = make(map[string]*rate.Limiter)
		rl.mu.Unlock()
	}
}

// Stop stops the background cleanup goroutine.
func (rl *RateLimiter) Stop() {
	rl.cleanupT.Stop()
}

func (rl *RateLimiter) getLimiter(key string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limiter, exists := rl.clients[key]
	if !exists {
		limiter = rate.NewLimiter(rl.rate, rl.burst)
		rl.clients[key] = limiter
	}
	return limiter
}

// Middleware returns an HTTP middleware that enforces rate limiting per IP.
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract client IP (respect X-Forwarded-For behind proxy)
		ip := r.Header.Get("X-Forwarded-For")
		if ip == "" {
			ip = r.RemoteAddr
		}

		limiter := rl.getLimiter(ip)
		if !limiter.Allow() {
			log.Printf("RateLimiter: rate limit exceeded for %s", ip)
			errors.ErrRateLimited.WriteJSON(w)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// ── Recovery Middleware ────────────────────────────────────────────────────────

// recoveryMiddleware catches panics in HTTP handlers and returns a 500 error.
func recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("RecoveryMiddleware: panic recovered: %v", rec)
				http.Error(w, `{"error":"INTERNAL_SERVER_ERROR","message":"Internal server error"}`, http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
