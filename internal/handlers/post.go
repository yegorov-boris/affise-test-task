package handlers

import (
	"net/http"
	"yegorov-boris/affise-test-task/internal/contracts"
)

func NewPost() contracts.HandlerWithErr {
	return func(w http.ResponseWriter, r *http.Request) error {
		return nil
	}

}
