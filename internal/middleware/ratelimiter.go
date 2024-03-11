package middleware

import (
	"net/http"
	"yegorov-boris/affise-test-task/internal/contracts"
)

func NewRateLimiter(bucket chan struct{}, inner contracts.Handler) contracts.Handler {
	return func(w http.ResponseWriter, r *http.Request) {
		select {
		case bucket <- struct{}{}:
			inner(w, r)
			<-bucket
		default:
			http.Error(w, "Sorry, your request can not be currently served. Please, try again a bit later.", http.StatusTooManyRequests)
		}
	}
}
