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
	MaterialID int    `json:"materialId"`
	DestY      int    `json:"destY"`
	DestX      int    `json:"destX"`
	DestStage  string `json:"destStage"`
}

type Area struct {
	Name       string      `json:"name"`
	Safe       bool        `json:"safe"`
	Tiles      [][]int     `json:"tiles"`
	Transports []Transport `json:"transports"`
}

var selectedMaterial Material
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

func getAreaPage(w http.ResponseWriter, r *http.Request) {
	output := `	
	<div id="controls">
		<form hx-post="/createGrid" hx-target="#panel" hx-swap="outerHTML">
			<div>
				<label>Enter Height and Width:</label>
				<input type="text" name="height" value="10">
				<input type="text" name="width" value="14">
				<button>Create</button>
			</div>
		</form>
	</div>`
	io.WriteString(w, output)
}

func getEditAreaPage(w http.ResponseWriter, r *http.Request) {
	output := `
	<div>
		<label>Materials</label>
		<select name="materialId">
			<option value="">--</option>
	`
	for _, area := range areas {
		output += fmt.Sprintf(`<option value="%s">%s</option>`, area.Name, area.Name)
	}
	output += "</select></div>"
	io.WriteString(w, output)
}

func getHeightAndWidth(r *http.Request) (int, int, bool) {
	properties, _ := requestToProperties(r)
	height, _ := strconv.Atoi(properties["height"])
	width, _ := strconv.Atoi(properties["width"])

	return height, width, true
}

func getGridHTML(h int, w int) string {
	output := ""
	for y := 0; y < h; y++ {
		output += `<div class="grid-row">`
		for x := 0; x < w; x++ {
			var yStr = strconv.Itoa(y)
			var xStr = strconv.Itoa(x)
			output += `<div hx-post="/replace" hx-trigger="click" hx-headers='{"y": "` + yStr + `", "x": "` + xStr + `"}' class="grid-square" id="c` + yStr + `-` + xStr + `"></div>`
		}
		output += `</div>`
	}
	return output
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

	output := `
    <div id="panel">
        <div id="controls">
            <form hx-post="/createGrid" hx-target="#panel" hx-swap="outerHTML">
                <div>
                <label>Enter Height and Width:</label>
                <input type="text" name="height" value="10">
                <input type="text" name="width" value="10">
                </div>
                <button>Create</button>
            </form>
			<form hx-post="/saveArea">
				<label>Name:</label>
				<input type="text" name="areaName">
				<label>Safe:</label>
				<input type="checkbox" name="safe">
				<button>Save</button>
			</form>
        </div>
        <div class="grid" id="screen">`

	output += getGridHTML(height, width)
	output += `</div>
	<div class="color-selector">
		<label>Materials</label>
		<select name="materialId" hx-get="/select" hx-target="#selectedMaterial">
			<option value="">--</option>	
	`
	for _, material := range materials {
		output += fmt.Sprintf(`<option value="%d">%s</option>`, material.ID, material.CommonName)
	}

	output += `
		</select>
	</div>
	<div id="selectedMaterial" class="grid-row">`

	if &selectedMaterial != nil {
		output += fmt.Sprintf(`<div class="grid-square %s></div>`, selectedMaterial.CssClassName)
	}

	output += `</div></div></div>` // Too many /div?
	io.WriteString(w, output)

}

func dataFromRequest(r *http.Request) (int, int, bool) {
	yCoord, _ := strconv.Atoi(r.Header["Y"][0])
	xCoord, _ := strconv.Atoi(r.Header["X"][0])
	fmt.Printf("%d %d\n", yCoord, xCoord)

	return yCoord, xCoord, true
}

func replaceSquare(w http.ResponseWriter, r *http.Request) {
	y, x, success := dataFromRequest(r)
	if !success {
		panic(0)
	}
	className := ""
	if &selectedMaterial != nil {
		className = selectedMaterial.CssClassName
	}

	modifications[y][x] = selectedMaterial

	var yStr = strconv.Itoa(y)
	var xStr = strconv.Itoa(x)
	div := fmt.Sprintf(`<div hx-post="/replace" hx-trigger="click" hx-include="#selectedColor" hx-headers='{"y": "%s", "x": "%s"}' class="grid-square %s" id="c%s-%s"></div>`, yStr, xStr, className, yStr, xStr)
	io.WriteString(w, div)
}

func selectColor(w http.ResponseWriter, r *http.Request) {
	queryValues := r.URL.Query()
	id := queryValues.Get("materialId")

	fmt.Printf("Received id: %s", id)

	for _, material := range materials {
		if id, _ := strconv.Atoi(id); id == material.ID {
			selectedMaterial = material
		}
	}

	io.WriteString(w, fmt.Sprintf(`<div class="grid-square %s"></div>`, selectedMaterial.CssClassName))
}
