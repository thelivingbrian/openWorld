package main

import (
	"fmt"
	"net/http"
	"strconv"
)

type Fragment struct {
	Name    string `json:"name"`
	SetName string
	Tiles   [][]int `json:"tiles"`
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
	setName := queryValues.Get("fragment-set")
	fragmentName := queryValues.Get("fragment")
	fmt.Printf("%s %s %s\n", collectionName, setName, fragmentName)

	collection, ok := c.Collections[collectionName]
	if !ok {
		fmt.Println("Collection Name Invalid")
		return
	}

	var setOptions []string
	for key, _ := range collection.Fragments {
		setOptions = append(setOptions, key)
	}

	var fragDetails []*FragmentDetails
	if fragmentName != "" {
		fragment := getFragmentByName(collection.Fragments[setName], fragmentName)
		if fragment != nil {
			fragDetails = append(fragDetails, c.DetailsFromFragment(fragment, false))
		}
	} else {
		for i, fragment := range collection.Fragments[setName] {
			details := c.DetailsFromFragment(&fragment, false)
			details.GridDetails.ScreenID += "_" + strconv.Itoa(i)
			fragDetails = append(fragDetails, details)
		}
	}

	var PageData = struct {
		FragmentSets    []string
		CurrentSet      string
		Fragments       []Fragment
		CurrentFragment string
		FragmentDetails []*FragmentDetails
	}{
		FragmentSets:    setOptions,
		CurrentSet:      setName,
		Fragments:       collection.Fragments[setName],
		CurrentFragment: fragmentName,
		FragmentDetails: fragDetails,
	}
	tmpl.ExecuteTemplate(w, "fragments", PageData)
}

// Utilities
func (c Context) DetailsFromFragment(fragment *Fragment, clickable bool) *FragmentDetails {
	gridtype := "unclickable"
	if clickable {
		gridtype = "fragment"
	}
	return &FragmentDetails{
		Name: fragment.Name,
		GridDetails: GridDetails{
			MaterialGrid:     c.DereferenceIntMatrix(fragment.Tiles),
			DefaultTileColor: "",
			ScreenID:         fragment.SetName + "_" + fragment.Name,
			GridType:         gridtype},
	}
}

// Fragment

func (c Context) fragmentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		c.getFragment(w, r)
	}
}

func (c Context) getFragment(w http.ResponseWriter, r *http.Request) {
	queryValues := r.URL.Query()
	collectionName := queryValues.Get("currentCollection")
	setName := queryValues.Get("fragment-set")
	fragmentName := queryValues.Get("fragment")

	collection, ok := c.Collections[collectionName]
	if !ok {
		fmt.Println("Collection Name Invalid")
		return
	}

	fmt.Println("Number of fragments:")
	fmt.Println(len(collection.Fragments[setName]))

	fragment := getFragmentByName(collection.Fragments[setName], fragmentName)
	if fragment != nil {
		var pageData = struct {
			AvailableMaterials      []Material
			SelectedFragmentDetails []*FragmentDetails
		}{
			AvailableMaterials:      c.materials,
			SelectedFragmentDetails: append(make([]*FragmentDetails, 0), c.DetailsFromFragment(fragment, true)),
		}
		err := tmpl.ExecuteTemplate(w, "fragment-edit", pageData)
		if err != nil {
			fmt.Println(err)
		}
	} else {
		fmt.Println("No fragment with name: " + fragmentName)
	}
}
