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
		panic("invalid image path")
	}
	serveImage(w, r, parts[2])
}

func serveImage(w http.ResponseWriter, r *http.Request, fileName string) {
	dir := "./data/images/"
	fileName += ".png"
	filePath := filepath.Join(dir, fileName)
	http.ServeFile(w, r, filePath)
}
