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
var modifications [][]Material

//var currentTransports []Transport // This should only exist if its showing the highlights and not even then

func saveArea(w http.ResponseWriter, r *http.Request) {
	properties, _ := requestToProperties(r)
	name := properties["areaName"]
	safe := (properties["safe"] == "on")
	new := (properties["new"] == "true")
	defaultTileColor := properties["defaultTileColor"]

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

	index := getIndexOfAreaByName(name)
	if index < 0 {
		area := Area{Name: name, Safe: safe, Tiles: tiles, Transports: nil, DefaultTileColor: defaultTileColor}
		areas = append(areas, area)
	} else {
		if new {
			io.WriteString(w, `<h2>Invalid Name</h2>`)
			return
		}
		areas[index].Safe = safe
		areas[index].Tiles = tiles
		areas[index].DefaultTileColor = defaultTileColor
	}

	attempt := writeJsonFile(areaPath, areas)
	if attempt != nil {
		panic("Area Write Failure")
	}

	io.WriteString(w, `<h2>Success</h2>`)
}

func getCreateArea(w http.ResponseWriter, r *http.Request) {
	output := divCreateArea()
	io.WriteString(w, output)
}

func divCreateArea() string {
	return `	
	<div id="create_form">
		<form hx-post="/createGrid" hx-target="#panel">
			<div>
				<label>Input dimensions:</label>
			</div>
			<div>
				<label for="height">Height:</label>
				<input type="text" id="height" name="height" value="16">
			</div>
			<div>
				<label for="width">Width:</label>
				<input type="text" id="width" name="width" value="16">
			</div>
			<div>
				<button>Create</button>
			</div>
		</form>
	</div>`
}

func getEditAreaPage(w http.ResponseWriter, r *http.Request) {
	output := `
	<div>
		<label>Areas</label>
		<select name="area-name" hx-get="/edit" hx-target="#edit-area">
			<option value="">--</option>
	`
	for _, area := range areas {
		output += fmt.Sprintf(`<option value="%s">%s</option>`, area.Name, area.Name)
	}
	output += `</select>
	</div>
	<div id="edit-area">
	
	</div>`
	io.WriteString(w, output)
}

func edit(w http.ResponseWriter, r *http.Request) {
	queryValues := r.URL.Query()
	name := queryValues.Get("area-name")
	editByName(w, r, name)
}

func editFromTransport(w http.ResponseWriter, r *http.Request) {
	queryValues := r.URL.Query()
	name := queryValues.Get("north_input")
	if name != "" {
		editByName(w, r, name)
		return
	}
	name = queryValues.Get("south_input")
	if name != "" {
		editByName(w, r, name)
		return
	}
	name = queryValues.Get("east_input")
	if name != "" {
		editByName(w, r, name)
		return
	}
	name = queryValues.Get("west_input")
	if name != "" {
		editByName(w, r, name)
		return
	}

	io.WriteString(w, "<h2>invalid</h2>")

}

func editByName(w http.ResponseWriter, r *http.Request, name string) {
	index := getIndexOfAreaByName(name)
	if index < 0 {
		io.WriteString(w, "<h2>Invalid Area</h2>")
		return
	}
	selectedArea := areas[index]

	modifications = AreaToMaterialGrid(selectedArea)
	//currentTransports = selectedArea.Transports
	output := divSaveArea(selectedArea)
	output += `<div id="edit_window" class="side">
					<div id="edit_material" class="left">`
	output += getHTMLFromArea(selectedArea) + divToolSelect() + divMaterialSelect()

	output += `		</div>
					<div id="edit_sidebar" class="left">
						<div id="edit_options">
							<a hx-get="/editTransports" hx-target="#edit_tool" hx-include="[name='areaName']" href="#">Edit Transports</a> | 
							<a hx-get="/editDisplay" hx-target="#edit_tool" href="#">Edit Display</a> | 
							<a hx-get="/getEditNeighbors" hx-target="#edit_tool" hx-include="[name='areaName']" href="#">Edit Neighbors</a> |
							<a hx-get="/materialPage"  hx-target="#edit_tool" href="#">Edit Colors/Materials</a>
						</div>
						<div id="edit_tool">
						
						</div>
					</div>
				</div>`
	io.WriteString(w, output)
}

func getEditTransports(w http.ResponseWriter, r *http.Request) {
	queryValues := r.URL.Query()
	name := queryValues.Get("areaName")
	output := transportFormHtml(name)
	output += transportsAsOob(name)
	io.WriteString(w, output)
}

func editTransport(w http.ResponseWriter, r *http.Request) {
	properties, _ := requestToProperties(r)
	transportId, _ := strconv.Atoi(properties["transport-id"])
	destStage := properties["transport-stage-name"]
	destY, _ := strconv.Atoi(properties["transport-dest-y"])
	destX, _ := strconv.Atoi(properties["transport-dest-x"])
	sourceY, _ := strconv.Atoi(properties["transport-source-y"])
	sourceX, _ := strconv.Atoi(properties["transport-source-x"])
	areaName := properties["transport-area-name"]

	index := getIndexOfAreaByName(areaName)
	if index < 0 {
		io.WriteString(w, "<h2>Invalid Area</h2>")
		return
	}
	selectedArea := areas[index]

	currentTransport := &selectedArea.Transports[transportId]
	currentTransport.DestY = destY
	currentTransport.DestX = destX
	currentTransport.SourceY = sourceY
	currentTransport.SourceX = sourceX
	currentTransport.DestStage = destStage

	output := transportFormHtml(areaName)
	io.WriteString(w, output)
}

func newTransport(w http.ResponseWriter, r *http.Request) {
	properties, _ := requestToProperties(r)
	areaName := properties["areaName"]
	fmt.Println(areaName)

	index := getIndexOfAreaByName(areaName)
	if index < 0 {
		io.WriteString(w, "<h2>Invalid Area</h2>")
		return
	}
	selectedArea := &areas[index]

	selectedArea.Transports = append(selectedArea.Transports, Transport{})

	output := transportFormHtml(areaName)
	io.WriteString(w, output)

}

func transportFormHtml(areaName string) string {
	index := getIndexOfAreaByName(areaName)
	if index < 0 {
		return "<h2>Invalid Area</h2>"
	}
	selectedArea := areas[index]

	output := `<div id="edit_transports">
					<h4>Transports: </h4>
					<a hx-post="/newTransport" hx-include="[name='areaName']" hx-target="#edit_transports" href="#"> New </a><br />`
	for i, t := range selectedArea.Transports {
		output += editTransportForm(i, t, areaName)
	}
	output += `</div>`
	return output
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

func getEditNeighbors(w http.ResponseWriter, r *http.Request) {
	queryValues := r.URL.Query()
	name := queryValues.Get("areaName")

	index := getIndexOfAreaByName(name)
	if index < 0 {
		io.WriteString(w, "<h2>Invalid Area</h2>")
		return
	}
	selectedArea := areas[index]

	io.WriteString(w, divEditNeighborsForArea(selectedArea))
}

func divEditNeighborsForArea(selectedArea Area) string {
	output := `<div id="edit_neighbors">
					<div id="edit_north">
						<h3>North: </h3>
						<input type="text" name="north_input" value="` + selectedArea.North + `"/>
						<a hx-get="/editFromTransport" hx-include="[name='north_input']" hx-target="#edit-area" hx href="#">Go</a>
					</div>
					<div id="edit_south">
						<h3>South: </h3>
						<input type="text" name="south_input" value="` + selectedArea.South + `"/>
						<a hx-get="/editFromTransport" hx-include="[name='south_input']" hx-target="#edit-area" href="#">Go</a>
					</div>
					<div id="edit_east">
						<h3>East: </h3>
						<input type="text" name="east_input" value="` + selectedArea.East + `"/>
						<a hx-get="/editFromTransport" hx-include="[name='east_input']" hx-target="#edit-area" href="#">Go</a>
					</div>
					<div id="edit_west">
						<h3>West: </h3>
						<input type="text" name="west_input" value="` + selectedArea.West + `"/>
						<a hx-get="/editFromTransport" hx-include="[name='west_input']" hx-target="#edit-area" href="#">Go</a>
					</div>
					<a hx-post="/editNeighbors" hx-include="[name='areaName'],[name='north_input'],[name='south_input'],[name='east_input'],[name='west_input']" hx-target="#edit_tool" href="#">Save</a>
				</div>`
	return output
}

func editNeighbors(w http.ResponseWriter, r *http.Request) {
	properties, _ := requestToProperties(r)
	name := properties["areaName"]
	north := properties["north_input"]
	south := properties["south_input"]
	east := properties["east_input"]
	west := properties["west_input"]

	index := getIndexOfAreaByName(name)
	if index < 0 {
		io.WriteString(w, "<h2>Invalid Area</h2>")
		return
	}
	selectedArea := &areas[index]
	selectedArea.North = north
	selectedArea.South = south
	selectedArea.East = east
	selectedArea.West = west

	note := `<div id="confirmation_neighbor_change"><p>saved</p></div>`

	io.WriteString(w, note+divEditNeighborsForArea(*selectedArea))
}

func editTransportForm(i int, t Transport, sourceName string) string {
	output := fmt.Sprintf(`
	<form hx-post="/editTransport" hx-target="#edit_transports" hx-swap="outerHTML">
		<input type="hidden" name="transport-id" value="%d" />
		<input type="hidden" name="transport-area-name" value="%s" />
		<table>
			<tr>
				<td align="right">Dest stage-name:</td>
				<td align="left">
					<input type="text" name="transport-stage-name" value="%s" />
				</td>
			</tr>
			<tr>
				<td align="right">Dest y</td>
				<td align="left">
					<input type="text" name="transport-dest-y" value="%d" />
				</td>
				<td align="right">x</td>
				<td align="left">
					<input type="text" name="transport-dest-x" value="%d" />
				</td>
			</tr>
			<tr>
				<td align="right">Source y</td>
				<td align="left">
					<input type="text" name="transport-source-y" value="%d" />
				</td>
				<td align="right">x</td>
				<td align="left">
					<input type="text" name="transport-source-x" value="%d" />
				</td>
			</tr>
			<tr>
				<td align="right">Css-class:</td>
				<td align="left">
					<input type="text" name="transport-css-class" value="%s" />
				</td>
			<tr />
		</table>

		<button class="btn">Submit</button>
		<button class="btn" hx-post="/dupeTransport" hx-include="[name='areaName]">Duplicate</button>
		<button class="btn" hx-post="/deleteTransport" hx-include="[name='areaName]">Delete</button>
	</form>`, i, sourceName, t.DestStage, t.DestY, t.DestX, t.SourceY, t.SourceX, "pink")
	return output
}

func dupeTransport(w http.ResponseWriter, r *http.Request) {
	properties, _ := requestToProperties(r)
	transportId, _ := strconv.Atoi(properties["transport-id"])
	name := properties["transport-area-name"]

	index := getIndexOfAreaByName(name)
	if index < 0 {
		io.WriteString(w, "<h2>Invalid Area</h2>")
		return
	}
	selectedArea := &areas[index]

	currentTransport := &selectedArea.Transports[transportId]
	newTransport := *currentTransport
	selectedArea.Transports = append(selectedArea.Transports, newTransport)

	output := transportFormHtml(name)
	io.WriteString(w, output)
}

func deleteTransport(w http.ResponseWriter, r *http.Request) {
	properties, _ := requestToProperties(r)
	id, _ := strconv.Atoi(properties["transport-id"])
	name := properties["transport-area-name"]

	index := getIndexOfAreaByName(name)
	if index < 0 {
		io.WriteString(w, "<h2>Invalid Area</h2>")
		return
	}
	selectedArea := &areas[index]

	selectedArea.Transports = append(selectedArea.Transports[:id], selectedArea.Transports[id+1:]...)
	fmt.Println(len(selectedArea.Transports))

	output := transportFormHtml(selectedArea.Name)
	// Remove highlight for deleted transport
	io.WriteString(w, output)
}

func transportsAsOob(areaName string) string {
	index := getIndexOfAreaByName(areaName)
	if index < 0 {
		return "<h2>Invalid Area</h2>"
	}
	selectedArea := areas[index]
	output := ``
	for _, transport := range selectedArea.Transports {
		var yStr = strconv.Itoa(transport.SourceY)
		var xStr = strconv.Itoa(transport.SourceX)
		output += `<div hx-swap-oob="true" hx-post="/clickOnSquare" hx-trigger="click" hx-include="[name='radio-tool'],[name='selected-material']" hx-headers='{"y": "` + yStr + `", "x": "` + xStr + `"}' class="grid-square ` + modifications[transport.SourceY][transport.SourceX].CssColor + `" id="c` + yStr + `-` + xStr + `"><div class="box top med red-b"></div></div></div>`
	}
	output += ``
	return output
}

func getIndexOfAreaByName(name string) int {
	out := -1
	for i, area := range areas {
		if name == area.Name {
			out = i
			break
		}
	}
	return out
}

func AreaToMaterialGrid(area Area) [][]Material {
	out := make([][]Material, len(area.Tiles))
	for y := range area.Tiles {
		out[y] = make([]Material, len(area.Tiles[y]))
		for x := range area.Tiles[y] {
			out[y][x] = materials[area.Tiles[y][x]]
		}
	}
	return out
}

func createGrid(w http.ResponseWriter, r *http.Request) {
	height, width, success := getHeightAndWidth(r)
	if !success {
		panic(0)
	}

	modifications = make([][]Material, height)
	for i := range modifications {
		modifications[i] = make([]Material, width)
		for j := range modifications[i] {
			modifications[i][j] = Material{ID: 0, CommonName: "default", CssColor: "", Walkable: true, Floor1Css: "", Floor2Css: ""}
		}
	}

	output := ``
	output += divCreateArea()
	output += divNewSaveArea()
	output += divToolSelect()
	output += getEmptyGridHTML(height, width)
	output += divMaterialSelect()

	io.WriteString(w, output)
}

func getHeightAndWidth(r *http.Request) (int, int, bool) {
	properties, _ := requestToProperties(r)
	height, _ := strconv.Atoi(properties["height"])
	width, _ := strconv.Atoi(properties["width"])

	return height, width, true
}

// Have default tile color change trigger getHtmlFromModifications()
func divSaveArea(area Area) string {
	checked := ""
	if area.Safe {
		checked = "checked"
	}
	return `		
	<div id="saveForm">
		<div id="save_notice"></div>
		<form hx-post="/saveArea" hx-target="#save_notice">
			<label>Name:</label>
			<input type="text" name="areaName" value="` + area.Name + `">
			<label>Safe:</label>
			<input type="checkbox" name="safe" ` + checked + `>
			<input type="hidden" name="new" value="false"/>
			<button>Save</button><br />
			<input type="text" name="defaultTileColor" value="` + area.DefaultTileColor + `">
		</form>
	</div>`
}

func divNewSaveArea() string {
	return `		
	<div id="saveForm">
		<div id="save_notice"></div>
		<form hx-post="/saveArea" hx-target="#save_notice">
			<label>Name:</label>
			<input type="text" name="areaName" value="" />
			<label>Safe:</label>
			<input type="checkbox" name="safe" />
			<input type="hidden" name="new" value="true"/>
			<button>Save</button>
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

func getEmptyGridHTML(h int, w int) string {
	output := `<div class="grid" id="screen">`
	for y := 0; y < h; y++ {
		output += `<div class="grid-row">`
		for x := 0; x < w; x++ {
			var yStr = strconv.Itoa(y)
			var xStr = strconv.Itoa(x)
			output += `<div hx-post="/clickOnSquare" hx-trigger="click" hx-include="[name='radio-tool'],[name='selected-material'],[name='defaultTileColor']" hx-headers='{"y": "` + yStr + `", "x": "` + xStr + `"}' class="grid-square" id="c` + yStr + `-` + xStr + `"></div>`
		}
		output += `</div>`
	}
	output += `</div>`
	return output
}

func getHTMLFromArea(area Area) string {
	output := `<div class="grid" id="screen">`
	for y := range area.Tiles {
		output += `<div class="grid-row">`
		for x := range area.Tiles[y] {
			materialId := area.Tiles[y][x]
			output += squareUnselected(y, x, materials[materialId], area.DefaultTileColor)
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

func divMaterialSelect() string {
	output := `
	<div id="material-selector">
		<label>Materials</label>
		<select name="materialId" hx-get="/selectMaterial" hx-target="#selected-material-div">
			<option value="">--</option>	
	`
	for _, material := range materials {
		output += fmt.Sprintf(`<option value="%d">%s</option>`, material.ID, material.CommonName)
	}

	output += `
		</select>
	</div>
	<div id="selected-material-div" class="grid-row"></div>`

	return output
}

func selectMaterial(w http.ResponseWriter, r *http.Request) {
	queryValues := r.URL.Query()
	id := queryValues.Get("materialId")

	var selectedMaterial Material
	for _, material := range materials {
		if id, _ := strconv.Atoi(id); id == material.ID {
			selectedMaterial = material
		}
	}

	io.WriteString(w, exampleSquareFromMaterial(selectedMaterial))
}

func clickOnSquare(w http.ResponseWriter, r *http.Request) {
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
	selectedMaterial := materials[selectedMaterialId]
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

// This could be a good way to test arbitrary number of oobs
func fillAndCheckNeighbors(y int, x int, targetId int, selected Material, seen [][]bool) string {
	seen[y][x] = true
	modifications[y][x] = selected
	var yStr = strconv.Itoa(y)
	var xStr = strconv.Itoa(x)

	cells := fmt.Sprintf(`<div hx-swap-oob="true" hx-post="/clickOnSquare" hx-trigger="click" hx-include="[name='radio-tool'],[name='selected-material']" hx-headers='{"y": "%s", "x": "%s"}' class="grid-square %s" id="c%s-%s"></div>`, yStr, xStr, selected.CssColor, yStr, xStr)
	deltas := []int{-1, 1}
	for _, i := range deltas {
		if y+i >= 0 && y+i < len(modifications) {
			shouldfill := !seen[y+i][x] && modifications[y+i][x].ID == targetId
			if shouldfill {
				cells += fillAndCheckNeighbors(y+i, x, targetId, selected, seen)
			}
		}
		if x+i >= 0 && x+i < len(modifications[y]) {
			shouldfill := !seen[y][x+i] && modifications[y][x+i].ID == targetId
			if shouldfill {
				cells += fillAndCheckNeighbors(y, x+i, targetId, selected, seen)
			}
		}
	}
	return cells
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
