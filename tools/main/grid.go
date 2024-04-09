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
	Location         string
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
	fmt.Printf("%d %d %s %s\n", event.Y, event.X, event.Location, event.DefaultTileColor)

	parts := strings.Split(event.Location, "_")
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
			// should oob update hiddens
			io.WriteString(w, c.gridSelect(event, area.Tiles))
		} else if selectedTool == "replace" {
			selectedMaterial := c.getMaterialFromRequestProperties(properties)
			io.WriteString(w, gridReplace(event, area.Tiles, selectedMaterial))
		} else if selectedTool == "fill" {
			selectedMaterial := c.getMaterialFromRequestProperties(properties)
			gridFill(event, area.Tiles, selectedMaterial)
			var pageData = GridDetails{
				MaterialGrid:     c.DereferenceIntMatrix(area.Tiles),
				DefaultTileColor: area.DefaultTileColor,
				Location:         event.Location,
				ScreenID:         event.ScreenID,
				GridType:         "area",
				Oob:              true,
			}

			err := tmpl.ExecuteTemplate(w, "grid", pageData)
			if err != nil {
				fmt.Println(err)
			}
		} else if selectedTool == "between" {
			selectedMaterial := c.getMaterialFromRequestProperties(properties)
			selectedMaterial.ID += 0
			io.WriteString(w, c.gridFillBetween(event, area.Tiles, selectedMaterial))
		}
	}
}

func (c Context) gridClickFragmentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		return
	}
	// Get [][]int from Fragment and pass to tool for mutation
}

// / Tools
func (c Context) gridSelect(event ClickEvent, modifications [][]int) string {
	//output := ""
	var buf bytes.Buffer
	if haveSelection {
		var pageData = struct {
			Material   Material
			ClickEvent ClickEvent
		}{
			Material: c.materials[modifications[selectedY][selectedX]],
			ClickEvent: ClickEvent{
				Y:                selectedY,
				X:                selectedX,
				GridType:         "area",
				DefaultTileColor: event.DefaultTileColor,
				Location:         event.Location,
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
		Material:   c.materials[modifications[selectedY][selectedX]],
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

func (c Context) gridFillBetween(event ClickEvent, modifications [][]int, selectedMaterial Material) string {
	if !haveSelection {
		c.gridSelect(event, modifications)
	}
	var lowx, lowy, highx, highy int
	if event.Y <= selectedY {
		lowy = event.Y
		highy = selectedY
	} else {
		lowy = selectedY
		highy = event.Y
	}
	if event.X <= selectedX {
		lowx = event.X
		highx = selectedX
	} else {
		lowx = selectedX
		highx = event.X
	}
	output := ""
	for i := lowy; i <= highy; i++ {
		for j := lowx; j <= highx; j++ {
			newEvent := event
			newEvent.Y = i
			newEvent.X = j
			output += gridReplace(newEvent, modifications, selectedMaterial)
			//output += replaceSquare(i, j, selectedMaterial, defaultTileColor)
		}
	}
	output += c.gridSelect(event, modifications)
	return output
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
		for key := range collection.Fragments {
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

func (c Context) getMaterialFromRequestProperties(properties map[string]string) Material {
	//properties, _ := requestToProperties(r)
	fmt.Println(properties)
	fmt.Println(properties["selected-material"])
	selectedMaterialId, err := strconv.Atoi(properties["selected-material"])
	if err != nil {
		fmt.Println(err)
		selectedMaterialId = 0
	}
	return c.materials[selectedMaterialId]
}

func dataFromClickRequest(r *http.Request, gridtype string) (ClickEvent, bool) {
	yCoord, _ := strconv.Atoi(r.Header["Y"][0])
	xCoord, _ := strconv.Atoi(r.Header["X"][0])

	sidHeaders := r.Header["Sid"]
	sid := sidHeaders[0]
	fmt.Println(sid)

	LocationHeaders := r.Header["Location"]
	if len(LocationHeaders) == 0 {
		fmt.Println("No Location headers")
		return ClickEvent{Y: yCoord, X: xCoord, GridType: gridtype, DefaultTileColor: "", Location: ""}, false
	}
	location := LocationHeaders[0]
	fmt.Println(location)

	defaultTileColorHeaders := r.Header["Default-Tile-Color"]
	if len(defaultTileColorHeaders) == 0 {
		fmt.Println("No screen id headers")
		return ClickEvent{Y: yCoord, X: xCoord, GridType: gridtype, DefaultTileColor: "", Location: location}, false
	}
	dtc := defaultTileColorHeaders[0]
	fmt.Println(location)

	return ClickEvent{Y: yCoord, X: xCoord, GridType: gridtype, DefaultTileColor: dtc, Location: location, ScreenID: sid}, true
}

func exampleSquareFromMaterial(material Material) string {
	overlay := fmt.Sprintf(`<div class="box floor1 %s"></div><div class="box floor2 %s"></div><div class="box ceiling1 %s"></div><div class="box ceiling2 %s"></div>`, material.Floor1Css, material.Floor2Css, material.Ceiling1Css, material.Ceiling2Css)
	idHiddenInput := fmt.Sprintf(`<input name="selected-material" type="hidden" value="%d" />`, material.ID)
	return fmt.Sprintf(`<div class="grid-square %s" name="selected-material">%s%s</div>`, material.CssColor, overlay, idHiddenInput)
}
