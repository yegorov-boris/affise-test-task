package handlers

import (
	"net/url"
	"strconv"
	"strings"
)

func lastPathPart(basePath, path string) string {
	prefix, _ := url.JoinPath(basePath, "/")

	return strings.TrimPrefix(path, prefix)
}

func parseID(basePath, path string) (uint64, error) {
	return strconv.ParseUint(lastPathPart(basePath, path), 10, 64)
}
