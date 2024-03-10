package store

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"yegorov-boris/affise-test-task/internal/models"
)

type Store struct {
	logger *slog.Logger
	path   string
}

func New(logger *slog.Logger, path string) *Store {
	return &Store{
		logger: logger,
		path:   path,
	}
}

func (s *Store) Save(id uint64, results []models.Output) {
	if len(results) == 0 {
		return
	}

	b, err := json.Marshal(results)
	if err != nil {
		s.logger.Error(fmt.Sprintf("failed to JSON encode results: %s", err))
		return
	}

	name := filepath.Join(s.path, fmt.Sprintf("%d.json", id))
	if err := os.WriteFile(name, b, 0644); err != nil {
		s.logger.Error(fmt.Sprintf("failed to write results to disk: %s", err))
	}
}
