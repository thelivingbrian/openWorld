package main

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
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
var modifications [][]Material

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

func (c Context) areaHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		c.getArea(w, r)
	}
	if r.Method == "POST" {
		c.postArea(w, r)
	}
}

func (c Context) getArea(w http.ResponseWriter, r *http.Request) {
	selectedArea := c.areaFromGET(r)
	if selectedArea == nil {
		io.WriteString(w, "<h2>no Area</h2>")
		return
	}

	modifications = c.AreaToMaterialGrid(*selectedArea)

	var pageData = PageData{
		GridDetails: GridDetails{
			MaterialGrid:     modifications,
			DefaultTileColor: selectedArea.DefaultTileColor,
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
	out := make([][]Material, len(area.Tiles))
	for y := range area.Tiles {
		out[y] = make([]Material, len(area.Tiles[y]))
		for x := range area.Tiles[y] {
			out[y][x] = c.materials[area.Tiles[y][x]]
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

	if len(modifications) == 0 {
		return
	}

	tiles := make([][]int, len(modifications))
	for y := range modifications {
		tiles[y] = make([]int, len(modifications[y]))
		for x, material := range modifications[y] {
			tiles[y][x] = material.ID
		}
	}

	selectedArea := getAreaByName(space.Areas, name)
	if selectedArea == nil {
		area := Area{Name: name, Safe: safe, Tiles: tiles, Transports: nil, DefaultTileColor: defaultTileColor}
		space.Areas = append(space.Areas, area)
	} else {
		if new {
			io.WriteString(w, `<h2>Invalid Name</h2>`)
			return
		}
		selectedArea.Safe = safe
		selectedArea.Tiles = tiles
		selectedArea.DefaultTileColor = defaultTileColor
	}

	outFile := c.collectionPath + collectionName + "/spaces/" + spaceName + "2.json"
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

////////////////////////////////////////////////////////////////////////
//  Painting

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

func (c Context) clickOnSquare(w http.ResponseWriter, r *http.Request) {
	y, x, success := dataFromRequest(r)
	if !success {
		panic("No Coordinates provided")
	}
	properties, _ := requestToProperties(r)
	selectedTool, ok := properties["radio-tool"]
	if !ok {
		panic("No Tool Selected")
	}
	selectedMaterialId, err := strconv.Atoi(properties["selected-material"])
	if err != nil {
		fmt.Println("No Material Selected.")
		selectedMaterialId = 0
	}
	selectedMaterial := c.materials[selectedMaterialId]
	defaultTileColor := properties["defaultTileColor"]

	if selectedTool == "select" {
		io.WriteString(w, selectSquare(y, x, defaultTileColor))
	} else if selectedTool == "replace" {
		io.WriteString(w, replaceSquare(y, x, selectedMaterial, defaultTileColor))
	} else if selectedTool == "fill" {
		io.WriteString(w, fillFrom(y, x, selectedMaterial, defaultTileColor))
	} else if selectedTool == "between" {
		io.WriteString(w, fillBetween(y, x, selectedMaterial, defaultTileColor))
	}
}

func dataFromRequest(r *http.Request) (int, int, bool) {
	yCoord, _ := strconv.Atoi(r.Header["Y"][0])
	xCoord, _ := strconv.Atoi(r.Header["X"][0])

	return yCoord, xCoord, true
}

func selectSquare(y, x int, defaultTileColor string) string {
	output := ""
	if haveSelection {
		output += oobSquareUnselected(selectedY, selectedX, modifications[selectedY][selectedX], defaultTileColor)
	}
	haveSelection = true // Probably should be a hidden input
	selectedY = y
	selectedX = x
	return output + oobSquareSelected(y, x, modifications[y][x], defaultTileColor)
}

func replaceSquare(y int, x int, selectedMaterial Material, defaultTileColor string) string {
	modifications[y][x] = selectedMaterial
	return oobSquareUnselected(y, x, selectedMaterial, defaultTileColor)
}

func oobSquare(y int, x int, material Material, defaultColor string, selected bool, oob bool) string {
	redBox := ""
	if selected {
		redBox = `<div class="box top red-b med"></div>`
	}
	overlay := fmt.Sprintf(`<div class="box floor1 %s"></div><div class="box floor2 %s"></div><div class="box ceiling1 %s"></div><div class="box ceiling2 %s"></div>%s`, material.Floor1Css, material.Floor2Css, material.Ceiling1Css, material.Ceiling2Css, redBox)
	var yStr = strconv.Itoa(y)
	var xStr = strconv.Itoa(x)
	tileColor := material.CssColor
	if tileColor == "" {
		tileColor = defaultColor
	}
	oobString := ""
	if oob {
		oobString = `hx-swap-oob="true"`
	}
	return fmt.Sprintf(`<div hx-post="/clickOnSquare" hx-trigger="click" hx-include="[name='radio-tool'],[name='selected-material'],[name='defaultTileColor']" hx-headers='{"y": "%s", "x": "%s"}' class="grid-square %s" id="c%s-%s" %s>%s</div>`, yStr, xStr, tileColor, yStr, xStr, oobString, overlay)
}

func oobSquareSelected(y int, x int, material Material, defaultColor string) string {
	return oobSquare(y, x, material, defaultColor, true, true)
}

func oobSquareUnselected(y int, x int, material Material, defaultColor string) string {
	return oobSquare(y, x, material, defaultColor, false, true)
}

func squareUnselected(y int, x int, material Material, defaultColor string) string {
	return oobSquare(y, x, material, defaultColor, false, false)
}

func exampleSquareFromMaterial(material Material) string {
	overlay := fmt.Sprintf(`<div class="box floor1 %s"></div><div class="box floor2 %s"></div><div class="box ceiling1 %s"></div><div class="box ceiling2 %s"></div>`, material.Floor1Css, material.Floor2Css, material.Ceiling1Css, material.Ceiling2Css)
	idHiddenInput := fmt.Sprintf(`<input name="selected-material" type="hidden" value="%d" />`, material.ID)
	return fmt.Sprintf(`<div class="grid-square %s" name="selected-material">%s%s</div>`, material.CssColor, overlay, idHiddenInput)
}

func fillFrom(y int, x int, selectedMaterial Material, defaultTileColor string) string {
	targetId := modifications[y][x].ID
	seen := make([][]bool, len(modifications))
	for row := range seen {
		seen[row] = make([]bool, len(modifications[row]))
	}
	fillModifications(y, x, targetId, selectedMaterial, seen, defaultTileColor)
	return getHTMLFromModifications(defaultTileColor)
}

func getHTMLFromModifications(defaultTileColor string) string {
	output := `<div class="grid" id="screen" hx-swap-oob="true">`
	for y := range modifications {
		output += `<div class="grid-row">`
		for x := range modifications[y] {
			output += squareUnselected(y, x, modifications[y][x], defaultTileColor)
		}
		output += `</div>`
	}
	output += `</div>`
	return output
}

func fillModifications(y int, x int, targetId int, selected Material, seen [][]bool, defaultTileColor string) {
	seen[y][x] = true
	modifications[y][x] = selected
	deltas := []int{-1, 1}
	for _, i := range deltas {
		if y+i >= 0 && y+i < len(modifications) {
			shouldfill := !seen[y+i][x] && modifications[y+i][x].ID == targetId
			if shouldfill {
				fillModifications(y+i, x, targetId, selected, seen, defaultTileColor)
			}
		}
		if x+i >= 0 && x+i < len(modifications[y]) {
			shouldfill := !seen[y][x+i] && modifications[y][x+i].ID == targetId
			if shouldfill {
				fillModifications(y, x+i, targetId, selected, seen, defaultTileColor)
			}
		}
	}
}

func fillBetween(y int, x int, selectedMaterial Material, defaultTileColor string) string {
	if !haveSelection {
		selectSquare(y, x, defaultTileColor)
	}
	var lowx, lowy, highx, highy int
	if y <= selectedY {
		lowy = y
		highy = selectedY
	} else {
		lowy = selectedY
		highy = y
	}
	if x <= selectedX {
		lowx = x
		highx = selectedX
	} else {
		lowx = selectedX
		highx = x
	}
	output := ""
	for i := lowy; i <= highy; i++ {
		for j := lowx; j <= highx; j++ {
			output += replaceSquare(i, j, selectedMaterial, defaultTileColor)
		}
	}
	output += selectSquare(y, x, defaultTileColor)
	return output
}
