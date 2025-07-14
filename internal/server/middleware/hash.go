package middleware

import (
	"bytes"
	"compress/gzip"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"strings"
)

// HashMiddleware creates a middleware that verifies HMAC-SHA256 signatures.
// It checks the HashSHA256 header for incoming requests and adds it to responses.
// If key is empty, signature verification is skipped.
func HashMiddleware(key string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Если ключ не установлен, пропускаем проверку
			if key == "" {
				next.ServeHTTP(w, r)
				return
			}

			// Для POST запросов проверяем подпись
			if r.Method == http.MethodPost {
				hash := r.Header.Get("HashSHA256")
				if hash == "" {
					http.Error(w, "Missing HashSHA256 header", http.StatusBadRequest)
					return
				}

				// Читаем тело запроса
				var body []byte
				var err error

				// Если запрос сжат, распаковываем для проверки подписи
				if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
					gz, err := gzip.NewReader(r.Body)
					if err != nil {
						http.Error(w, "Failed to decompress request body", http.StatusBadRequest)
						return
					}
					defer func() {
						if err := gz.Close(); err != nil {
							http.Error(w, "Failed to close gzip reader", http.StatusInternalServerError)
							return
						}
					}()
					body, err = io.ReadAll(gz)
					if err != nil {
						http.Error(w, "Failed to read decompressed body", http.StatusBadRequest)
						return
					}
					// Заменяем тело для последующих обработчиков
					r.Body = io.NopCloser(bytes.NewReader(body))
				} else {
					body, err = io.ReadAll(r.Body)
					if err != nil {
						http.Error(w, "Failed to read request body", http.StatusBadRequest)
						return
					}
					r.Body = io.NopCloser(bytes.NewReader(body))
				}

				// Проверяем подпись
				h := hmac.New(sha256.New, []byte(key))
				if _, err := h.Write(body); err != nil {
					http.Error(w, "Failed to calculate hash", http.StatusInternalServerError)
					return
				}
				expectedHash := hex.EncodeToString(h.Sum(nil))

				if hash != expectedHash {
					http.Error(w, "Invalid HashSHA256", http.StatusBadRequest)
					return
				}
			}

			// Обертываем ResponseWriter для добавления подписи к ответу
			writer := &hashResponseWriter{
				ResponseWriter: w,
				key:            []byte(key),
			}
			next.ServeHTTP(writer, r)
		})
	}
}

// hashResponseWriter обертка для http.ResponseWriter, добавляющая подпись к ответу
type hashResponseWriter struct {
	http.ResponseWriter
	key []byte
}

func (w *hashResponseWriter) Write(b []byte) (int, error) {
	// Добавляем подпись, если ключ установлен
	if len(w.key) > 0 {
		h := hmac.New(sha256.New, w.key)
		if _, err := h.Write(b); err != nil {
			return 0, err
		}
		hash := hex.EncodeToString(h.Sum(nil))
		w.Header().Set("HashSHA256", hash)
	}
	n, err := w.ResponseWriter.Write(b)
	if err != nil {
		return 0, err
	}
	return n, nil
}
