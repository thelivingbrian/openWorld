package main

import (
	"fmt"
	"io"
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
	SetName     string
	GridDetails GridDetails
}

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
	for key := range collection.Fragments {
		setOptions = append(setOptions, key)
	}

	var fragmentDetails []*FragmentDetails
	if fragmentName != "" {
		fragment := getFragmentByName(collection.Fragments[setName], fragmentName)
		if fragment != nil {
			fragmentDetails = append(fragmentDetails, c.DetailsFromFragment(fragment, false))
		}
	} else {
		for i, fragment := range collection.Fragments[setName] {
			details := c.DetailsFromFragment(&fragment, false)
			details.GridDetails.ScreenID += strconv.Itoa(i)
			fragmentDetails = append(fragmentDetails, details)
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
		FragmentDetails: fragmentDetails,
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
		Name:    fragment.Name,
		SetName: fragment.SetName,
		GridDetails: GridDetails{
			MaterialGrid:     c.DereferenceIntMatrix(fragment.Tiles),
			DefaultTileColor: "",
			Location:         fragment.SetName + "." + fragment.Name,
			ScreenID:         "fragment",
			GridType:         gridtype},
	}
}

// Fragment

func (c *Context) fragmentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		c.getFragment(w, r)
	}
	if r.Method == "POST" {
		c.postFragment(w, r)
	}
}

func (c *Context) getFragment(w http.ResponseWriter, r *http.Request) {
	queryValues := r.URL.Query()
	collectionName := queryValues.Get("currentCollection")
	setName := queryValues.Get("fragment-set")
	fragmentName := queryValues.Get("fragment")

	collection, ok := c.Collections[collectionName]
	if !ok {
		panic("Collection Name Invalid")
	}
	if len(collection.Fragments[setName]) == 0 {
		panic("No Fragments in set: " + setName)
	}

	fragment := getFragmentByName(collection.Fragments[setName], fragmentName)
	if fragment == nil {
		panic("No fragment with name: " + fragmentName)
	}
	var pageData = struct {
		AvailableMaterials []Material
		FragmentDetails    *FragmentDetails
	}{
		AvailableMaterials: c.materials,
		FragmentDetails:    c.DetailsFromFragment(fragment, true),
	}
	err := tmpl.ExecuteTemplate(w, "fragment-edit", pageData)
	if err != nil {
		fmt.Println(err)
	}
}

func (c Context) postFragment(w http.ResponseWriter, r *http.Request) {
	fmt.Println("POST for /fragment")

	properties, _ := requestToProperties(r)
	collectionName := properties["currentCollection"]
	setName := properties["fragment-set"]

	collection, ok := c.Collections[collectionName]
	if !ok {
		panic("no collection")
	}
	set, ok := collection.Fragments[setName]
	if !ok {
		panic("no set")
	}

	outFile := c.collectionPath + collectionName + "/fragments/" + setName + ".json"
	err := writeJsonFile(outFile, set)
	if err != nil {
		panic(err)
	}

	io.WriteString(w, "<h3>Done.</h3>")
}
