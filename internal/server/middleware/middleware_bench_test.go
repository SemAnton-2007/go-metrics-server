package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func BenchmarkGzipMiddleware(b *testing.B) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte("test response"))
		if err != nil {
			b.Errorf("Failed to write response in handler: %v", err)
		}
	})
	middleware := GzipMiddleware(handler)

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		middleware.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			b.Errorf("Unexpected status code: %d", w.Code)
		}
	}
}

func BenchmarkHashMiddleware(b *testing.B) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte("test response"))
		if err != nil {
			b.Errorf("Failed to write response in handler: %v", err)
		}
	})
	middleware := HashMiddleware("test-key")(handler)

	req := httptest.NewRequest("POST", "/", nil)
	req.Header.Set("Content-Type", "application/json")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		middleware.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			b.Errorf("Unexpected status code: %d", w.Code)
		}
	}
}
