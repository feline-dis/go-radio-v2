package middleware

import (
	"log"
	"net/http"
	"os"
	"time"
)

var (
	// logger is a custom logger that includes timestamps and writes to stdout immediately
	logger = log.New(os.Stdout, "", log.LstdFlags|log.Lmicroseconds)
)

// LoggingMiddleware creates a middleware that logs HTTP request details
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Log the incoming request
		log.Printf("[DEBUG] LoggingMiddleware: Incoming request: %s %s", r.Method, r.URL.Path)

		// Create a custom response writer to capture the status code
		rw := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		// Process the request
		log.Printf("[DEBUG] LoggingMiddleware: Calling next handler")
		next.ServeHTTP(rw, r)
		log.Printf("[DEBUG] LoggingMiddleware: Next handler completed")

		// Calculate duration
		duration := time.Since(start)

		// Log the request details with a more structured format
		log.Printf("[DEBUG] LoggingMiddleware: Request completed: %s %s %d %s %s %s",
			r.Method,
			r.URL.Path,
			rw.statusCode,
			duration.Round(time.Millisecond),
			r.RemoteAddr,
			r.UserAgent(),
		)
	})
}

// responseWriter is a custom response writer that captures the status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader captures the status code before writing the header
func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
