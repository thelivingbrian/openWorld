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

func (c *Context) blueprintInstructionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "DELETE" {
		c.deleteInstruction(w, r)
	}
}

func (c *Context) deleteInstruction(w http.ResponseWriter, r *http.Request) {
	properties, _ := requestToProperties(r)
	instructionId := properties["instruction-id"]

	blueprint := c.blueprintFromProperties(properties)

	for i := range blueprint.Instructions {
		if blueprint.Instructions[i].ID == instructionId {
			blueprint.Instructions = append(blueprint.Instructions[:i], blueprint.Instructions[i+1:]...)
			break
		}
	}

	fmt.Printf("Removing %s \r\n", instructionId)
}

func (c *Context) blueprintFromProperties(properties map[string]string) *Blueprint {
	collectionName := properties["currentCollection"]
	spaceName := properties["currentSpace"]
	name := properties["area-name"]

	collection, ok := c.Collections[collectionName]
	if !ok {
		panic("invalid collection")
	}
	space, ok := collection.Spaces[spaceName]
	if !ok {
		panic("invalid space name")
	}
	area := getAreaByName(space.Areas, name)
	if area == nil {
		panic("Invalid area")
	}

	if area.Blueprint == nil {
		panic("No Blueprint for: " + area.Name)
	}
	return area.Blueprint
}

func (c *Context) instructionOrderHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "PUT" {
		c.putInstructionOrder(w, r)
	}
}

func (c *Context) putInstructionOrder(w http.ResponseWriter, r *http.Request) {
	properties, _ := requestToProperties(r)
	instructionId := properties["instruction-id"]

	blueprint := c.blueprintFromProperties(properties)

	for i := range blueprint.Instructions {
		if blueprint.Instructions[i].ID == instructionId {
			hold := blueprint.Instructions[i]
			blueprint.Instructions[i] = blueprint.Instructions[i+1%len(blueprint.Instructions)]
			blueprint.Instructions[i+1%len(blueprint.Instructions)] = hold
			break
		}
	}

}
func (c *Context) instructionRotationHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "PUT" {
		c.putInstructionRotation(w, r)
	}
}

func (c *Context) putInstructionRotation(w http.ResponseWriter, r *http.Request) {
	properties, _ := requestToProperties(r)
	instructionId := properties["instruction-id"]

	blueprint := c.blueprintFromProperties(properties)

	for i := range blueprint.Instructions {
		if blueprint.Instructions[i].ID == instructionId {
			blueprint.Instructions[i].ClockwiseRotations += 1
		}
	}
}
