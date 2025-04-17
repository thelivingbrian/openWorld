package main

import (
	"net/http"
	"path/filepath"
	"strings"
)

func imageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		getImage(w, r)
	}
}

func getImage(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) != 3 {
		// ./images/{{file}}
		// 0/ 1 / 2
		http.Error(w, "Invalid image path", http.StatusBadRequest)
		return
	}
	serveImage(w, r, parts[2])
}

// DDOS risk?
func serveImage(w http.ResponseWriter, r *http.Request, fileName string) {
	dir := "./data/images/"
	fileName += ".png"
	filePath := filepath.Join(dir, fileName)
	http.ServeFile(w, r, filePath)
}
