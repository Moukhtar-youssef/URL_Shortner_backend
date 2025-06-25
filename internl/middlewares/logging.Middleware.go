package middlewares

import (
	"log"
	"net/http"
	"time"
)

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		lrw := &loggingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		defer func() {
			// Recover from panics and log them
			if err := recover(); err != nil {
				lrw.statusCode = http.StatusInternalServerError
				log.Printf("PANIC: %v", err)
			}

			// Log request details
			log.Printf(
				"%s %s %s %d %s %s %v",
				r.RemoteAddr,
				r.Method,
				r.URL.Path,
				lrw.statusCode,
				r.UserAgent(),
				r.Referer(),
				time.Since(start),
			)
		}()

		next.ServeHTTP(lrw, r)
	})
}

// loggingResponseWriter wraps http.ResponseWriter to capture status code
type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}
