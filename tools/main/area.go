package main

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
)

type AreaDescription struct {
	Name           string      `json:"name"`
	Safe           bool        `json:"safe"`
	Blueprint      *Blueprint  `json:"blueprint"`
	Transports     []Transport `json:"transports"`
	North          string      `json:"north,omitempty"`
	South          string      `json:"south,omitempty"`
	East           string      `json:"east,omitempty"`
	West           string      `json:"west,omitempty"`
	MapId          string      `json:"mapId"`
	LoadStrategy   string      `json:"loadStrategy"`
	SpawnStrategy  string      `json:"spawnStrategy"`
	BroadcastGroup string      `json:"broadcastGroup,omitempty"`
	Weather        string      `json:"weather,omitempty"`
}

// Import from the other project instead? Or import from here. Transport too
type AreaOutput struct {
	Name             string                       `json:"name"`
	Safe             bool                         `json:"safe"`
	Tiles            [][]Material                 `json:"tiles"`
	Interactables    [][]*InteractableDescription `json:"interactables"`
	Transports       []Transport                  `json:"transports"`
	DefaultTileColor string                       `json:"defaultTileColor"`
	North            string                       `json:"north,omitempty"`
	South            string                       `json:"south,omitempty"`
	East             string                       `json:"east,omitempty"`
	West             string                       `json:"west,omitempty"`
	MapId            string                       `json:"mapId,omitempty"`
	LoadStrategy     string                       `json:"loadStrategy,omitempty"`
	SpawnStrategy    string                       `json:"spawnStrategy"`
	BroadcastGroup   string                       `json:"broadcastGroup,omitempty"`
	Weather          string                       `json:"weather,omitempty"`
}

type AreaEditPageData struct {
	AreaWithGrid
	PrototypeSelect PrototypeSelectPage
}

type AreaWithGrid struct {
	GridDetails    GridDetails
	SelectedArea   AreaDescription
	NavHasHadClick bool
}

// ///////////////////////////////////////////////////////////
// Areas

func (c Context) areasHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		c.getAreas(w, r)
	}
	if r.Method == "POST" {
		c.postAreas(w, r)
	}
}

func (c Context) getAreas(w http.ResponseWriter, r *http.Request) {
	queryValues := r.URL.Query()
	collectionName := queryValues.Get("currentCollection")
	col, ok := c.Collections[collectionName]
	if !ok {
		panic("No collection rn")
	}

	space := c.spaceFromGET(r)

	var input = struct {
		Collection *Collection
		Space      *Space
	}{Collection: col, Space: space}
	err := tmpl.ExecuteTemplate(w, "areas", input)
	if err != nil {
		fmt.Println(err)
	}
}

func (c Context) postAreas(w http.ResponseWriter, r *http.Request) {
	properties, _ := requestToProperties(r)
	name := properties["new-area-name"]
	safe := (properties["safe"] == "on")
	defaultTileColor := properties["default-tile-color"]
	defaultTileColor1 := properties["default-tile-color1"]
	collectionName := properties["currentCollection"]
	spaceName := properties["currentSpace"]
	panicIfAnyEmpty("POST to /area", collectionName, spaceName, name)

	height, _ := strconv.Atoi(properties["area-height"])
	width, _ := strconv.Atoi(properties["area-width"])

	tiles := make([][]TileData, height)
	for i := range tiles {
		tiles[i] = make([]TileData, width)
	}

	blueprint := &Blueprint{Tiles: tiles, DefaultTileColor: defaultTileColor, DefaultTileColor1: defaultTileColor1, Instructions: make([]Instruction, 0)}

	space := c.spaceFromNames(collectionName, spaceName)
	space.Areas = append(space.Areas, AreaDescription{Name: name, Safe: safe, Blueprint: blueprint, Transports: make([]Transport, 0)})
	io.WriteString(w, "<h3>Done.</h3>")
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
	queryValues := r.URL.Query()
	space := c.spaceFromGET(r)
	name := queryValues.Get("area-name-selected")
	selectedArea := getAreaByName(space.Areas, name)
	if selectedArea == nil {
		io.WriteString(w, "<h2>no Area</h2>")
		return
	}

	collection := c.collectionFromGet(r)

	var setOptions []string
	for key := range collection.PrototypeSets {
		setOptions = append(setOptions, key)
	}

	modifications := collection.generateMaterials(selectedArea.Blueprint)

	var pageData = AreaEditPageData{
		AreaWithGrid: AreaWithGrid{
			GridDetails: GridDetails{
				MaterialGrid:     modifications,
				InteractableGrid: collection.generateInteractables(selectedArea.Blueprint.Tiles),
				Location:         locationStringFromArea(selectedArea, space.Name),
				GridType:         "area",
				ScreenID:         "screen",
			},
			SelectedArea:   *selectedArea,
			NavHasHadClick: false,
		},
		PrototypeSelect: PrototypeSelectPage{
			PrototypeSets: setOptions,
			CurrentSet:    "",
			Prototypes:    nil,
		},
	}
	err := tmpl.ExecuteTemplate(w, "area-edit", pageData)
	if err != nil {
		fmt.Println(err)
	}
}

// For tiny nav with minimal disruption:
func (c *Context) areaGridHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		c.getAreaGrid(w, r)
	}
}

func (c *Context) getAreaGrid(w http.ResponseWriter, r *http.Request) {
	queryValues := r.URL.Query()
	space := c.spaceFromGET(r)
	name := queryValues.Get("area-name-selected")
	selectedArea := getAreaByName(space.Areas, name)
	if selectedArea == nil {
		io.WriteString(w, "<h2>no Area</h2>")
		return
	}

	collection := c.collectionFromGet(r)

	modifications := collection.generateMaterials(selectedArea.Blueprint)

	areaWithGrid := AreaWithGrid{
		GridDetails: GridDetails{
			MaterialGrid:     modifications,
			InteractableGrid: collection.generateInteractables(selectedArea.Blueprint.Tiles),
			Location:         locationStringFromArea(selectedArea, space.Name),
			GridType:         "area",
			ScreenID:         "screen",
		},
		SelectedArea:   *selectedArea,
		NavHasHadClick: true,
	}

	err := tmpl.ExecuteTemplate(w, "area-grid", areaWithGrid)
	if err != nil {
		fmt.Println(err)
	}
}

func (c Context) postArea(w http.ResponseWriter, r *http.Request) {
	properties, _ := requestToProperties(r)
	name := properties["originalAreaName"]
	newName := properties["areaName"]
	safe := (properties["safe"] == "on")
	defaultTileColor := properties["defaultTileColor"]
	defaultTileColor1 := properties["defaultTileColor1"]
	weather := properties["weather"]
	loadStrategy := properties["loadStrategy"]
	broadcastGroup := properties["broadcastGroup"]
	spawnStrategy := properties["spawnStrategy"]
	collectionName := properties["currentCollection"]
	spaceName := properties["currentSpace"]
	panicIfAnyEmpty("POST to /area", collectionName, spaceName, name)

	space := c.spaceFromNames(collectionName, spaceName)

	selectedArea := getAreaByName(space.Areas, name)
	if selectedArea == nil {
		panic("Orignal area data has been lost or submitted orignal name is invalid")
	}

	if newName == name {
		// These changes will get lost if concurrent with name change
		selectedArea.Safe = safe
		selectedArea.Blueprint.DefaultTileColor = defaultTileColor
		selectedArea.Blueprint.DefaultTileColor1 = defaultTileColor1
		selectedArea.Weather = weather
		selectedArea.LoadStrategy = loadStrategy
		selectedArea.SpawnStrategy = spawnStrategy
		selectedArea.BroadcastGroup = broadcastGroup
	} else {
		if getAreaByName(space.Areas, newName) != nil {
			panic("Invalid name") // This check doesn't look at other spaces
		}
		newBlueprint := copyBlueprint(selectedArea.Blueprint)
		area := AreaDescription{Name: newName, Safe: safe, Blueprint: &newBlueprint, Transports: append([]Transport{}, selectedArea.Transports...)}
		space.Areas = append(space.Areas, area)
	}

	collection, ok := c.Collections[collectionName]
	if !ok {
		panic("no collection")
	}
	collection.Spaces[spaceName] = space // Maybe uneeded because space is a pointer?
	collection.saveSpace(spaceName)

	io.WriteString(w, `<h2>Success</h2>`)
}

func copyBlueprint(bp *Blueprint) Blueprint {
	tiles := make([][]TileData, len(bp.Tiles))
	for i := range tiles {
		tiles[i] = append(tiles[i], bp.Tiles[i]...)
	}
	return Blueprint{Tiles: tiles, DefaultTileColor: bp.DefaultTileColor, Instructions: append([]Instruction{}, bp.Instructions...)}

}

func (c Context) newAreaHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		queryValues := r.URL.Query()
		colName := queryValues.Get("currentCollection")
		spaceName := queryValues.Get("currentSpace")
		fmt.Println("Collection Name: " + colName)
		fmt.Println("Space Name Name: " + spaceName)
		if col, ok := c.Collections[colName]; ok {
			col.getNewArea(w, r)
		}
	}
}

func (col Collection) getNewArea(w http.ResponseWriter, _ *http.Request) {
	fmt.Println("HI")
	err := tmpl.ExecuteTemplate(w, "area-new", nil)
	if err != nil {
		fmt.Println(err)
	}
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
		Area    *AreaDescription
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

func (c Context) getAreaDisplay(w http.ResponseWriter, _ *http.Request) {
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
	space := c.spaceFromNames(collectionName, spaceName)

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
	space := c.spaceFromNames(collectionName, spaceName)
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
