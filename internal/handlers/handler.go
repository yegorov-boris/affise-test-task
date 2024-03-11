package handlers

import (
	"net/url"
	"strconv"
	"strings"
)

func parseID(basePath, path string) (uint64, error) {
	prefix, _ := url.JoinPath(basePath, "/")
	idStr := strings.TrimPrefix(path, prefix)

	return strconv.ParseUint(idStr, 10, 64)
}
