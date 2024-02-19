package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
)

type Area struct {
	Name       string      `json:"name"`
	Safe       bool        `json:"safe"`
	Tiles      [][]int     `json:"tiles"`
	Transports []Transport `json:"transports"`
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
var currentTransports []Transport

func saveArea(w http.ResponseWriter, r *http.Request) {
	properties, _ := requestToProperties(r)
	name := properties["areaName"]
	safe := (properties["safe"] == "on")
	new := (properties["new"] == "true")

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

	area := Area{Name: name, Safe: safe, Tiles: tiles, Transports: nil}

	// This will delete any Transports
	index := getIndexOfAreaByName(name)
	if index < 0 {
		areas = append(areas, area)
	} else {
		if new {
			io.WriteString(w, `<h2>Invalid Name</h2>`)
			return
		}
		area.Transports = currentTransports
		areas[index] = area
	}

	data, err := json.Marshal(areas)
	if err != nil {
		return
	}

	file, err := os.Create("./level/data/areas.json")
	if err != nil {
		return
	}
	defer file.Close()

	_, err = file.Write(data)
	if err != nil {
		return
	}

	io.WriteString(w, `<h2>Sucess</h2>`)
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
				<input type="text" id="height" name="height" value="10">
			</div>
			<div>
				<label for="width">Width:</label>
				<input type="text" id="width" name="width" value="14">
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
		<labelAreas</label>
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

	index := getIndexOfAreaByName(name)
	if index < 0 {
		io.WriteString(w, "<h2>Invalid Area</h2>")
		return
	}
	selectedArea := areas[index]

	modifications = AreaToMaterialGrid(selectedArea)
	currentTransports = selectedArea.Transports
	output := divSaveArea(selectedArea)
	output += `<div id="edit_window" class="side">
					<div id="edit_material" class="left">`
	output += getHTMLFromArea(selectedArea) + divToolSelect() + divMaterialSelect()

	output += `		</div>
					<div id="edit_sidebar" class="left">
						<div id="edit_options">
							<a hx-get="/editTransports" hx-target="#edit_tool" href="#">Edit Transports</a> | 
							<a hx-get="/editDisplay" hx-target="#edit_tool" href="#">Edit Display</a> | 
							<a hx-get="/editNeighbors" hx-target="#edit_tool" href="#">Edit Neighbors</a>
						</div>
						<div id="edit_tool">
						
						</div>
					</div>
				</div>`
	io.WriteString(w, output)
}

func editTransports(w http.ResponseWriter, r *http.Request) {
	output := transportFormHtml(currentTransports)
	output += transportsAsOob(currentTransports)
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

	currentTransport := &currentTransports[transportId]
	currentTransport.DestY = destY
	currentTransport.DestX = destX
	currentTransport.SourceY = sourceY
	currentTransport.SourceX = sourceX
	currentTransport.DestStage = destStage

	output := transportFormHtml(currentTransports)
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

func editNeighbors(w http.ResponseWriter, r *http.Request) {
	output := `<div id="edit_neighbors">
					<h3>North: </h3>
					<h3>South: </h3>
					<h3>East: </h3>
					<h3>West: </h3>
				</div>`
	io.WriteString(w, output)
}

func transportFormHtml(transports []Transport) string {
	output := `<div id="edit_transports">
					<h4>Transports: </h4>`
	for i := range transports {
		output += editTransportForm(i, transports[i])
	}
	output += `</div>`
	return output
}

func transportsAsOob(transports []Transport) string {
	output := ``
	for _, transport := range transports {
		var yStr = strconv.Itoa(transport.SourceY)
		var xStr = strconv.Itoa(transport.SourceX)
		output += `<div hx-swap-oob="true" hx-post="/clickOnSquare" hx-trigger="click" hx-include="[name='radio-tool'],[name='selected-material']" hx-headers='{"y": "` + yStr + `", "x": "` + xStr + `"}' class="grid-square ` + modifications[transport.SourceY][transport.SourceX].CssColor + `" id="c` + yStr + `-` + xStr + `"><div class="box0 med red-b"></div></div></div>`
	}
	output += ``
	return output
}

func editTransportForm(i int, t Transport) string {
	output := fmt.Sprintf(`
	<form hx-post="/editTransport" hx-target="#edit_transports" hx-swap="outerHTML">
		<input type="hidden" name="transport-id" value="%d" />
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
		<button class="btn" hx-post="/duplicateTransport">Duplicate</button>
		<button class="btn" hx-post="/deleteTransport">Delete</button>
	</form>`, i, t.DestStage, t.DestY, t.DestX, t.SourceY, t.SourceX, "pink")
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
			modifications[i][j] = Material{ID: 0, CommonName: "default", CssColor: "", Walkable: true, Layer1Css: "", Layer2Css: ""}
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
			<button>Save</button>
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
			output += `<div hx-post="/clickOnSquare" hx-trigger="click" hx-include="[name='radio-tool'],[name='selected-material']" hx-headers='{"y": "` + yStr + `", "x": "` + xStr + `"}' class="grid-square" id="c` + yStr + `-` + xStr + `"></div>`
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
			var yStr = strconv.Itoa(y)
			var xStr = strconv.Itoa(x)
			output += `<div hx-post="/clickOnSquare" hx-trigger="click" hx-include="[name='radio-tool'],[name='selected-material']" hx-headers='{"y": "` + yStr + `", "x": "` + xStr + `"}' class="grid-square ` + materials[area.Tiles[y][x]].CssColor + `" id="c` + yStr + `-` + xStr + `"></div>`
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

	io.WriteString(w, fmt.Sprintf(`<div class="grid-square %s"><input name="selected-material" type="hidden" value="%d" /></div>`, selectedMaterial.CssColor, selectedMaterial.ID))
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

	if selectedTool == "select" {
		io.WriteString(w, selectSquare(y, x))
	} else if selectedTool == "replace" {
		io.WriteString(w, replaceSquare(y, x, selectedMaterial))
	} else if selectedTool == "fill" {
		io.WriteString(w, fillFrom(y, x, selectedMaterial))
	} else if selectedTool == "between" {
		io.WriteString(w, fillBetween(y, x, selectedMaterial))
	}
}

func dataFromRequest(r *http.Request) (int, int, bool) {
	yCoord, _ := strconv.Atoi(r.Header["Y"][0])
	xCoord, _ := strconv.Atoi(r.Header["X"][0])

	return yCoord, xCoord, true
}

func selectSquare(y, x int) string {
	output := ""
	if haveSelection {
		var yStr = strconv.Itoa(selectedY)
		var xStr = strconv.Itoa(selectedX)
		output += `<div hx-post="/clickOnSquare" hx-swap-oob="true" hx-trigger="click" hx-include="[name='radio-tool'],[name='selected-material']" hx-headers='{"y": "` + yStr + `", "x": "` + xStr + `"}' class="grid-square ` + modifications[selectedY][selectedX].CssColor + `" id="c` + yStr + `-` + xStr + `"></div>`
	}
	haveSelection = true // Probably should be a hidden input
	selectedY = y
	selectedX = x
	var yStr = strconv.Itoa(y)
	var xStr = strconv.Itoa(x)
	return output + `<div hx-post="/clickOnSquare" hx-swap-oob="true" hx-trigger="click" hx-include="[name='radio-tool'],[name='selected-material']" hx-headers='{"y": "` + yStr + `", "x": "` + xStr + `"}' class="grid-square ` + modifications[y][x].CssColor + `" id="c` + yStr + `-` + xStr + `"><div class="box0 med red-b" /></div>`
}

func replaceSquare(y int, x int, selectedMaterial Material) string {
	modifications[y][x] = selectedMaterial

	var yStr = strconv.Itoa(y)
	var xStr = strconv.Itoa(x)
	return fmt.Sprintf(`<div hx-post="/clickOnSquare" hx-swap-oob="true" hx-trigger="click" hx-include="[name='radio-tool'],[name='selected-material']" hx-headers='{"y": "%s", "x": "%s"}' class="grid-square %s" id="c%s-%s"></div>`, yStr, xStr, selectedMaterial.CssColor, yStr, xStr)
}

func fillFrom(y int, x int, selectedMaterial Material) string {
	targetId := modifications[y][x].ID
	seen := make([][]bool, len(modifications))
	for row := range seen {
		seen[row] = make([]bool, len(modifications[row]))
	}
	return fillAndCheckNeighbors(y, x, targetId, selectedMaterial, seen)
}

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

func fillBetween(y int, x int, selectedMaterial Material) string {
	if !haveSelection {
		selectSquare(y, x)
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
			output += replaceSquare(i, j, selectedMaterial)
		}
	}
	output += selectSquare(y, x)
	return output
}
