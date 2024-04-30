package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

type Collection struct {
	Name       string
	Spaces     map[string]*Space
	Fragments  map[string][]Fragment
	Prototypes map[string][]Prototype
}

func (c Context) collectionsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		c.getCollections(w, r)
	}
	if r.Method == "POST" {
		c.postCollections(w, r)
	}
}

func (c Context) getCollections(w http.ResponseWriter, r *http.Request) {
	tmpl.ExecuteTemplate(w, "collections", c.Collections)
}

func (c Context) postCollections(w http.ResponseWriter, r *http.Request) {
	props, ok := requestToProperties(r)
	if !ok {
		fmt.Println("Invalid POST to /collections. Properties are invalid.")
		io.WriteString(w, `<h3> Properties are invalid. </h3>`)
		return
	}
	name, ok := props["newCollectionName"]
	if !ok {
		fmt.Println("Invalid POST to /collections. Collection not found.")
		io.WriteString(w, `<h3> Collection not found. </h3>`)
		return
	}
	fmt.Println(name)
	c.Collections[name] = &Collection{Name: name, Spaces: make(map[string]*Space), Fragments: make(map[string][]Fragment)}
	c.createCollectionDirectories(name)
	spacesTmpl.Execute(w, c.Collections[name])
}

func (c Context) createCollectionDirectories(name string) {
	dirs := []string{"prototypes", "fragments", "spaces"}

	for _, dir := range dirs {
		fullPath := filepath.Join(c.collectionPath, name, dir)
		// Create the directory with os.MkdirAll which also creates all necessary parent directories
		err := os.MkdirAll(fullPath, os.ModePerm) // os.ModePerm is 0777, allowing read, write, and execute permissions
		if err != nil {
			fmt.Println("Error creating directory:", err)
			return
		}
		fmt.Println("Created directory:", fullPath)
	}
}
