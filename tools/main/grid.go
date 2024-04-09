package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

type ClickEvent struct {
	Y                int
	X                int
	GridType         string
	DefaultTileColor string
	ScreenID         string
	Selected         bool
}

func (c Context) gridEditHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		c.getGridEdit(w, r)
	}
}

func (c Context) getGridEdit(w http.ResponseWriter, r *http.Request) {
	tmpl.ExecuteTemplate(w, "grid-modify", c.materials)
}

func (c Context) gridClickAreaHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		return
	}
	event, success := dataFromClickRequest(r, "area")
	if !success {
		panic("No Coordinates provided")
	}
	fmt.Printf("%d %d %s %s\n", event.Y, event.X, event.ScreenID, event.DefaultTileColor)

	parts := strings.Split(event.ScreenID, "_")
	if len(parts) < 2 {
		fmt.Println("Invalid Sid")
	}
	spaceName := parts[0]
	areaName := parts[1]

	properties, _ := requestToProperties(r)
	selectedTool, ok := properties["radio-tool"]
	if !ok {
		fmt.Println("No Tool Selected")
		return
	}
	collectionName, ok := properties["currentCollection"]
	if !ok {
		fmt.Println("No Collection")
		return
	}

	// Move into headers?
	//defaultTileColor := properties["defaultTileColor"]

	space := c.getSpace(collectionName, spaceName)
	if space != nil {
		area := getAreaByName(space.Areas, areaName)
		fmt.Println("Have: " + area.Name)
		if selectedTool == "select" {
			io.WriteString(w, gridSelect(event))
		} else if selectedTool == "replace" {
			selectedMaterial := c.getMaterialFromRequestProperties(properties)
			io.WriteString(w, gridReplace(event, area.Tiles, selectedMaterial))
		} else if selectedTool == "fill" {
			selectedMaterial := c.getMaterialFromRequestProperties(properties)
			gridFill(event, area.Tiles, selectedMaterial)
			var pageData = GridDetails{
				MaterialGrid:     c.DereferenceIntMatrix(area.Tiles),
				DefaultTileColor: area.DefaultTileColor,
				ScreenID:         event.ScreenID, // "screen?"
				GridType:         "area",
				Oob:              true,
			}

			err := tmpl.ExecuteTemplate(w, "grid", pageData)
			if err != nil {
				fmt.Println(err)
			}
		}

	}

	//fmt.Println("using: " + selectedTool + " : " + event.DefaultTileColor)
	//c.clickOnSquare(w, r)
	// Get [][]int from Area and pass to tool for mutation
}

func (c Context) gridClickFragmentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		return
	}
	// Get [][]int from Fragment and pass to tool for mutation
}

// / Tools
func gridSelect(event ClickEvent) string {
	//output := ""
	var buf bytes.Buffer
	if haveSelection {
		var pageData = struct {
			Material   Material
			ClickEvent ClickEvent
		}{
			Material: modifications[selectedY][selectedX],
			ClickEvent: ClickEvent{
				Y:                selectedY,
				X:                selectedX,
				GridType:         "area",
				DefaultTileColor: event.DefaultTileColor,
				ScreenID:         event.ScreenID},
		}
		err := tmpl.ExecuteTemplate(&buf, "grid-square", pageData)
		if err != nil {
			fmt.Println(err)
		}
	}
	haveSelection = true // Probably should be a hidden input
	selectedY = event.Y
	selectedX = event.X
	event.Selected = true
	var pageData = struct {
		Material   Material
		ClickEvent ClickEvent
	}{
		Material:   modifications[selectedY][selectedX],
		ClickEvent: event,
	}
	err := tmpl.ExecuteTemplate(&buf, "grid-square", pageData)
	if err != nil {
		fmt.Println(err)
	}
	return buf.String()
}

func gridReplace(event ClickEvent, modifications [][]int, selectedMaterial Material) string {
	modifications[event.Y][event.X] = selectedMaterial.ID
	var buf bytes.Buffer
	var pageData = struct {
		Material   Material
		ClickEvent ClickEvent
	}{
		Material:   selectedMaterial,
		ClickEvent: event,
	}
	err := tmpl.ExecuteTemplate(&buf, "grid-square", pageData)
	if err != nil {
		fmt.Println(err)
	}
	return buf.String()
}

func gridFill(event ClickEvent, modifications [][]int, selectedMaterial Material) {
	targetId := modifications[event.Y][event.X]
	seen := make([][]bool, len(modifications))
	for row := range seen {
		seen[row] = make([]bool, len(modifications[row]))
	}
	fill(event, modifications, selectedMaterial, seen, targetId)
	//return getHTMLFromModifications(event.DefaultTileColor)
}

func fill(event ClickEvent, modifications [][]int, selectedMaterial Material, seen [][]bool, targetId int) {
	seen[event.Y][event.X] = true
	modifications[event.Y][event.X] = selectedMaterial.ID
	deltas := []int{-1, 1}
	for _, i := range deltas {
		if event.Y+i >= 0 && event.Y+i < len(modifications) {
			shouldfill := !seen[event.Y+i][event.X] && modifications[event.Y+i][event.X] == targetId
			if shouldfill {
				event.Y += i
				fill(event, modifications, selectedMaterial, seen, targetId)
			}
		}
		if event.X+i >= 0 && event.X+i < len(modifications[event.Y]) {
			shouldfill := !seen[event.Y][event.X+i] && modifications[event.Y][event.X+i] == targetId
			if shouldfill {
				event.X += i
				fill(event, modifications, selectedMaterial, seen, targetId)
			}
		}
	}
}

///

func (c Context) selectMaterial(w http.ResponseWriter, r *http.Request) {
	queryValues := r.URL.Query()
	id := queryValues.Get("materialId")

	var selectedMaterial Material
	for _, material := range c.materials {
		if id, _ := strconv.Atoi(id); id == material.ID {
			selectedMaterial = material
		}
	}

	io.WriteString(w, exampleSquareFromMaterial(selectedMaterial))
}
func (c Context) selectFixture(w http.ResponseWriter, r *http.Request) {
	queryValues := r.URL.Query()
	fixtureType := queryValues.Get("current-fixture")

	if fixtureType == "material" {
		//fmt.Println("Fixture Material")
		tmpl.ExecuteTemplate(w, "fixture-material", c.materials)
	}
	if fixtureType == "fragment" {
		//fmt.Println("Fixture Fragments")
		collectionName := queryValues.Get("currentCollection")
		collection, ok := c.Collections[collectionName]
		if !ok {
			fmt.Println("Collection Name Invalid")
			return
		}

		var setOptions []string
		for key, _ := range collection.Fragments {
			fmt.Println(key)
			setOptions = append(setOptions, key)
		}
		//fmt.Println(setOptions)
		var pageData = struct {
			FragmentSets    []string
			CurrentSet      string
			Fragments       []Fragment
			CurrentFragment string
			FragmentDetails []*FragmentDetails
		}{
			FragmentSets:    setOptions,
			CurrentSet:      "",
			CurrentFragment: "",
		}
		tmpl.ExecuteTemplate(w, "fixture-fragment", pageData)
	}
}
