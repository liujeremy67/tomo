package middleware

import (
	"log"
	"net/http"
	"time"
)

// this function takes one argument called next
// returns another handler, which is a function HTTP server can call
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()  // record time before handling request
		next.ServeHTTP(w, r) // pass request down the chain to next handler

		log.Printf("%s %s %v", r.Method, r.URL.Path, time.Since(start)) // after it returns, log method, path, duration
	})
}
