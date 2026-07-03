package main

import (
	"net/http"
	"path/filepath"
	"strings"
)

func imageHandler(w http.ResponseWriter, r *http.Request) {
	createImageHandler("./data/images")(w, r)
}

func createImageHandler(directory string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			getImageFromDirectory(w, r, directory)
		}
	}
}

func getImage(w http.ResponseWriter, r *http.Request) {
	getImageFromDirectory(w, r, "./data/images")
}

func getImageFromDirectory(w http.ResponseWriter, r *http.Request, directory string) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) != 3 {
		// ./images/{{file}}
		// 0/ 1 / 2
		http.Error(w, "Invalid image path", http.StatusBadRequest)
		return
	}
	serveImageFromDirectory(w, r, directory, parts[2])
}

// DDOS risk?
func serveImage(w http.ResponseWriter, r *http.Request, fileName string) {
	serveImageFromDirectory(w, r, "./data/images", fileName)
}

func serveImageFromDirectory(w http.ResponseWriter, r *http.Request, dir, fileName string) {
	if filepath.Base(fileName) != fileName || strings.Contains(fileName, "..") {
		http.Error(w, "Invalid image path", http.StatusBadRequest)
		return
	}
	fileName += ".png"
	filePath := filepath.Join(dir, fileName)
	http.ServeFile(w, r, filePath)
}
