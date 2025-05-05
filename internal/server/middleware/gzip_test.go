package middleware_test

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-metrics-server/internal/server/middleware"
)

func TestGzipMiddleware(t *testing.T) {
	t.Run("should compress JSON response when client accepts gzip", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"status":"ok"}`))
		})

		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Accept-Encoding", "gzip")

		rr := httptest.NewRecorder()
		middleware.GzipMiddleware(handler).ServeHTTP(rr, req)

		if rr.Header().Get("Content-Encoding") != "gzip" {
			t.Error("Response should be gzipped")
		}
	})

	t.Run("should decompress gzipped request", func(t *testing.T) {
		var buf bytes.Buffer
		gz := gzip.NewWriter(&buf)
		gz.Write([]byte("test data"))
		gz.Close()

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			data, _ := io.ReadAll(r.Body)
			if string(data) != "test data" {
				t.Error("Request body was not decompressed")
			}
		})

		req := httptest.NewRequest("POST", "/", &buf)
		req.Header.Set("Content-Encoding", "gzip")

		rr := httptest.NewRecorder()
		middleware.GzipMiddleware(handler).ServeHTTP(rr, req)
	})

	t.Run("should not compress when client doesn't accept gzip", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("raw data"))
		})

		req := httptest.NewRequest("GET", "/", nil) // No Accept-Encoding header

		rr := httptest.NewRecorder()
		middleware.GzipMiddleware(handler).ServeHTTP(rr, req)

		if rr.Header().Get("Content-Encoding") == "gzip" {
			t.Error("Response should not be gzipped")
		}
		if rr.Body.String() != "raw data" {
			t.Error("Response body was corrupted")
		}
	})
}
