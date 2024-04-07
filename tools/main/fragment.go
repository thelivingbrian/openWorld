package main

import (
	"fmt"
	"net/http"
)

type Fragment struct {
	Name  string  `json:"name"`
	Tiles [][]int `json:"tiles"`
}
type FragmentDetails struct {
	Name        string
	GridDetails GridDetails
}

// placeholder tile?

func (c Context) fragmentsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		c.getFragments(w, r)
	}
}

func (c Context) getFragments(w http.ResponseWriter, r *http.Request) {
	queryValues := r.URL.Query()
	collectionName := queryValues.Get("currentCollection")
	set := queryValues.Get("fragment-set")
	fragment := queryValues.Get("fragment")
	fmt.Printf("%s %s %s\n", collectionName, set, fragment)
	var setOptions []string
	collection := c.Collections[collectionName]
	fmt.Println(len(collection.Fragments))

	for key, _ := range collection.Fragments {
		fmt.Println(key)
		setOptions = append(setOptions, key)
	}

	fmt.Println(setOptions)
	fmt.Println("Set: " + set)
	fmt.Println(len(collection.Fragments[set]))

	var PageData = struct {
		FragmentSets    []string
		CurrentSet      string
		Fragments       []Fragment
		FragmentDetails *FragmentDetails
	}{
		FragmentSets: setOptions,
		CurrentSet:   set,
		Fragments:    collection.Fragments[set],
	}
	tmpl.ExecuteTemplate(w, "fragments", PageData)
}

/*
func (c Context) fragmentSelectHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		c.getFragmentSelect(w, r)
	}
}

func (c Context) getFragmentSelect(w http.ResponseWriter, r *http.Request) {
	queryValues := r.URL.Query()
	collectionName := queryValues.Get("currentCollection")
	set := queryValues.Get("fragment-set")
	fragment := queryValues.Get("fragment")
	fmt.Printf("%s %s %s", collectionName, set, fragment)
	var setOptions []string
	collection := c.Collections[collectionName]
	for key, _ := range collection.Fragments {
		setOptions = append(setOptions, key)
	}
	var PageData = struct {
		FragmentSet     []string
		Fragments       []Fragment
		FragmentDetails *FragmentDetails
	}{
		FragmentSet: setOptions,
	}
	tmpl.ExecuteTemplate(w, "fragment-select", PageData)
}
*/

// Utilities
func (c Context) GridDetailsFromFragment(fragment Fragment) FragmentDetails {
	return FragmentDetails{
		Name: fragment.Name,
		GridDetails: GridDetails{
			MaterialGrid:     c.DereferenceIntMatrix(fragment.Tiles),
			DefaultTileColor: "",
			ScreenID:         "Fragment"},
	}
}
