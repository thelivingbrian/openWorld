package main

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
)

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

	// This is a problem for fragments
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
	if r.Method == "PUT" {
		c.putInstruction(w, r)
	}
	if r.Method == "DELETE" {
		c.deleteInstruction(w, r)
	}
}

func (c *Context) putInstruction(w http.ResponseWriter, r *http.Request) {
	properties, _ := requestToProperties(r)
	instructionId := properties["instruction-id"]
	inputY, err := strconv.Atoi(properties["instruction-y"])
	if err != nil {
		panic("invalid Y")
	}
	inputX, err := strconv.Atoi(properties["instruction-x"])
	if err != nil {
		panic("invalid X")
	}
	inputRot, err := strconv.Atoi(properties["instruction-rot"])
	if err != nil {
		panic("invalid Rotation")
	}

	area := c.areaFromProperties(properties)
	blueprint := area.Blueprint
	col := c.collectionFromProperties(properties)

	for i := range blueprint.Instructions {
		if blueprint.Instructions[i].ID == instructionId {
			// Reset
			currentRotations := blueprint.Instructions[i].ClockwiseRotations
			grid := col.getTileGridByAssetId(blueprint.Instructions[i].GridAssetId)
			if currentRotations%2 == 1 {
				clearTiles(blueprint.Instructions[i].Y, blueprint.Instructions[i].X, len(grid[0]), len(grid), blueprint.Tiles)
			} else {
				clearTiles(blueprint.Instructions[i].Y, blueprint.Instructions[i].X, len(grid), len(grid[0]), blueprint.Tiles)
			}
			// update
			blueprint.Instructions[i].Y = inputY
			blueprint.Instructions[i].X = inputX
			blueprint.Instructions[i].ClockwiseRotations = mod(inputRot, 4)
		}
	}

	// Fresh apply
	for i := range blueprint.Instructions {
		col.applyInstruction(blueprint.Tiles, blueprint.Instructions[i])
	}

	// Instead of just bvlueprint can display whole area edit page
	err = tmpl.ExecuteTemplate(w, "area-blueprint", area)
	if err != nil {
		fmt.Println(err)
	}
}

func (c *Context) deleteInstruction(w http.ResponseWriter, r *http.Request) {
	properties, _ := requestToProperties(r)
	instructionId := properties["instruction-id"]

	area := c.areaFromProperties(properties)
	blueprint := area.Blueprint

	for i := range blueprint.Instructions {
		if blueprint.Instructions[i].ID == instructionId {
			blueprint.Instructions = append(blueprint.Instructions[:i], blueprint.Instructions[i+1:]...)
			break
		}
	}

	fmt.Printf("Removing %s \r\n", instructionId)
	err := tmpl.ExecuteTemplate(w, "area-blueprint", area)
	if err != nil {
		fmt.Println(err)
	}
}

/*
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
*/

func (c *Context) areaFromProperties(properties map[string]string) *AreaDescription {
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

	return area
}

func (c *Context) instructionOrderHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "PUT" {
		c.putInstructionOrder(w, r)
	}
}

func (c *Context) putInstructionOrder(w http.ResponseWriter, r *http.Request) {
	properties, _ := requestToProperties(r)
	instructionId := properties["instruction-id"]

	area := c.areaFromProperties(properties)
	blueprint := area.Blueprint

	// should clear everything and reapply new order
	for i := range blueprint.Instructions {
		if blueprint.Instructions[i].ID == instructionId {
			hold := blueprint.Instructions[i]
			blueprint.Instructions[i] = blueprint.Instructions[i+1%len(blueprint.Instructions)]
			blueprint.Instructions[i+1%len(blueprint.Instructions)] = hold
			break
		}
	}

	err := tmpl.ExecuteTemplate(w, "area-blueprint", area)
	if err != nil {
		fmt.Println(err)
	}
}

/*
func (c *Context) instructionRotationHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "PUT" {
		c.putInstructionRotation(w, r)
	}
}

func (c *Context) putInstructionRotation(_ http.ResponseWriter, r *http.Request) {
	properties, _ := requestToProperties(r)
	instructionId := properties["instruction-id"]

	blueprint := c.blueprintFromProperties(properties)
	col := c.collectionFromProperties(properties)

	for i := range blueprint.Instructions {
		if blueprint.Instructions[i].ID == instructionId {
			currentRotations := blueprint.Instructions[i].ClockwiseRotations
			grid := col.getTileGridByAssetId(blueprint.Instructions[i].GridAssetId)
			if currentRotations%2 == 1 {
				clearTiles(blueprint.Instructions[i].Y, blueprint.Instructions[i].X, len(grid[0]), len(grid), blueprint.Tiles)
			} else {
				clearTiles(blueprint.Instructions[i].Y, blueprint.Instructions[i].X, len(grid), len(grid[0]), blueprint.Tiles)
			}
			blueprint.Instructions[i].ClockwiseRotations = mod(currentRotations+1, 4)
		}
		col.applyInstruction(blueprint.Tiles, blueprint.Instructions[i])
	}
}
*/

func rotateTimesN(input [][]TileData, n int) [][]TileData {
	rotations := mod(n, 4)
	out := input
	for i := 0; i < rotations; i++ {
		out = rotateClockwise(out)
		for y := range out {
			for x := range out[y] {
				out[y][x].Transformation.ClockwiseRotations++
			}
		}
	}
	return out
}

func rotateClockwise[T any](input [][]T) [][]T {
	outheight := len(input[0])
	out := make([][]T, outheight)
	for i := 0; i < outheight; i++ {
		out[i] = make([]T, len(input))
		for j := 0; j < len(input); j++ {
			out[i][j] = input[len(input)-j-1][i]
		}
	}
	return out
}
