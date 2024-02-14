package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
)

type Transport struct {
	SourceY   int    `json:"sourceY"`
	SourceX   int    `json:"sourceX"`
	DestY     int    `json:"destY"`
	DestX     int    `json:"destX"`
	DestStage string `json:"destStage"`
}

type Area struct {
	Name       string      `json:"name"`
	Safe       bool        `json:"safe"`
	Tiles      [][]int     `json:"tiles"`
	Transports []Transport `json:"transports"`
}

var haveSelection bool = false
var selectedX int
var selectedY int
var modifications [][]Material

func saveArea(w http.ResponseWriter, r *http.Request) {
	properties, _ := requestToProperties(r)
	name := properties["areaName"]
	safe := (properties["safe"] == "on")

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

	areas = append(areas, area)

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

	return
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
		<select name="area-name" hx-get="/edit" hx-target="#panel">
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

	var selectedArea Area
	for _, area := range areas {
		if name == area.Name {
			selectedArea = area
		}
	}

	output := getHTMLFromArea(selectedArea)
	io.WriteString(w, output)
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
			modifications[i][j] = Material{ID: 0, CommonName: "default", CssClassName: "", Walkable: true, R: 255, G: 255, B: 255}
		}
	}

	output := ``
	output += divCreateArea()
	output += divSaveArea()
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

func divSaveArea() string {
	return `		
	<div id="saveForm">
		<form hx-post="/saveArea">
			<label>Name:</label>
			<input type="text" name="areaName">
			<label>Safe:</label>
			<input type="checkbox" name="safe">
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
			output += `<div hx-post="/clickOnSquare" hx-trigger="click" hx-include="[name='radio-tool'],[name='selected-material']" hx-headers='{"y": "` + yStr + `", "x": "` + xStr + `"}' class="grid-square ` + materials[area.Tiles[y][x]].CssClassName + `" id="c` + yStr + `-` + xStr + `"></div>`
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

	io.WriteString(w, fmt.Sprintf(`<div class="grid-square %s"><input name="selected-material" type="hidden" value="%d" /></div>`, selectedMaterial.CssClassName, selectedMaterial.ID))
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
		output += `<div hx-post="/clickOnSquare" hx-swap-oob="true" hx-trigger="click" hx-include="[name='radio-tool'],[name='selected-material']" hx-headers='{"y": "` + yStr + `", "x": "` + xStr + `"}' class="grid-square ` + modifications[selectedY][selectedX].CssClassName + `" id="c` + yStr + `-` + xStr + `"></div>`
	}
	haveSelection = true // Probably should be a hidden input
	selectedY = y
	selectedX = x
	var yStr = strconv.Itoa(y)
	var xStr = strconv.Itoa(x)
	return output + `<div hx-post="/clickOnSquare" hx-swap-oob="true" hx-trigger="click" hx-include="[name='radio-tool'],[name='selected-material']" hx-headers='{"y": "` + yStr + `", "x": "` + xStr + `"}' class="grid-square selected ` + modifications[y][x].CssClassName + `" id="c` + yStr + `-` + xStr + `"></div>`
}

func replaceSquare(y int, x int, selectedMaterial Material) string {
	modifications[y][x] = selectedMaterial

	var yStr = strconv.Itoa(y)
	var xStr = strconv.Itoa(x)
	return fmt.Sprintf(`<div hx-post="/clickOnSquare" hx-swap-oob="true" hx-trigger="click" hx-include="[name='radio-tool'],[name='selected-material']" hx-headers='{"y": "%s", "x": "%s"}' class="grid-square %s" id="c%s-%s"></div>`, yStr, xStr, selectedMaterial.CssClassName, yStr, xStr)
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

	cells := fmt.Sprintf(`<div hx-swap-oob="true" hx-post="/clickOnSquare" hx-trigger="click" hx-include="[name='radio-tool'],[name='selected-material']" hx-headers='{"y": "%s", "x": "%s"}' class="grid-square %s" id="c%s-%s"></div>`, yStr, xStr, selected.CssClassName, yStr, xStr)
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
