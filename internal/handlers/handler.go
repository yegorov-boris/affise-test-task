package handlers

import (
	"log/slog"
	"yegorov-boris/affise-test-task/internal/contracts"
	"yegorov-boris/affise-test-task/internal/middleware"
)

func New(
	rateLimiter contracts.Tryer,
	logger *slog.Logger,
	inner contracts.HandlerWithErr,
) contracts.Handler {
	return middleware.NewRateLimiter(rateLimiter, middleware.NewLogger(logger, inner))
}
