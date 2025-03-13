package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

// Does location get or stringifyLocation get used by template?
type GridClickDetails struct {
	CollectionName   string
	Location         []string
	GridType         string
	ScreenID         string // known if editing bp, either "screen" or "fragment"
	Y                int
	X                int
	DefaultTileColor string
	Selected         bool
	Tool             string
	SelectedAssetId  string // used for identifying multiple non-interactive grids? is a click detail not square detail
	haveASelection   bool
	selectedX        int
	selectedY        int
}

var CONNECTING_CHAR = "."

func (gridSquare GridClickDetails) stringifyLocation() string {
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
	details := createClickDetailsFromProps(properties, "area")
	collectionName := properties["currentCollection"]
	collection, ok := c.Collections[collectionName]
	if !ok {
		panic("No Collection")
	}

	// new func
	spaceName := details.Location[0]
	areaName := details.Location[1]
	space := c.spaceFromNames(details.CollectionName, spaceName)
	if space == nil {
		panic("No Space")
	}
	area := getAreaByName(space.Areas, areaName)

	result := collection.gridClickAction(details, area.Blueprint)
	io.WriteString(w, result)
	if result == "" {
		var pageData = GridDetails{
			MaterialGrid:     collection.generateMaterials(area.Blueprint.Tiles),
			InteractableGrid: collection.generateInteractables(area.Blueprint.Tiles),
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
	details := createClickDetailsFromProps(properties, "fragment")

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
	result := col.gridClickAction(details, fragment.Blueprint)
	io.WriteString(w, result)
	if result == "" {
		var pageData = GridDetails{
			MaterialGrid:     col.generateMaterials(fragment.Blueprint.Tiles),
			InteractableGrid: col.generateInteractables(fragment.Blueprint.Tiles),
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

func createClickDetailsFromProps(properties map[string]string, gridType string) GridClickDetails {
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
	tool, ok := properties["radio-tool"]
	if !ok {
		fmt.Println("No Tool Selected for grid click")
	}
	protoId := properties["selected-asset-id"]

	gridSelectedX, okX := properties["grid-selected-x"]
	selectedX := 0
	if okX && gridSelectedX != "" {
		selectedX, err = strconv.Atoi(gridSelectedX)
		if err != nil {
			panic("invalid selected x")
		}
	}
	gridSelectedY, okY := properties["grid-selected-y"]
	selectedY := 0
	if okY && gridSelectedY != "" {
		selectedY, err = strconv.Atoi(gridSelectedY)
		if err != nil {
			panic("invalid selected Y")
		}
	}
	haveASelection := okX && okY

	return GridClickDetails{CollectionName: currentCollection, Location: parts, GridType: gridType, ScreenID: sid, Y: yInt, X: xInt, DefaultTileColor: defaultTileColor, Tool: tool, SelectedAssetId: protoId, haveASelection: haveASelection, selectedX: selectedX, selectedY: selectedY}
}

// / Tools
func (col *Collection) gridClickAction(details GridClickDetails, blueprint *Blueprint) string {
	if details.Tool == "select" {
		// should oob update hiddens
		return col.gridSelect(details, blueprint.Tiles)
	} else if details.Tool == "replace" {
		selectedPrototype := col.getPrototypeOrCreateInvalid(details.SelectedAssetId)
		return col.gridReplace(details, blueprint.Tiles, selectedPrototype)
	} else if details.Tool == "fill" {
		selectedPrototype := col.getPrototypeOrCreateInvalid(details.SelectedAssetId)
		gridFill(details, blueprint.Tiles, selectedPrototype)
		return ""
	} else if details.Tool == "between" {
		selectedPrototype := col.getPrototypeOrCreateInvalid(details.SelectedAssetId)
		return col.gridFillBetween(details, blueprint.Tiles, selectedPrototype)
	} else if details.Tool == "place" {
		// Pull isSelected & location (selectedLocation) into hidden field
		fragment := col.getFragmentFromAssetId(details.SelectedAssetId)
		gridPlaceFragment(details, blueprint.Tiles, fragment)
	} else if details.Tool == "rotate" {
		gridRotate(details, blueprint.Tiles)
	} else if details.Tool == "place-blueprint" {
		if details.SelectedAssetId != "" {
			blueprint.Instructions = append(blueprint.Instructions, Instruction{
				ID:                 uuid.New().String(),
				X:                  details.X,
				Y:                  details.Y,
				GridAssetId:        details.SelectedAssetId,
				ClockwiseRotations: 0,
			})
		}
		for _, instruction := range blueprint.Instructions {
			col.applyInstruction(blueprint.Tiles, instruction)
		}

	} else if details.Tool == "interactable-replace" {
		interactable := col.findInteractableById(details.SelectedAssetId)
		return col.interactableReplace(details, blueprint.Tiles, interactable)
	} else if details.Tool == "interactable-delete" {
		return col.interactableReplace(details, blueprint.Tiles, nil)
	}
	return ""
}

func (col *Collection) getTileGridByAssetId(assetId string) [][]TileData {
	fragment := col.getFragmentById(assetId)
	if fragment != nil {
		return fragment.Blueprint.Tiles
	}
	out := make([][]TileData, 0)
	proto := col.findPrototypeById(assetId)
	if proto != nil {
		out = append(out, append(make([]TileData, 0), TileData{PrototypeId: assetId, Transformation: Transformation{}})) // })//} )
	}
	return out
}

func pasteTiles(y, x int, source, dest [][]TileData) {
	for i := range dest {
		if y+i >= len(source) {
			break
		}
		for j := range dest[i] {
			if x+j >= len(source[y+i]) {
				break
			}
			if dest[i][j].PrototypeId != "" {
				source[y+i][x+j].PrototypeId = dest[i][j].PrototypeId
				source[y+i][x+j].Transformation = dest[i][j].Transformation
			}
			if dest[i][j].InteractableId != "" {
				source[y+i][x+j].InteractableId = dest[i][j].InteractableId
			}
		}
	}
}

func clearTiles(y, x, height, width int, source [][]TileData) {
	for i := 0; i < height; i++ {
		if y+i >= len(source) {
			break
		}
		for j := 0; j < width; j++ {
			if x+j >= len(source[y+i]) {
				break
			}
			source[y+i][x+j].PrototypeId = ""
			source[y+i][x+j].InteractableId = ""
		}
	}
}

func (col *Collection) applyInstruction(source [][]TileData, instruction Instruction) {
	gridToApply := rotateTimesN(col.getTileGridByAssetId(instruction.GridAssetId), instruction.ClockwiseRotations)
	pasteTiles(instruction.Y, instruction.X, source, gridToApply)
}

func (col *Collection) getPrototypeOrCreateInvalid(protoId string) Prototype {
	proto := col.findPrototypeById(protoId)
	if proto == nil {
		fmt.Println("Requested invalid proto: " + protoId)
		return Prototype{ID: "INVALID-" + protoId, CssColor: "blue", Floor1Css: "green red-b thick"}
	}

	return *proto
}

func gridPlaceFragment(details GridClickDetails, modifications [][]TileData, selectedFragment Fragment) {
	for i := range selectedFragment.Blueprint.Tiles {
		if details.Y+i < len(modifications) {
			for j := range selectedFragment.Blueprint.Tiles[i] {
				if details.X+j < len(modifications[details.Y+i]) {
					modifications[details.Y+i][details.X+j] = selectedFragment.Blueprint.Tiles[i][j]
				}
			}
		}
	}
}

func (col *Collection) getFragmentFromAssetId(fragmentID string) Fragment {
	fragment := col.getFragmentById(fragmentID)
	if fragment == nil {
		panic("No Fragment with ID: " + fragmentID)
	}
	return *fragment
}

func (col *Collection) gridSelect(event GridClickDetails, grid [][]TileData) string {
	var buf bytes.Buffer
	if event.haveASelection && event.selectedY < len(grid) && event.selectedX < len(grid[0]) {
		selectedCell := grid[event.selectedY][event.selectedX]
		var pageData = struct {
			Material     Material
			ClickEvent   GridClickDetails
			Interactable *InteractableDescription
		}{
			Material: col.findPrototypeById(selectedCell.PrototypeId).applyTransform(selectedCell.Transformation),
			ClickEvent: GridClickDetails{
				Y:                event.selectedY,
				X:                event.selectedX,
				GridType:         event.GridType,
				DefaultTileColor: event.DefaultTileColor,
				Location:         event.Location,
				ScreenID:         event.ScreenID},
			Interactable: col.findInteractableById(selectedCell.InteractableId),
		}
		err := tmpl.ExecuteTemplate(&buf, "grid-square", pageData)
		if err != nil {
			fmt.Println(err)
		}
	}

	selectedCell := grid[event.Y][event.X]
	event.Selected = true
	var pageData = struct {
		Material     Material
		ClickEvent   GridClickDetails
		Interactable *InteractableDescription
	}{
		Material:     col.findPrototypeById(selectedCell.PrototypeId).applyTransform(selectedCell.Transformation),
		ClickEvent:   event,
		Interactable: col.findInteractableById(selectedCell.InteractableId),
	}
	err := tmpl.ExecuteTemplate(&buf, "grid-square", pageData)
	if err != nil {
		fmt.Println(err)
	}
	return buf.String()
}

func (col *Collection) gridReplace(event GridClickDetails, modifications [][]TileData, selectedProto Prototype) string {
	modifications[event.Y][event.X].PrototypeId = selectedProto.ID
	var buf bytes.Buffer
	var pageData = struct {
		Material     Material
		ClickEvent   GridClickDetails
		Interactable *InteractableDescription
	}{
		Material:     selectedProto.applyTransform(modifications[event.Y][event.X].Transformation),
		ClickEvent:   event,
		Interactable: col.findInteractableById(modifications[event.Y][event.X].InteractableId),
	}
	err := tmpl.ExecuteTemplate(&buf, "grid-square", pageData)
	if err != nil {
		fmt.Println(err)
	}
	return buf.String()
}

func (col *Collection) interactableReplace(event GridClickDetails, modifications [][]TileData, selectedInteractable *InteractableDescription) string {
	modifications[event.Y][event.X].InteractableId = ""
	if selectedInteractable != nil {
		modifications[event.Y][event.X].InteractableId = selectedInteractable.ID
	}
	selectedProto := col.getPrototypeOrCreateInvalid(modifications[event.Y][event.X].PrototypeId)
	var buf bytes.Buffer
	var pageData = struct {
		Material     Material
		ClickEvent   GridClickDetails
		Interactable *InteractableDescription
	}{
		Material:     selectedProto.applyTransform(modifications[event.Y][event.X].Transformation),
		ClickEvent:   event,
		Interactable: col.findInteractableById(modifications[event.Y][event.X].InteractableId),
	}
	err := tmpl.ExecuteTemplate(&buf, "grid-square", pageData)
	if err != nil {
		fmt.Println(err)
	}
	return buf.String()
}

func gridFill(event GridClickDetails, grid [][]TileData, selectedPrototype Prototype) {
	targetId := grid[event.Y][event.X].PrototypeId
	seen := make([][]bool, len(grid))
	for row := range seen {
		seen[row] = make([]bool, len(grid[row]))
	}
	fill(event, grid, selectedPrototype, seen, targetId)
}

func fill(event GridClickDetails, modifications [][]TileData, selectedPrototype Prototype, seen [][]bool, targetId string) {
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

// include selection in gridClickDetails
func (col *Collection) gridFillBetween(event GridClickDetails, modifications [][]TileData, selectedPrototype Prototype) string {
	if !event.haveASelection {
		col.gridSelect(event, modifications)
	}
	var lowx, lowy, highx, highy int
	if event.Y <= event.selectedY {
		lowy = event.Y
		highy = event.selectedY
	} else {
		lowy = event.selectedY
		highy = event.Y
	}
	if event.X <= event.selectedX {
		lowx = event.X
		highx = event.selectedX
	} else {
		lowx = event.selectedX
		highx = event.X
	}
	output := ""
	for i := lowy; i <= highy; i++ {
		for j := lowx; j <= highx; j++ {
			// unsafe out of bounds
			newEvent := event
			newEvent.Y = i
			newEvent.X = j
			output += col.gridReplace(newEvent, modifications, selectedPrototype)
		}
	}
	output += col.gridSelect(event, modifications)
	return output
}

func gridRotate(event GridClickDetails, modifications [][]TileData) {
	transformation := &modifications[event.Y][event.X].Transformation
	transformation.ClockwiseRotations = mod(transformation.ClockwiseRotations+1, 4)
}

///

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
			FragmentSets []string
			CurrentSet   string
		}{
			FragmentSets: setOptions,
			CurrentSet:   "",
		}
		tmpl.ExecuteTemplate(w, "fixture-fragment", pageData)
	}
	if fixtureType == "prototype" {
		tmpl.ExecuteTemplate(w, "fixture-prototype", c.prototypeSelectFromRequest(r))

	}
	if fixtureType == "transformation" {
		tmpl.ExecuteTemplate(w, "fixture-transformation", nil)
	}
	if fixtureType == "blueprint" {
		c.getBlueprint(w, r) // only gets blueprint for area
	}
	if fixtureType == "interactable" {
		collectionName := queryValues.Get("currentCollection")
		collection, ok := c.Collections[collectionName]
		if !ok {
			fmt.Println("Collection Name Invalid")
			return
		}
		var setOptions []string
		for key := range collection.InteractableSets {
			setOptions = append(setOptions, key)
		}

		var pageData = struct {
			SetNames      []string
			CurrentSet    string
			Interactables []InteractableDescription
		}{
			SetNames:      setOptions,
			CurrentSet:    "",
			Interactables: nil,
		}
		err := tmpl.ExecuteTemplate(w, "fixture-interactables", pageData)
		if err != nil {
			fmt.Println(err)
		}
	}

}
