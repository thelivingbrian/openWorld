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
	SetName     string `json:"setName"`
}

type Transformation struct {
	ClockwiseRotations int    `json:"clockwiseRotations,omitempty"`
	ColorPalette       string `json:"colorPalette,omitempty"`
}

type TileData struct {
	PrototypeId    string         `json:"prototypeId,omitempty"`
	Transformation Transformation `json:"transformation,omitempty"`
}

func (c Context) PrototypesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		c.getPrototypes(w, r)
	}
}

type PrototypeSelectPage struct {
	PrototypeSets []string
	CurrentSet    string
	Prototypes    []Prototype
}

func (c *Context) getPrototypes(w http.ResponseWriter, r *http.Request) {
	var PageData = c.prototypeSelectFromRequest(r)
	err := tmpl.ExecuteTemplate(w, "prototype-select", PageData)
	if err != nil {
		fmt.Println(err)
	}
}

func (c *Context) prototypeSelectFromRequest(r *http.Request) PrototypeSelectPage {
	queryValues := r.URL.Query()
	collectionName := queryValues.Get("currentCollection")
	setName := queryValues.Get("prototype-set")

	collection, ok := c.Collections[collectionName]
	if !ok {
		fmt.Println("Collection Name Invalid")
		return PrototypeSelectPage{}
	}

	var setOptions []string
	for key := range collection.PrototypeSets {
		setOptions = append(setOptions, key)
	}

	protos := collection.PrototypeSets[setName]
	transformedProtos := make([]Prototype, len(protos))
	for i := range protos {
		transformedProtos[i] = protos[i].peekTransform(Transformation{})
	}
	return PrototypeSelectPage{
		PrototypeSets: setOptions,
		CurrentSet:    setName,
		Prototypes:    transformedProtos,
	}
}
