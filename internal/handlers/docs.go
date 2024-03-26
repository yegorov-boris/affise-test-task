package handlers

import (
	"net/http"
	"path/filepath"
	"yegorov-boris/affise-test-task/internal/contracts"
)

func NewDocs(basePath string) contracts.Handler {
	return func(w http.ResponseWriter, r *http.Request) {
		name := lastPathPart(basePath, r.URL.Path)
		http.ServeFile(w, r, filepath.Join("./docs", name))
	}
}
