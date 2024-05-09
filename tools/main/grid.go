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
}

var CONNECTING_CHAR = "."

func (gridSquare GridSquareDetails) stringifyLocation() string {
	return strings.Join(gridSquare.Location, CONNECTING_CHAR)
}

func locationStringFromArea(area *AreaDescription, spacename string) string {
	return spacename + CONNECTING_CHAR + area.Name
}

func (c *Context) gridEditHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		c.getGridEdit(w, r)
	}
}

func (c *Context) getGridEdit(w http.ResponseWriter, r *http.Request) {
	queryValues := r.URL.Query()
	collectionName := queryValues.Get("currentCollection")
	col, ok := c.Collections[collectionName]
	if !ok {
		panic("invalid collection")
	}
	tmpl.ExecuteTemplate(w, "grid-modify", col.getProtoSelect())
}

func (c Context) gridClickAreaHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		return
	}

	properties, _ := requestToProperties(r)
	details := createGridSquareDetails(properties, "area")
	collectionName := properties["currentCollection"]
	collection, ok := c.Collections[collectionName]
	if !ok {
		panic("No Collection")
	}

	// new func
	spaceName := details.Location[0]
	areaName := details.Location[1]
	space := c.getSpace(details.CollectionName, spaceName)
	if space == nil {
		panic("No Space")
	}
	area := getAreaByName(space.Areas, areaName)

	result := c.gridAction(details, area.Tiles, properties)
	io.WriteString(w, result)
	if result == "" {
		var pageData = GridDetails{
			MaterialGrid:     collection.generateMaterials(area.Tiles),
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
			MaterialGrid:     col.generateMaterials(fragment.Tiles),
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
	}
	x, ok := properties["x"]
	if !ok {
		panic("No x")
	}
	xInt, err := strconv.Atoi(x)
	if err != nil {
		panic("invalid x")
	}
	y, ok := properties["y"]
	if !ok {
		panic("No y")
	}
	yInt, err := strconv.Atoi(y)
	if err != nil {
		panic("invalid y")
	}
	sid, ok := properties["sid"]
	if !ok {
		panic("No sid")
	}
	defaultTileColor, ok := properties["default-tile-color"]
	if !ok {
		panic("location")
	}
	location, ok := properties["location"]
	if !ok {
		panic("location")
	}
	parts := strings.Split(location, ".")
	if len(parts) < 2 {
		fmt.Println("Invalid Location")
	}
	return GridSquareDetails{CollectionName: currentCollection, Location: parts, GridType: gridType, ScreenID: sid, Y: yInt, X: xInt, DefaultTileColor: defaultTileColor}
}

// / Tools
func (c *Context) gridAction(details GridSquareDetails, grid [][]TileData, properties map[string]string) string {
	tool, ok := properties["radio-tool"]
	if !ok {
		panic("No Tool Selected")
	}
	currentCollection, ok := properties["currentCollection"]
	if !ok {
		panic("No Collection Name")
	}
	col, ok := c.Collections[currentCollection]
	if !ok {
		panic("no collection")
	}
	if tool == "select" {
		// should oob update hiddens
		return col.gridSelect(details, grid)
	} else if tool == "replace" {
		selectedPrototype := col.getPrototypeFromRequestProperties(properties)
		return gridReplace(details, grid, selectedPrototype)
	} else if tool == "fill" {
		selectedPrototype := col.getPrototypeFromRequestProperties(properties)
		gridFill(details, grid, selectedPrototype)
		return ""
	} else if tool == "between" {
		selectedPrototype := col.getPrototypeFromRequestProperties(properties)
		return col.gridFillBetween(details, grid, selectedPrototype)
	} else if tool == "place" {
		// Pull isSelected & location (selectedLocation) into hidden field
		fragment := col.getFragmentFromRequestProperties(properties)
		gridPlaceFragment(details, grid, fragment)
	} else if tool == "rotate" {
		gridRotate(details, grid)
	}
	return ""
}

func (col *Collection) getPrototypeFromRequestProperties(properties map[string]string) Prototype {
	protoId := properties["selected-prototype-id"]
	//return *col.Prototypes[protoId]

	return *col.findPrototypeById(protoId)
}

func gridPlaceFragment(details GridSquareDetails, modifications [][]TileData, selectedFragment Fragment) {
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

func (col *Collection) getFragmentFromRequestProperties(properties map[string]string) Fragment {
	setName, ok := properties["fragment-set"]
	if !ok {
		panic("no set name")
	}
	set, ok := col.Fragments[setName]
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

func (col *Collection) gridSelect(event GridSquareDetails, grid [][]TileData) string {
	//output := ""
	var buf bytes.Buffer
	if haveSelection {
		selectedCell := grid[selectedY][selectedX]
		var pageData = struct {
			Material   Material
			ClickEvent GridSquareDetails
		}{
			Material: col.findPrototypeById(selectedCell.PrototypeId).applyTransform(selectedCell.Transformation),
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
	selectedCell := grid[selectedY][selectedX]
	event.Selected = true
	var pageData = struct {
		Material   Material
		ClickEvent GridSquareDetails
	}{
		Material:   col.findPrototypeById(selectedCell.PrototypeId).applyTransform(selectedCell.Transformation),
		ClickEvent: event,
	}
	err := tmpl.ExecuteTemplate(&buf, "grid-square", pageData)
	if err != nil {
		fmt.Println(err)
	}
	return buf.String()
}

func gridReplace(event GridSquareDetails, modifications [][]TileData, selectedProto Prototype) string {
	modifications[event.Y][event.X].PrototypeId = selectedProto.ID
	var buf bytes.Buffer
	var pageData = struct {
		Material   Material
		ClickEvent GridSquareDetails
	}{
		Material:   selectedProto.applyTransform(modifications[event.Y][event.X].Transformation),
		ClickEvent: event,
	}
	err := tmpl.ExecuteTemplate(&buf, "grid-square", pageData)
	if err != nil {
		fmt.Println(err)
	}
	return buf.String()
}

func gridFill(event GridSquareDetails, grid [][]TileData, selectedPrototype Prototype) {
	targetId := grid[event.Y][event.X].PrototypeId
	seen := make([][]bool, len(grid))
	for row := range seen {
		seen[row] = make([]bool, len(grid[row]))
	}
	fill(event, grid, selectedPrototype, seen, targetId)
}

func fill(event GridSquareDetails, modifications [][]TileData, selectedPrototype Prototype, seen [][]bool, targetId string) {
	seen[event.Y][event.X] = true
	modifications[event.Y][event.X].PrototypeId = selectedPrototype.ID
	deltas := []int{-1, 1}
	for _, i := range deltas {
		if event.Y+i >= 0 && event.Y+i < len(modifications) {
			shouldfill := !seen[event.Y+i][event.X] && modifications[event.Y+i][event.X].PrototypeId == targetId
			if shouldfill {
				newEvent := event
				newEvent.Y += i
				fill(newEvent, modifications, selectedPrototype, seen, targetId)
			}
		}
		if event.X+i >= 0 && event.X+i < len(modifications[event.Y]) {
			shouldfill := !seen[event.Y][event.X+i] && modifications[event.Y][event.X+i].PrototypeId == targetId
			if shouldfill {
				newEvent := event
				newEvent.X += i
				fill(newEvent, modifications, selectedPrototype, seen, targetId)
			}
		}
	}
}

func (col *Collection) gridFillBetween(event GridSquareDetails, modifications [][]TileData, selectedPrototype Prototype) string {
	if !haveSelection {
		col.gridSelect(event, modifications)
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
			output += gridReplace(newEvent, modifications, selectedPrototype)
		}
	}
	output += col.gridSelect(event, modifications)
	return output
}

func gridRotate(event GridSquareDetails, modifications [][]TileData) {
	transformation := &modifications[event.Y][event.X].Transformation
	transformation.ClockwiseRotations = mod(transformation.ClockwiseRotations+1, 4)
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

	if fixtureType == "fragment" {
		collectionName := queryValues.Get("currentCollection")
		collection, ok := c.Collections[collectionName]
		if !ok {
			fmt.Println("Collection Name Invalid")
			return
		}

		var setOptions []string
		for key := range collection.Fragments {
			setOptions = append(setOptions, key)
		}

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
	if fixtureType == "prototype" {
		tmpl.ExecuteTemplate(w, "fixture-prototype", c.prototypeSelectFromRequest(r))

	}
	if fixtureType == "transformation" {
		tmpl.ExecuteTemplate(w, "fixture-transformation", nil)
	}
}

func exampleSquareFromMaterial(material Material) string {
	overlay := fmt.Sprintf(`<div class="box floor1 %s"></div><div class="box floor2 %s"></div><div class="box ceiling1 %s"></div><div class="box ceiling2 %s"></div>`, material.Floor1Css, material.Floor2Css, material.Ceiling1Css, material.Ceiling2Css)
	idHiddenInput := fmt.Sprintf(`<input name="selected-material" type="hidden" value="%d" />`, material.ID)
	return fmt.Sprintf(`<div class="grid-square %s" name="selected-material">%s%s</div>`, material.CssColor, overlay, idHiddenInput)
}
