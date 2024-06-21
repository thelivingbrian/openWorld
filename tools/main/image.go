package main

import (
	"fmt"
	"net/http"
	"path/filepath"
)

func (c Context) imagesHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("sending an image")

	// This only takes the very last portion of the path
	fileName := filepath.Base(r.URL.Path)
	fmt.Println(fileName)

	dir := "./data/images"
	filePath := filepath.Join(dir, fileName)
	http.ServeFile(w, r, filePath)
}
