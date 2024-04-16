package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

// Does location get or stringifyLocation get used by template?
type GridSquareDetails struct {
	CollectionName   string
	Location         []string
	GridType         string
	ScreenID         string
	Y                int
	X                int
	DefaultTileColor string
	Selected         bool
	//SelectedTool     string
}

func (gridSquare GridSquareDetails) stringifyLocation() string {
	return strings.Join(gridSquare.Location, ".")
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

	properties, _ := requestToProperties(r)
	details := createGridSquareDetails(properties, "area")

	// new func
	//parts := strings.Split(event.Location, "_")
	//if len(parts) < 2 {
	//	fmt.Println("Invalid Sid")
	//}
	spaceName := details.Location[0]
	areaName := details.Location[1]
	space := c.getSpace(details.CollectionName, spaceName)
	if space == nil {
		panic("Hey")
	}
	area := getAreaByName(space.Areas, areaName)
	fmt.Println("Have: " + area.Name)

	//c.gridAction(event, area.Tiles, properties)
	result := c.gridAction(details, area.Tiles, properties)
	io.WriteString(w, result)
	if result == "" {
		var pageData = GridDetails{
			MaterialGrid:     c.DereferenceIntMatrix(area.Tiles),
			DefaultTileColor: details.DefaultTileColor,
			Location:         details.stringifyLocation(),
			ScreenID:         details.ScreenID,
			GridType:         details.GridType,
			Oob:              true,
		}

		err := tmpl.ExecuteTemplate(w, "grid", pageData)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func (c Context) gridClickFragmentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		return
	}

	properties, _ := requestToProperties(r)
	details := createGridSquareDetails(properties, "fragment")

	// new func

	setName := details.Location[0]
	fragmentName := details.Location[1]

	col, ok := c.Collections[details.CollectionName]
	if !ok {
		panic("no collection")
	}
	set, ok := col.Fragments[setName]
	if !ok {
		panic("no Set")
	}
	fragment := getFragmentByName(set, fragmentName)
	result := c.gridAction(details, fragment.Tiles, properties)
	io.WriteString(w, result)
	if result == "" {
		var pageData = GridDetails{
			MaterialGrid:     c.DereferenceIntMatrix(fragment.Tiles),
			DefaultTileColor: details.DefaultTileColor,
			Location:         details.stringifyLocation(),
			ScreenID:         details.ScreenID,
			GridType:         details.GridType,
			Oob:              true,
		}

		err := tmpl.ExecuteTemplate(w, "grid", pageData)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func createGridSquareDetails(properties map[string]string, gridType string) GridSquareDetails {
	currentCollection, ok := properties["currentCollection"]
	if !ok {
		panic("No Collection")
		//return
	}
	x, ok := properties["x"]
	if !ok {
		panic("No x")
		//return
	}
	xInt, err := strconv.Atoi(x)
	if err != nil {
		panic("invalid x")
	}
	y, ok := properties["y"]
	if !ok {
		panic("No y")
		//return
	}
	yInt, err := strconv.Atoi(y)
	if err != nil {
		panic("invalid y")
	}
	sid, ok := properties["sid"]
	if !ok {
		panic("No sid")
		//return
	}
	defaultTileColor, ok := properties["default-tile-color"]
	if !ok {
		panic("location")
		//return
	}
	location, ok := properties["location"]
	if !ok {
		panic("location")
		//return
	}
	parts := strings.Split(location, ".")
	if len(parts) < 2 {
		fmt.Println("Invalid Location")
	}
	return GridSquareDetails{CollectionName: currentCollection, Location: parts, GridType: gridType, ScreenID: sid, Y: yInt, X: xInt, DefaultTileColor: defaultTileColor}
}

// / Tools
func (c Context) gridAction(details GridSquareDetails, grid [][]int, properties map[string]string) string {
	tool, ok := properties["radio-tool"]
	if !ok {
		panic("No Tool Selected")
		//return
	}
	if tool == "select" {
		// should oob update hiddens
		return c.gridSelect(details, grid)
	} else if tool == "replace" {
		selectedMaterial := c.getMaterialFromRequestProperties(properties)
		return gridReplace(details, grid, selectedMaterial)
	} else if tool == "fill" {
		selectedMaterial := c.getMaterialFromRequestProperties(properties)
		gridFill(details, grid, selectedMaterial)
		return ""
	} else if tool == "between" {
		selectedMaterial := c.getMaterialFromRequestProperties(properties)
		//selectedMaterial.ID += 0
		return c.gridFillBetween(details, grid, selectedMaterial)
	} else if tool == "place" {
		// Pull isSelected & location (selectedLocation) into hidden field
		//panic("Tell me your status") // No response
		fragment := c.getFragmentFromRequestProperties(properties)
		c.gridPlaceFragment(details, grid, fragment)
	}
	return ""
}

func (c Context) gridPlaceFragment(details GridSquareDetails, modifications [][]int, selectedFragment Fragment) {
	for i := range selectedFragment.Tiles {
		if details.Y+i < len(modifications) {
			for j := range selectedFragment.Tiles[i] {
				if details.X+j < len(modifications[i]) {
					modifications[details.Y+i][details.X+j] = selectedFragment.Tiles[i][j]
				}
			}
		}
	}
}

func (c Context) getFragmentFromRequestProperties(properties map[string]string) Fragment {
	currentCollection, ok := properties["currentCollection"]
	if !ok {
		panic("No Collection Name")
		//return
	}
	collection, ok := c.Collections[currentCollection]
	if !ok {
		panic("no collection")
	}
	setName, ok := properties["fragment-set"]
	if !ok {
		panic("no set name")
	}
	set, ok := collection.Fragments[setName]
	if !ok {
		panic("no set")
	}
	fragmentName, ok := properties["selected-fragment-name"]
	if !ok {
		panic("no fragment name")
	}
	fragment := getFragmentByName(set, fragmentName)
	if fragment == nil {
		panic("No Fragment")
	}
	return *fragment
}

func (c Context) gridSelect(event GridSquareDetails, modifications [][]int) string {
	//output := ""
	var buf bytes.Buffer
	if haveSelection {
		var pageData = struct {
			Material   Material
			ClickEvent GridSquareDetails
		}{
			Material: c.materials[modifications[selectedY][selectedX]],
			ClickEvent: GridSquareDetails{
				Y:                selectedY,
				X:                selectedX,
				GridType:         event.GridType,
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
		ClickEvent GridSquareDetails
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

func gridReplace(event GridSquareDetails, modifications [][]int, selectedMaterial Material) string {
	modifications[event.Y][event.X] = selectedMaterial.ID
	var buf bytes.Buffer
	var pageData = struct {
		Material   Material
		ClickEvent GridSquareDetails
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

func gridFill(event GridSquareDetails, modifications [][]int, selectedMaterial Material) {
	targetId := modifications[event.Y][event.X]
	seen := make([][]bool, len(modifications))
	for row := range seen {
		seen[row] = make([]bool, len(modifications[row]))
	}
	fill(event, modifications, selectedMaterial, seen, targetId)
}

func fill(event GridSquareDetails, modifications [][]int, selectedMaterial Material, seen [][]bool, targetId int) {
	seen[event.Y][event.X] = true
	modifications[event.Y][event.X] = selectedMaterial.ID
	deltas := []int{-1, 1}
	for _, i := range deltas {
		if event.Y+i >= 0 && event.Y+i < len(modifications) {
			shouldfill := !seen[event.Y+i][event.X] && modifications[event.Y+i][event.X] == targetId
			if shouldfill {
				newEvent := event
				newEvent.Y += i
				fill(newEvent, modifications, selectedMaterial, seen, targetId)
			}
		}
		if event.X+i >= 0 && event.X+i < len(modifications[event.Y]) {
			shouldfill := !seen[event.Y][event.X+i] && modifications[event.Y][event.X+i] == targetId
			if shouldfill {
				newEvent := event
				newEvent.X += i
				fill(newEvent, modifications, selectedMaterial, seen, targetId)
			}
		}
	}
}

func (c Context) gridFillBetween(event GridSquareDetails, modifications [][]int, selectedMaterial Material) string {
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
		tmpl.ExecuteTemplate(w, "fixture-material", c.materials)
	}
	if fixtureType == "fragment" {
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

		// Make type
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
	fmt.Println(properties["selected-material"])
	selectedMaterialId, err := strconv.Atoi(properties["selected-material"])
	if err != nil {
		fmt.Println(err)
		selectedMaterialId = 0
	}
	return c.materials[selectedMaterialId]
}

func exampleSquareFromMaterial(material Material) string {
	overlay := fmt.Sprintf(`<div class="box floor1 %s"></div><div class="box floor2 %s"></div><div class="box ceiling1 %s"></div><div class="box ceiling2 %s"></div>`, material.Floor1Css, material.Floor2Css, material.Ceiling1Css, material.Ceiling2Css)
	idHiddenInput := fmt.Sprintf(`<input name="selected-material" type="hidden" value="%d" />`, material.ID)
	return fmt.Sprintf(`<div class="grid-square %s" name="selected-material">%s%s</div>`, material.CssColor, overlay, idHiddenInput)
}
