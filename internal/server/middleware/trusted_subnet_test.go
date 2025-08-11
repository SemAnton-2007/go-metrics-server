package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"go-metrics-server/internal/server/middleware"
)

func TestTrustedSubnetMiddleware(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	tests := []struct {
		name          string
		header        string
		trustedSubnet string
		expectedCode  int
	}{
		{
			name:          "valid ip in subnet",
			header:        "192.168.1.10",
			trustedSubnet: "192.168.1.0/24",
			expectedCode:  http.StatusOK,
		},
		{
			name:          "ip not in subnet",
			header:        "10.0.0.1",
			trustedSubnet: "192.168.1.0/24",
			expectedCode:  http.StatusForbidden,
		},
		{
			name:          "missing header",
			trustedSubnet: "192.168.1.0/24",
			expectedCode:  http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := middleware.TrustedSubnetMiddleware(tt.trustedSubnet)
			req := httptest.NewRequest("GET", "/", nil)
			if tt.header != "" {
				req.Header.Set("X-Real-IP", tt.header)
			}

			rr := httptest.NewRecorder()
			middleware(handler).ServeHTTP(rr, req)

			if rr.Code != tt.expectedCode {
				t.Errorf("Expected status %d, got %d", tt.expectedCode, rr.Code)
			}
		})
	}
}
