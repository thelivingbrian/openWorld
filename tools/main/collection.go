package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

type Collection struct {
	Name              string
	Spaces            map[string]*Space
	Fragments         map[string][]Fragment
	PrototypeSets     map[string][]Prototype
	ProceeduralProtos map[string][]Prototype
	InteractableSets  map[string][]InteractableDescription
	StructureSets     map[string][]Structure
}

func (c *Context) collectionsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		c.getCollections(w, r)
	}
	if r.Method == "POST" {
		c.postCollections(w, r)
	}
}

func (c *Context) getCollections(w http.ResponseWriter, _ *http.Request) {
	tmpl.ExecuteTemplate(w, "collections", c.Collections)
}

func (c *Context) postCollections(w http.ResponseWriter, r *http.Request) {
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
	createCollectionDirectories(name, c.collectionPath)
	tmpl.ExecuteTemplate(w, "space-page", c.Collections[name])
}

func createCollectionDirectories(name string, path string) {
	dirs := []string{"prototypes", "fragments", "spaces", "interactables", "proc/prototypes", "proc/structures"}

	for _, dir := range dirs {
		fullPath := filepath.Join(path, name, dir)
		err := os.MkdirAll(fullPath, os.ModePerm)
		if err != nil {
			fmt.Println("Error creating directory:", err)
			return
		}
		fmt.Println("Created directory:", fullPath)
	}
}

func (col *Collection) getProtoSelect() PrototypeSelectPage {
	return PrototypeSelectPage{
		PrototypeSets: col.getProtoSets(),
		CurrentSet:    "",
		Prototypes:    nil,
	}
}

func (col *Collection) getProtoSets() []string {
	var setOptions []string
	for key := range col.PrototypeSets {
		setOptions = append(setOptions, key)
	}
	return setOptions
}

func (col *Collection) findPrototypeById(id string) *Prototype {
	for _, set := range col.PrototypeSets {
		for i := range set {
			if set[i].ID == id {
				return &set[i]
			}
		}
	}
	for _, set := range col.ProceeduralProtos {
		for i := range set {
			if set[i].ID == id {
				return &set[i]
			}
		}
	}
	fmt.Println("Invalid Prototype lookup: " + id)
	return nil
}

func (col *Collection) findInteractableById(id string) *InteractableDescription {
	for _, set := range col.InteractableSets {
		for i := range set {
			if set[i].ID == id {
				return &set[i]
			}
		}
	}
	return nil
}

func (c *Context) collectionFromProperties(properties map[string]string) *Collection {
	collectionName := properties["currentCollection"]

	collection, ok := c.Collections[collectionName]
	if !ok {
		panic("invalid collection")
	}
	return collection
}
