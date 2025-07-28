package middleware

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHashMiddleware(t *testing.T) {
	t.Run("no key skips validation", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest("POST", "/", bytes.NewBufferString("test"))
		rr := httptest.NewRecorder()

		middleware := HashMiddleware("")
		middleware(handler).ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("valid hash passes", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		body := []byte("test")
		h := hmac.New(sha256.New, []byte("testkey"))
		h.Write(body)
		hash := hex.EncodeToString(h.Sum(nil))

		req := httptest.NewRequest("POST", "/", bytes.NewBuffer(body))
		req.Header.Set("HashSHA256", hash)
		rr := httptest.NewRecorder()

		middleware := HashMiddleware("testkey")
		middleware(handler).ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("invalid hash fails", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest("POST", "/", bytes.NewBufferString("test"))
		req.Header.Set("HashSHA256", "invalidhash")
		rr := httptest.NewRecorder()

		middleware := HashMiddleware("testkey")
		middleware(handler).ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("GET request skips validation", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest("GET", "/", nil)
		rr := httptest.NewRecorder()

		middleware := HashMiddleware("testkey")
		middleware(handler).ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
	})
}
