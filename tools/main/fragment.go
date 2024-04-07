package main

import "net/http"

type Fragment struct {
	width  int
	height int
	tiles  [][]int
}

// placeholder tile?

func (c Context) fragmentsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		c.getFragments(w, r)
	}
}

func (c Context) getFragments(w http.ResponseWriter, r *http.Request) {
	tmpl.ExecuteTemplate(w, "fragments", nil)
}
