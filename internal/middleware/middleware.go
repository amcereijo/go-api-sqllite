package middleware

import (
	"log"
	"net/http"
	"time"
)

// load env variable
var DEBUG = false

func printHeaders(r *http.Request) {
	if DEBUG {
		log.Printf("Handling CORS for request from %s", r.RemoteAddr)

		for name, values := range r.Header {
			for _, value := range values {
				log.Printf("Header: %s = %s", name, value)
			}
		}
	}
}

// CORS middleware to handle CORS requests
func OptionsCors(w http.ResponseWriter, r *http.Request) {
	// read all headers
	printHeaders(r)

	w.Header().Set("Access-Control-Allow-Origin", "*")
	// w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Methods", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("Access-Control-Max-Age", "86400")

	// Handle preflight requests
	w.WriteHeader(http.StatusOK)
}

func CorsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		printHeaders(r)

		w.Header().Set("Access-Control-Allow-Origin", "*")
		// w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Methods", "*")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		w.Header().Set("Access-Control-Max-Age", "86400")

		next.ServeHTTP(w, r)
	})
}

// Logger is a middleware that logs HTTP requests
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf(
			"%s %s %s",
			r.Method,
			r.RequestURI,
			time.Since(start),
		)
	})
}
