package handlers

import (
	"context"
	"net/http"
	"time"
)

// PingHandler handles health check requests to verify database connectivity.
type PingHandler struct {
	db interface {
		Ping(ctx context.Context) error
	}
}

// NewPingHandler creates a new PingHandler instance with the given database connection.
func NewPingHandler(db interface {
	Ping(ctx context.Context) error
}) *PingHandler {
	return &PingHandler{db: db}
}

// Ping checks database connectivity and returns appropriate HTTP status.
// Returns 200 OK if database is reachable, 500 otherwise.
func (h *PingHandler) Ping(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	if err := h.db.Ping(ctx); err != nil {
		http.Error(w, "Database connection failed", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("OK")); err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
		return
	}
}
