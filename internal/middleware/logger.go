package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"yegorov-boris/affise-test-task/internal/contracts"
)

func NewLogger(logger *slog.Logger, inner contracts.HandlerWithErr) contracts.Handler {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := inner(w, r); err != nil {
			logger.Error(fmt.Sprintf("failed to handle %s %s: %s", r.Method, r.URL, err))
		}
	}
}
