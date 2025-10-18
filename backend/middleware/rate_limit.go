package middleware

import (
	"net/http"
	"sync"
	"time"
)

var (
	lastRequest = make(map[string]time.Time)
	mu          sync.Mutex
)

func RateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr
		mu.Lock()
		defer mu.Unlock()
		if last, ok := lastRequest[ip]; ok && time.Since(last) < time.Second {
			http.Error(w, "Too many requests", http.StatusTooManyRequests)
			return
		}
		lastRequest[ip] = time.Now()
		next.ServeHTTP(w, r)
	})
}
