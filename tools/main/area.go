package main

import (
	"fmt"
	"io"
	"net/http"
)

type Area struct {
	Name             string      `json:"name"`
	Safe             bool        `json:"safe"`
	Tiles            [][]int     `json:"tiles"`
	Transports       []Transport `json:"transports"`
	DefaultTileColor string      `json:"defaultTileColor"`
	North            string      `json:"north,omitempty"`
	South            string      `json:"south,omitempty"`
	East             string      `json:"east,omitempty"`
	West             string      `json:"west,omitempty"`
}

type GridDetails struct {
	MaterialGrid     [][]Material
	DefaultTileColor string
	Location         string
	ScreenID         string
	GridType         string
	Oob              bool
}

type PageData struct {
	GridDetails        GridDetails
	AvailableMaterials []Material
	Name               string
}

// //////////////////////////////////////////////////////////
// Globals (fix)

var haveSelection bool = false
var selectedX int
var selectedY int

// ///////////////////////////////////////////////////////////
// Areas

func (c Context) areasHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		c.getAreas(w, r)
	}
}

func (c Context) getAreas(w http.ResponseWriter, r *http.Request) {
	space := c.spaceFromGET(r)
	err := tmpl.ExecuteTemplate(w, "areas", *space)
	if err != nil {
		fmt.Println(err)
	}
}

// ///////////////////////////////////////////////////////////
// Area

func (c *Context) areaHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		c.getArea(w, r)
	}
	if r.Method == "POST" {
		c.postArea(w, r)
	}
}

func (c *Context) getArea(w http.ResponseWriter, r *http.Request) {
	space := c.spaceFromGET(r)
	selectedArea := c.areaFromGET(r)
	if selectedArea == nil {
		io.WriteString(w, "<h2>no Area</h2>")
		return
	}

	modifications := c.AreaToMaterialGrid(*selectedArea)

	fmt.Printf("Materials Available: %d", len(c.materials))

	var pageData = PageData{
		GridDetails: GridDetails{
			MaterialGrid:     modifications,
			DefaultTileColor: selectedArea.DefaultTileColor,
			Location:         locationStringFromArea(selectedArea, space.Name),
			GridType:         "area",
			ScreenID:         "screen",
		},
		AvailableMaterials: c.materials,
		Name:               selectedArea.Name,
	}
	err := tmpl.ExecuteTemplate(w, "area-edit", pageData)
	if err != nil {
		fmt.Println(err)
	}
}

func (c Context) AreaToMaterialGrid(area Area) [][]Material {
	return c.DereferenceIntMatrix(area.Tiles)
}

func (c Context) DereferenceIntMatrix(matrix [][]int) [][]Material {
	out := make([][]Material, len(matrix))
	for y := range matrix {
		out[y] = make([]Material, len(matrix[y]))
		for x := range matrix[y] {
			out[y][x] = c.materials[matrix[y][x]]
		}
	}
	return out
}

func (c Context) postArea(w http.ResponseWriter, r *http.Request) {
	properties, _ := requestToProperties(r)
	name := properties["areaName"]
	safe := (properties["safe"] == "on")
	new := (properties["new"] == "true")
	defaultTileColor := properties["defaultTileColor"]
	collectionName := properties["currentCollection"]
	spaceName := properties["currentSpace"]

	space := c.getSpace(collectionName, spaceName)

	// This needs changing
	// Can make name immutable or add oldname as property
	selectedArea := getAreaByName(space.Areas, name)
	if selectedArea == nil {
		area := Area{Name: name, Safe: safe, Tiles: nil, Transports: nil, DefaultTileColor: defaultTileColor}
		space.Areas = append(space.Areas, area)
	} else {
		if new {
			io.WriteString(w, `<h2>Invalid Name</h2>`)
			return
		}
	}

	outFile := c.collectionPath + collectionName + "/spaces/" + spaceName + ".json"
	err := writeJsonFile(outFile, space.Areas)
	if err != nil {
		panic(err)
	}

	io.WriteString(w, `<h2>Success</h2>`)
}

// ///////////////////////////////////////////////////////////
// Pages

func (c Context) areaDetailsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		c.getAreaDetails(w, r)
	}
}

func (c Context) getAreaDetails(w http.ResponseWriter, r *http.Request) {
	space := c.spaceFromGET(r)
	area := c.areaFromGET(r)
	checked := ""
	if area.Safe {
		checked = "checked"
	}
	var page = struct {
		Space   *Space
		Area    *Area
		Checked string
	}{Space: space, Area: area, Checked: checked}

	// Have default tile color change trigger redisplay screen
	err := tmpl.ExecuteTemplate(w, "area-details", page)
	if err != nil {
		fmt.Println(err)
	}
}

func (c Context) areaDisplayHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		c.getAreaDisplay(w, r)
	}
}

func (c Context) getAreaDisplay(w http.ResponseWriter, r *http.Request) {
	err := tmpl.ExecuteTemplate(w, "area-display", nil)
	if err != nil {
		fmt.Println(err)
	}
}

func (c Context) areaNeighborsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		c.getNeighbors(w, r)
	}
	if r.Method == "POST" {
		c.postNeighbors(w, r)
	}
}

func (c Context) getNeighbors(w http.ResponseWriter, r *http.Request) {
	queryValues := r.URL.Query()

	collectionName := queryValues.Get("currentCollection")
	spaceName := queryValues.Get("currentSpace")
	space := c.getSpace(collectionName, spaceName)

	name := queryValues.Get("area-name")
	selectedArea := getAreaByName(space.Areas, name)
	if selectedArea == nil {
		io.WriteString(w, "<h2>no Area</h2>")
		return
	}

	err := tmpl.ExecuteTemplate(w, "neighbors-edit", *selectedArea)
	if err != nil {
		fmt.Println(err)
	}
}

func (c Context) postNeighbors(w http.ResponseWriter, r *http.Request) {
	properties, _ := requestToProperties(r)
	name := properties["area-name"]
	north := properties["north_input"]
	south := properties["south_input"]
	east := properties["east_input"]
	west := properties["west_input"]

	collectionName := properties["currentCollection"]
	spaceName := properties["currentSpace"]
	space := c.getSpace(collectionName, spaceName)
	selectedArea := getAreaByName(space.Areas, name)
	if selectedArea == nil {
		io.WriteString(w, "<h2>no Area</h2>")
		return
	}

	selectedArea.North = north
	selectedArea.South = south
	selectedArea.East = east
	selectedArea.West = west

	note := `<div id="confirmation_neighbor_change"><p>saved</p></div>`

	io.WriteString(w, note)

	tmpl.ExecuteTemplate(w, "neighbors-edit", *selectedArea)
}
