package middleware

import (
	"context"
	"net"
	"net/http"

	"github.com/rs/zerolog/log"
)

type contextKey string

const (
	ipContextKey contextKey = "ip"
)

func TrustedSubnetMiddleware(trustedSubnet string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if trustedSubnet == "" {
				next.ServeHTTP(w, r)
				return
			}

			ipStr := r.Header.Get("X-Real-IP")
			if ipStr == "" {
				log.Warn().Msg("X-Real-IP header missing")
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			ip := net.ParseIP(ipStr)
			if ip == nil {
				log.Warn().Str("ip", ipStr).Msg("Invalid IP format")
				http.Error(w, "Invalid IP address", http.StatusForbidden)
				return
			}

			_, cidrNet, err := net.ParseCIDR(trustedSubnet)
			if err != nil {
				log.Error().Err(err).Str("cidr", trustedSubnet).Msg("CIDR parse error")
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}

			if !cidrNet.Contains(ip) {
				log.Warn().Str("ip", ipStr).Str("cidr", trustedSubnet).Msg("IP not in trusted subnet")
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			ctx := context.WithValue(r.Context(), ipContextKey, ipStr)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
