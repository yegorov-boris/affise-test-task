package middleware

import (
	"net/http"
	"yegorov-boris/affise-test-task/internal/contracts"
)

func NewRateLimiter(limiter contracts.Tryer, inner contracts.Handler) contracts.Handler {
	return func(w http.ResponseWriter, r *http.Request) {
		if !limiter.Try() {
			http.Error(w, "Sorry, your request can not be currently served. Please, try again a bit later.", http.StatusTooManyRequests)
			return
		}

		inner(w, r)
	}
}
