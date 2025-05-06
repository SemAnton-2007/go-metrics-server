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

// HashMiddleware создает middleware для проверки подписи запросов и подписи ответов
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
					defer gz.Close()
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
				h.Write(body)
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
		h.Write(b)
		hash := hex.EncodeToString(h.Sum(nil))
		w.Header().Set("HashSHA256", hash)
	}
	return w.ResponseWriter.Write(b)
}
