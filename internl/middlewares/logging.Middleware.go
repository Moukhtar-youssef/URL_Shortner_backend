package middlewares

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

var (
	filelogger *log.Logger
	logChannel chan string
	done       chan struct{}
)

func StartAsyncStreamLogger(buffersize int) {
	logChannel = make(chan string, buffersize)
	done = make(chan struct{})
	logfile, err := os.OpenFile("Stream.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Error Opening log file: %v", err)
	}
	filelogger = log.New(logfile, "", log.LstdFlags|log.Lshortfile)

	go func() {
		for {
			select {
			case logEntry := <-logChannel:
				filelogger.Println(logEntry)
			case <-done:
				logfile.Close()
				return
			}
		}
	}()
}

func StopAsyncStreamLogger() {
	close(done)
}

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

func FileLoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		lrw := &loggingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		defer func() {
			if err := recover(); err != nil {
				lrw.statusCode = http.StatusInternalServerError
				log.Printf("PANIC: %v", err)
			}

			logEntry := fmt.Sprintf(
				"%s %s %s %d %s %s %v",
				r.RemoteAddr,
				r.Method,
				r.URL.Path,
				lrw.statusCode,
				r.UserAgent(),
				r.Referer(),
				time.Since(start),
			)
			select {
			case logChannel <- logEntry:
			default:
			}
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
