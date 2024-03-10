package main

import (
	"fmt"
	"html/template"
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

type Transport struct {
	SourceY   int    `json:"sourceY"`
	SourceX   int    `json:"sourceX"`
	DestY     int    `json:"destY"`
	DestX     int    `json:"destX"`
	DestStage string `json:"destStage"`
}

var haveSelection bool = false
var selectedX int
var selectedY int
var modifications [][]Material // This is probably going to be a problem. At minimum implies one user max

var divArea = `
	<div id="area-page">
		<input type="hidden" name="currentCollection" value="{{.CollectionName}}" />
		<input type="hidden" name="currentSpace" value="{{.Name}}" />
		<div id="select-area">
			<label>Areas</label>
			<select name="area-name" hx-get="/edit" hx-include="[name='currentSpace'],[name='currentCollection']" hx-target="#edit-area">
				<option value="">--</option>
				{{range  $i, $area := .Areas}}
					<option value="{{$area.Name}}">{{$area.Name}}</option>
				{{end}}
			</select>
		</div>
		<div id="edit-area">
		
		</div>
	</div>
`

var areaTmpl = template.Must(template.New("AreaSelect").Parse(divArea))

func (c Context) areasHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		queryValues := r.URL.Query()
		collectionName := queryValues.Get("currentCollection")
		spaceName := queryValues.Get("spaceName")
		s := c.getSpace(collectionName, spaceName)
		getAreaPage(w, r, s)
	}
}

func getAreaPage(w http.ResponseWriter, r *http.Request, space *Space) {
	err := areaTmpl.Execute(w, *space)
	if err != nil {
		fmt.Println(err)
	}
}

func (c Context) saveArea(w http.ResponseWriter, r *http.Request) {
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

func (c Context) getEditArea(w http.ResponseWriter, r *http.Request) {
	queryValues := r.URL.Query()
	collectionName := queryValues.Get("currentCollection")
	spaceName := queryValues.Get("currentSpace")
	name := queryValues.Get("area-name")

	space := c.getSpace(collectionName, spaceName)

	c.editArea(w, r, space, name)
}

func (c Context) editFromTransport(w http.ResponseWriter, r *http.Request) {
	queryValues := r.URL.Query()

	collectionName := queryValues.Get("currentCollection")
	spaceName := queryValues.Get("currentSpace")
	space := c.getSpace(collectionName, spaceName)

	name := queryValues.Get("north_input")
	if name != "" {
		c.editArea(w, r, space, name)
		return
	}
	name = queryValues.Get("south_input")
	if name != "" {
		c.editArea(w, r, space, name)
		return
	}
	name = queryValues.Get("east_input")
	if name != "" {
		c.editArea(w, r, space, name)
		return
	}
	name = queryValues.Get("west_input")
	if name != "" {
		c.editArea(w, r, space, name)
		return
	}

	io.WriteString(w, "<h2>invalid</h2>")

}

func (c Context) editArea(w http.ResponseWriter, r *http.Request, space *Space, name string) {
	selectedArea := getAreaByName(space.Areas, name)
	if selectedArea == nil {
		io.WriteString(w, "<h2>no Area</h2>")
		return
	}

	modifications = c.AreaToMaterialGrid(*selectedArea)

	output := divSaveArea(*selectedArea, space.CollectionName, space.Name)
	output += `<div id="edit_window" class="side">
					<div id="edit_material" class="left">`
	output += c.getHTMLFromArea(*selectedArea) + divToolSelect() + c.divMaterialSelect()

	output += `		</div>
					<div id="edit_sidebar" class="left">
						<div id="edit_options">
							<a hx-get="/editTransports" hx-target="#edit_tool" hx-include="[name='areaName'],[name='currentCollection'],[name='currentSpace']" href="#">Edit Transports</a> | 
							<a hx-get="/editDisplay" hx-target="#edit_tool" href="#">Edit Display</a> | 
							<a hx-get="/getEditNeighbors" hx-target="#edit_tool" hx-include="[name='areaName'],[name='currentCollection'],[name='currentSpace']" href="#">Edit Neighbors</a> |
							<a hx-get="/materialPage"  hx-target="#edit_tool" href="#">Edit Colors/Materials</a>
						</div>
						<div id="edit_tool">
						
						</div>
					</div>
				</div>`
	io.WriteString(w, output)
}

func editDisplay(w http.ResponseWriter, r *http.Request) {
	output := `<div id="edit_display">
					<h3>Select BgColor</h3>
					<h3>Show/Hide Grid-lines</h3>
					<h3>Show/Hide Transports</h3>
					<h3>Show/Hide Ceiling</h3>
				</div>`
	io.WriteString(w, output)
}

func (c Context) getEditNeighbors(w http.ResponseWriter, r *http.Request) {
	queryValues := r.URL.Query()
	name := queryValues.Get("areaName")

	collectionName := queryValues.Get("currentCollection")
	spaceName := queryValues.Get("currentSpace")
	space := c.getSpace(collectionName, spaceName)
	selectedArea := getAreaByName(space.Areas, name)
	if selectedArea == nil {
		io.WriteString(w, "<h2>no Area</h2>")
		return
	}

	io.WriteString(w, divEditNeighborsForArea(*selectedArea))
}

// template
func divEditNeighborsForArea(selectedArea Area) string {
	output := `<div id="edit_neighbors">
					<div id="edit_north">
						<h3>North: </h3>
						<input type="text" name="north_input" value="` + selectedArea.North + `"/>
						<a hx-get="/editFromTransport" hx-include="[name='north_input'],[name='currentCollection'],[name='currentSpace']" hx-target="#edit-area" hx href="#">Go</a>
					</div>
					<div id="edit_south">
						<h3>South: </h3>
						<input type="text" name="south_input" value="` + selectedArea.South + `"/>
						<a hx-get="/editFromTransport" hx-include="[name='south_input'],[name='currentCollection'],[name='currentSpace']" hx-target="#edit-area" href="#">Go</a>
					</div>
					<div id="edit_east">
						<h3>East: </h3>
						<input type="text" name="east_input" value="` + selectedArea.East + `"/>
						<a hx-get="/editFromTransport" hx-include="[name='east_input'],[name='currentCollection'],[name='currentSpace']" hx-target="#edit-area" href="#">Go</a>
					</div>
					<div id="edit_west">
						<h3>West: </h3>
						<input type="text" name="west_input" value="` + selectedArea.West + `"/>
						<a hx-get="/editFromTransport" hx-include="[name='west_input'],[name='currentCollection'],[name='currentSpace']" hx-target="#edit-area" href="#">Go</a>
					</div>
					<a hx-post="/editNeighbors" hx-include="[name='areaName'],[name='north_input'],[name='south_input'],[name='east_input'],[name='west_input'],[name='currentCollection'],[name='currentSpace']" hx-target="#edit_tool" href="#">Save</a>
				</div>`
	return output
}

func (c Context) editNeighbors(w http.ResponseWriter, r *http.Request) {
	properties, _ := requestToProperties(r)
	name := properties["areaName"]
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

	io.WriteString(w, note+divEditNeighborsForArea(*selectedArea))
}

func getAreaByName(areas []Area, name string) *Area {
	for i, area := range areas {
		if name == area.Name {
			return &areas[i]
		}
	}
	return nil
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

// Have default tile color change trigger getHtmlFromModifications()
func divSaveArea(area Area, currentCollection string, currentSpace string) string {
	checked := ""
	if area.Safe {
		checked = "checked"
	}
	return `		
	<div id="saveForm">
		<div id="save_notice"></div>
		<form hx-post="/saveArea" hx-include hx-target="#save_notice">
			<input type="hidden" name="currentCollection" value="` + currentCollection + `"/>
			<input type="hidden" name="currentSpace" value="` + currentSpace + `"/>
			<input type="hidden" name="new" value="false"/>
			<label>Name:</label>
			<input type="text" name="areaName" value="` + area.Name + `">
			<label>Safe:</label>
			<input type="checkbox" name="safe" ` + checked + `>
			<button>Save</button><br />
			<label>Default Tile Color:</label>
			<input type="text" name="defaultTileColor" value="` + area.DefaultTileColor + `">
		</form>
	</div>`
}

func divToolSelect() string {
	return `
	<div id="tool_select"> 
		<input type="radio" id="rt-select" name="radio-tool" value="select" checked>
		<label for="rt-select">Select</label>
		<input type="radio" id="rt-replace" name="radio-tool" value="replace">
		<label for="rt-replace">Replace</label>
		<input type="radio" id="rt-fill" name="radio-tool" value="fill">
		<label for="rt-fill">Fill</label>
		<input type="radio" id="rt-right" name="radio-tool" value="between">
		<label for="rt-right">Fill between selected</label>
	</div>
	`
}

func (c Context) getHTMLFromArea(area Area) string {
	output := `<div class="grid" id="screen">`
	for y := range area.Tiles {
		output += `<div class="grid-row">`
		for x := range area.Tiles[y] {
			materialId := area.Tiles[y][x]
			output += squareUnselected(y, x, c.materials[materialId], area.DefaultTileColor)
		}
		output += `</div>`
	}
	output += `</div>`
	return output
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

func (c Context) divMaterialSelect() string {
	output := `
	<div id="material-selector">
		<label>Materials</label>
		<select name="materialId" hx-get="/selectMaterial" hx-target="#selected-material-div">
			<option value="">--</option>	
	`
	for _, material := range c.materials {
		output += fmt.Sprintf(`<option value="%d">%s</option>`, material.ID, material.CommonName)
	}

	output += `
		</select>
	</div>
	<div id="selected-material-div" class="grid-row"></div>`

	return output
}

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
