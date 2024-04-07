package main

import (
	"net/http"
)

type Collection struct {
	Name      string
	Spaces    map[string]*Space
	Fragments map[string][]Fragment
}

func (c Context) collectionsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		c.getCollections(w, r)
	}
}

func (c Context) getCollections(w http.ResponseWriter, r *http.Request) {
	tmpl.ExecuteTemplate(w, "collections", c.Collections)
}
