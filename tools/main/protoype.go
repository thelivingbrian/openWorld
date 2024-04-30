package main

import (
	"fmt"
	"net/http"
)

type Prototype struct {
	ID          string `json:"id"`
	CommonName  string `json:"commonName"`
	CssColor    string `json:"cssColor"`
	Walkable    bool   `json:"walkable"` // More complex than bool? Or seperate OnStep Property? Transports?
	Floor1Css   string `json:"layer1css"`
	Floor2Css   string `json:"layer2css"`
	Ceiling1Css string `json:"ceiling1css"`
	Ceiling2Css string `json:"ceiling2css"`
	// Tranformation pointer
}

type Transformation struct {
	ClockwiseRotations int
	ColorPalette       string
}

func (c Context) PrototypesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		c.getPrototypes(w, r)
	}
}

func (c Context) getPrototypes(w http.ResponseWriter, r *http.Request) {
	queryValues := r.URL.Query()
	collectionName := queryValues.Get("currentCollection")
	setName := queryValues.Get("prototype-set")
	fmt.Printf("%s %s \n", collectionName, setName)

	collection, ok := c.Collections[collectionName]
	if !ok {
		fmt.Println("Collection Name Invalid")
		return
	}

	var setOptions []string
	for key := range collection.Prototypes {
		setOptions = append(setOptions, key)
	}
	fmt.Println(setOptions)

	var PageData = struct {
		PrototypeSets []string
		CurrentSet    string
		Prototypes    []Prototype
	}{
		PrototypeSets: setOptions,
		CurrentSet:    setName,
		Prototypes:    collection.Prototypes[setName],
	}
	err := tmpl.ExecuteTemplate(w, "prototype-select", PageData)
	if err != nil {
		fmt.Println(err)
	}
}
