package handlers

import (
	"context"
	"net/http"
	"time"
)

type PingHandler struct {
	db interface {
		Ping(ctx context.Context) error
	}
}

func NewPingHandler(db interface {
	Ping(ctx context.Context) error
}) *PingHandler {
	return &PingHandler{db: db}
}

func (h *PingHandler) Ping(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	if err := h.db.Ping(ctx); err != nil {
		http.Error(w, "Database connection failed", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
