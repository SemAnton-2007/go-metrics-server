package middleware

import (
	"bytes"
	"io"
	"net/http"

	"go-metrics-server/internal/crypto/hybrid"
	"go-metrics-server/internal/crypto/keys"
	"go-metrics-server/internal/server/config"

	"github.com/sirupsen/logrus"
)

func DecryptionMiddleware(cfg *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if cfg.CryptoKey == "" || r.Header.Get("Encryption") != "hybrid" {
				next.ServeHTTP(w, r)
				return
			}

			privateKey, err := keys.LoadPrivateKey(cfg.CryptoKey)
			if err != nil {
				logrus.Errorf("Failed to load private key: %v", err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}

			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "Failed to read request body", http.StatusBadRequest)
				return
			}
			defer r.Body.Close()

			decryptedBody, err := hybrid.Decrypt(body, privateKey)
			if err != nil {
				logrus.Errorf("Decryption failed: %v", err)
				http.Error(w, "Bad request", http.StatusBadRequest)
				return
			}

			r.Body = io.NopCloser(bytes.NewReader(decryptedBody))
			r.Header.Del("Encryption")

			next.ServeHTTP(w, r)
		})
	}
}
