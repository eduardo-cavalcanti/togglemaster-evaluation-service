package main

import (
	"log"
	"net/http"
	"time"
)

type accessLogWriter struct {
	http.ResponseWriter
	status int
}

func (w *accessLogWriter) WriteHeader(status int) {
	if w.status != 0 {
		return
	}
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *accessLogWriter) Write(body []byte) (int, error) {
	if w.status == 0 {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(body)
}

func (w *accessLogWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" {
			next.ServeHTTP(w, r)
			return
		}
		started := time.Now()
		writer := &accessLogWriter{ResponseWriter: w}
		next.ServeHTTP(writer, r)
		status := writer.status
		if status == 0 {
			status = http.StatusOK
		}
		// Keep credentials and query parameters out of access logs.
		log.Printf("http_request method=%s path=%q status=%d duration_ms=%d",
			r.Method, r.URL.Path, status, time.Since(started).Milliseconds())
	})
}
