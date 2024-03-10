package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"yegorov-boris/affise-test-task/internal/contracts"
)

func NewGet(basePath, storePath string, state contracts.State) contracts.HandlerWithErr {
	return func(w http.ResponseWriter, r *http.Request) error {
		id, err := parseID(basePath, r.URL.Path)
		if err != nil {
			http.Error(w, "Invalid ID", http.StatusBadRequest)

			return fmt.Errorf("invalid ID: %w", err)
		}

		if state.Check(id) {
			if _, err := fmt.Fprintf(w, "Your request is in progress. Please, try a bit later."); err != nil {
				return fmt.Errorf("failed to write response body: %w", err)
			}

			return nil
		}

		path := filepath.Join(storePath, fmt.Sprintf("%d.json", id))
		f, err := os.Open(path)
		if err != nil {
			http.Error(w, "Output not found by ID", http.StatusNotFound)

			return fmt.Errorf("failed to open file %q: %w", path, err)
		}

		if _, err := io.Copy(w, f); err != nil {
			_ = f.Close()
			return fmt.Errorf("failed to read file %q: %w", path, err)
		}

		if err := f.Close(); err != nil {
			return fmt.Errorf("failed to close file %q: %w", path, err)
		}

		return nil
	}
}
