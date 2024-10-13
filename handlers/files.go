package handlers

import (
	"fmt"
	"net/http"
	"path"
	"strings"
)

type Files struct {
	basePath string
	handler  http.Handler
}

func NewFiles(dirPath string, basePath string) Files {
	return Files{handler: http.FileServer(http.Dir(dirPath)), basePath: basePath}
}

func (f Files) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cleanedPath, err := removeBaseURLPath(f.basePath, r.URL.Path)
	if err == nil {
		r.URL.Path = cleanedPath
	}

	f.handler.ServeHTTP(w, r)
}

func removeBaseURLPath(basePath, fullPath string) (string, error) {
	// Ensure paths start with "/"
	basePath = "/" + strings.Trim(basePath, "/")
	fullPath = "/" + strings.Trim(fullPath, "/")

	// Normalize paths
	basePath = path.Clean(basePath)
	fullPath = path.Clean(fullPath)

	// Check if the full path starts with the base path
	if !strings.HasPrefix(fullPath, basePath) {
		return "", fmt.Errorf("full path %s is not in base path %s", fullPath, basePath)
	}

	// Remove the base path
	relPath := strings.TrimPrefix(fullPath, basePath)

	// Ensure the relative path starts with "/"
	relPath = "/" + strings.TrimPrefix(relPath, "/")

	return relPath, nil
}
