package handlers

import (
	"log/slog"
	"net/url"
	"strconv"
	"strings"
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

func parseID(basePath, path string) (uint64, error) {
	prefix, _ := url.JoinPath(basePath, "/")
	idStr := strings.TrimPrefix(path, prefix)

	return strconv.ParseUint(idStr, 10, 64)
}
