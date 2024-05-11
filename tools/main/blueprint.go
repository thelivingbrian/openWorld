package main

import (
	"fmt"
	"io"
	"net/http"
)

// On load make a ref map of grids and include bpth fragments and protos

// Consolidate Fragments, and make a combined lookup func for tile grids including protos

type Blueprint struct {
	Tiles        [][]TileData `json:"tiles"`
	Instructions []Instruction
}

type Instruction struct {
	ID          string
	X           int
	Y           int
	GridAssetId string
	// Transformation or something new?
	ClockwiseRotations int
}

func (c *Context) areaBlueprintHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		c.getBlueprint(w, r)
	}
}

func (c *Context) getBlueprint(w http.ResponseWriter, r *http.Request) {
	queryValues := r.URL.Query()

	collectionName := queryValues.Get("currentCollection")
	collection, ok := c.Collections[collectionName]
	if !ok {
		panic("invalid collection")
	}
	spaceName := queryValues.Get("currentSpace")
	space, ok := collection.Spaces[spaceName]
	if !ok {
		panic("invalid space name")
	}
	//space := c.getSpace(collectionName, spaceName)

	name := queryValues.Get("area-name")
	selectedArea := getAreaByName(space.Areas, name)
	if selectedArea == nil {
		io.WriteString(w, "<h2>no Area</h2>")
		return
	}

	err := tmpl.ExecuteTemplate(w, "area-blueprint", selectedArea)
	if err != nil {
		fmt.Println(err)
	}
}
