package handlers

import (
	"context"
	"net/http"
	"time"
)

// PingHandler возвращает обработчик для проверки соединения с БД
func PingHandler(db interface {
	Ping(ctx context.Context) error
}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()

		if err := db.Ping(ctx); err != nil {
			http.Error(w, "Database connection failed", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}
}
