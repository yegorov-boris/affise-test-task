package handlers

import (
	"fmt"
	"net/http"
	"yegorov-boris/affise-test-task/internal/contracts"
)

func NewDelete(basePath, storePath string, state contracts.State) contracts.HandlerWithErr {
	return func(w http.ResponseWriter, r *http.Request) error {
		id, err := parseID(basePath, r.URL.Path)
		if err != nil {
			http.Error(w, "Invalid ID", http.StatusBadRequest)

			return fmt.Errorf("invalid ID: %w", err)
		}

		if state.Cancel(id) {
			w.WriteHeader(http.StatusNoContent)

			return nil
		}

		http.Error(w, "Request in progress not found by ID", http.StatusNotFound)

		return nil
	}
}
