package main

import (
	"fmt"
	"image/png"
	"net/http"
	"path/filepath"
	"strings"
)

func (c *Context) imageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		queryValues := r.URL.Query()
		name := queryValues.Get("currentCollection")
		//fmt.Println("Getting image from collection: " + name)
		if col, ok := c.Collections[name]; ok {
			c.getImage(w, r, col)
		}
	}
}

func (c *Context) getImage(w http.ResponseWriter, r *http.Request, col *Collection) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		panic("invalid image path")
	}
	if parts[2] == "map" {
		fmt.Println("Serving map")
		if len(parts) == 5 {
			col.serveMap(w, r, parts[3], parts[4])
		} else {
			col.serveMap(w, r, parts[3], parts[3])
		}

	} else if parts[2] == "make" {
		space, ok := col.Spaces[parts[3]]
		if !ok {
			panic("Invalid space for make image request")
		}
		area := getAreaByName(space.Areas, parts[4])
		if area == nil {
			panic("invalid area for make image request")
		}
		img := c.generateImgFromArea(area, *col)
		w.Header().Set("Content-Type", "image/png")

		if err := png.Encode(w, img); err != nil {
			http.Error(w, "Unable to encode image.", http.StatusInternalServerError)
			return
		}
	} else {
		panic("unsupported image request")
	}
}

func (col *Collection) serveMap(w http.ResponseWriter, r *http.Request, spaceName, fileName string) {
	dir := "./data/collections/" + col.Name + "/spaces/maps/" + spaceName // A function returning this string already exists but on a *Context
	fileName += ".png"
	filePath := filepath.Join(dir, fileName)
	http.ServeFile(w, r, filePath)
}
