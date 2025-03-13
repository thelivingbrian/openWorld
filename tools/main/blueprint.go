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
	Ground       [][]Cell
}

type Ground struct {
	Pattern           [][]Cell
	DefaultTileColor  string
	DefaultTileColor1 string
}

type Instruction struct {
	ID                 string
	X                  int
	Y                  int
	GridAssetId        string
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

	fragmentSet := queryValues.Get("fragment-set")
	fragmentName := queryValues.Get("fragment")
	if fragmentName != "" && fragmentSet != "" {
		set, ok := collection.Fragments[fragmentSet]
		if !ok {
			io.WriteString(w, "<h2>no Fragment set</h2>")
			return
		}
		fragment := getFragmentByName(set, fragmentName)
		err := tmpl.ExecuteTemplate(w, "fragment-blueprint", fragment)
		if err != nil {
			fmt.Println(err)
		}
		return
	}

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

	area, fragment := c.areaOrFragmentFromProperties(properties)
	blueprint := c.blueprintFromAreaOrFragment(area, fragment)

	col := c.collectionFromProperties(properties)

	for i := range blueprint.Instructions {
		if blueprint.Instructions[i].ID == instructionId {
			// Reset
			col.UndoInstruction(blueprint, i)
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

	executeBlueprintTemplate(w, fragment, area)
}

func (col *Collection) UndoInstruction(blueprint *Blueprint, i int) {
	currentRotations := blueprint.Instructions[i].ClockwiseRotations
	grid := col.getTileGridByAssetId(blueprint.Instructions[i].GridAssetId)
	if currentRotations%2 == 1 {
		clearTiles(blueprint.Instructions[i].Y, blueprint.Instructions[i].X, len(grid[0]), len(grid), blueprint.Tiles)
	} else {
		clearTiles(blueprint.Instructions[i].Y, blueprint.Instructions[i].X, len(grid), len(grid[0]), blueprint.Tiles)
	}
}

func executeBlueprintTemplate(w http.ResponseWriter, fragment *Fragment, area *AreaDescription) {
	if fragment == nil {
		// Instead of just blueprint can display whole area edit page
		err := tmpl.ExecuteTemplate(w, "area-blueprint", area)
		if err != nil {
			fmt.Println(err)
		}
	} else {
		err := tmpl.ExecuteTemplate(w, "fragment-blueprint", fragment)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func (c *Context) deleteInstruction(w http.ResponseWriter, r *http.Request) {
	properties, _ := requestToProperties(r)
	instructionId := properties["instruction-id"]

	area, fragment := c.areaOrFragmentFromProperties(properties)
	blueprint := c.blueprintFromAreaOrFragment(area, fragment)
	col := c.collectionFromProperties(properties)

	for i := range blueprint.Instructions {
		if blueprint.Instructions[i].ID == instructionId {
			col.UndoInstruction(blueprint, i)
			blueprint.Instructions = append(blueprint.Instructions[:i], blueprint.Instructions[i+1:]...)
			break
		}
	}

	fmt.Printf("Removing %s \r\n", instructionId)
	executeBlueprintTemplate(w, fragment, area)
}

func (c *Context) blueprintInstructionHighlightHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		c.postInstructionHighlight(w, r)
	}
	if r.Method == "DELETE" {
		c.deleteInstruction(w, r)
	}
}

func (c *Context) postInstructionHighlight(w http.ResponseWriter, r *http.Request) {
	properties, _ := requestToProperties(r)
	instructionId := properties["instruction-id"]

	col := c.collectionFromProperties(properties)
	if col == nil {
		panic("invalid collection")
	}

	area, fragment := c.areaOrFragmentFromProperties(properties)
	blueprint := c.blueprintFromAreaOrFragment(area, fragment)
	gridType, screenId := "fragment", "fragment"
	defaultTileColor := ""
	if fragment == nil {
		gridType, screenId = "area", "screen"
		defaultTileColor = area.DefaultTileColor
	}

	for i := range blueprint.Instructions {
		if blueprint.Instructions[i].ID == instructionId {
			details := GridClickDetails{GridType: gridType, ScreenID: screenId, X: blueprint.Instructions[i].X, Y: blueprint.Instructions[i].Y, DefaultTileColor: defaultTileColor}
			io.WriteString(w, col.gridSelect(details, blueprint.Tiles))
		}
	}

}

func (c *Context) blueprintFromAreaOrFragment(area *AreaDescription, fragment *Fragment) *Blueprint {
	var blueprint *Blueprint
	if area != nil {
		blueprint = area.Blueprint
	} else if fragment != nil {
		blueprint = fragment.Blueprint
	} else {
		panic("Failed to retrieve blueprint from area or fragment.")
	}
	return blueprint
}

func (c *Context) areaOrFragmentFromProperties(properties map[string]string) (*AreaDescription, *Fragment) {
	area := c.areaFromProperties(properties)
	if area != nil {
		return area, nil
	}
	fragment := c.fragmentFromProperties(properties)
	if fragment != nil {
		return nil, fragment
	}
	return nil, nil
}

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
	return area
}

func (c *Context) fragmentFromProperties(properties map[string]string) *Fragment {
	collectionName := properties["currentCollection"]
	fragmentSet := properties["fragment-set"]
	name := properties["fragment"]
	if fragmentSet == "" || name == "" {
		return nil
	}

	collection, ok := c.Collections[collectionName]
	if !ok {
		panic("invalid collection")
	}
	fragments, ok := collection.Fragments[fragmentSet]
	if !ok {
		panic("invalid space name")
	}
	fragment := getFragmentByName(fragments, name)
	if fragment == nil {
		panic("Invalid area")
	}

	return fragment
}

func (c *Context) instructionOrderHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "PUT" {
		c.putInstructionOrder(w, r)
	}
}

func (c *Context) putInstructionOrder(w http.ResponseWriter, r *http.Request) {
	properties, _ := requestToProperties(r)
	instructionId := properties["instruction-id"]

	area, fragment := c.areaOrFragmentFromProperties(properties)
	blueprint := c.blueprintFromAreaOrFragment(area, fragment)

	// should clear everything and reapply new order
	for i := range blueprint.Instructions {
		if blueprint.Instructions[i].ID == instructionId {
			hold := blueprint.Instructions[i]
			blueprint.Instructions[i] = blueprint.Instructions[i+1%len(blueprint.Instructions)]
			blueprint.Instructions[i+1%len(blueprint.Instructions)] = hold
			break
		}
	}

	executeBlueprintTemplate(w, fragment, area)
}

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

// ///////////////////////////////////////////////////////////
// Ground

func (c *Context) blueprintGroundHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		c.getGroundEdit(w, r)
	}
}

func (c *Context) getGroundEdit(w http.ResponseWriter, r *http.Request) {
	queryValues := r.URL.Query()
	space := c.spaceFromGET(r)
	name := queryValues.Get("area-name")
	selectedArea := getAreaByName(space.Areas, name)
	if selectedArea == nil {
		io.WriteString(w, "<h2>no Area</h2>")
		return
	}

	collection := c.collectionFromGet(r)
	modifications := collection.generateMaterials(selectedArea.Blueprint.Tiles)

	var pageData = AreaEditPageData{
		AreaWithGrid: AreaWithGrid{
			GridDetails: GridDetails{
				MaterialGrid:     modifications,
				InteractableGrid: collection.generateInteractables(selectedArea.Blueprint.Tiles),
				DefaultTileColor: selectedArea.DefaultTileColor,
				Location:         locationStringFromArea(selectedArea, space.Name),
				GridType:         "ground",
				ScreenID:         "screen",
			},
			SelectedArea:   *selectedArea,
			NavHasHadClick: false,
		},
	}
	err := tmpl.ExecuteTemplate(w, "ground-edit", pageData)
	if err != nil {
		fmt.Println(err)
	}
}
