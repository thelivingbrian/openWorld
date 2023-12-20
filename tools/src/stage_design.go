package main

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

type RequestData struct {
	height int
	width  int
}

func contentFromRequest(r *http.Request) (RequestData, bool) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("Error reading body: %v", err)
	}

	bodyS := string(body[:])
	input := strings.Split(bodyS, "&")
	height, _ := strconv.Atoi(strings.Split(input[0], "=")[1])
	width, _ := strconv.Atoi(strings.Split(input[1], "=")[1])

	fmt.Printf(bodyS)
	return RequestData{height, width}, true
}

func getGridHTML(h int, w int) string {
	output := ""
	for y := 0; y < h; y++ {
		output += `<div class="grid-row">`
		for x := 0; x < w; x++ {
			var yStr = strconv.Itoa(y)
			var xStr = strconv.Itoa(x)
			output += `<div hx-post="/new" hx-trigger="click" hx-include="#selectedColor" hx-headers='{"y": "` + yStr + `", "x": "` + xStr + `"}' class="grid-square id="c` + yStr + `-` + xStr + `"></div>`
		}
		output += `</div>`
	}
	return output
}

func createGrid(w http.ResponseWriter, r *http.Request) {
	content, success := contentFromRequest(r)
	if !success {
		panic(0)
	}

	output := `
    <div id="page">
        <div id="controls">
            <form hx-post="/createGrid" hx-target="#page" hx-swap="outerHTML">
                <div>
                <label>Enter Height and Width:</label>
                <input type="text" name="height" value="10">
                <input type="text" name="width" value="10">
                </div>
                <button>Create</button>
            </form>
        </div>
        <div class="grid" id="screen">`

	output += getGridHTML(content.height, content.width)
	output += `</div>
	<div class="color-selector">
		<input hx-post="/select" type="radio" id="colorRed" name="color" value="red" checked>
		<label for="colorRed" class="color-box red"></label>

		<input type="radio" id="colorGreen" name="color" value="green">
		<label for="colorGreen" class="color-box green"></label>

		<!-- Repeat for other colors -->
		<input id="selectedColor" type="hidden" name="selectedColor" value="xyz" />
	</div>
</div></div>`
	io.WriteString(w, output)

}

func dataFromRequest(r *http.Request) (ColorData, bool) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("Error reading body: %v", err)
	}

	bodyS := string(body[:])
	fmt.Printf(bodyS)
	fmt.Println("Body string above")

	for key, header := range r.Header {
		fmt.Printf(key)
		fmt.Println(header[0])
	}

	return ColorData{"blue", 0, 0}, true
}

func selectColor(w http.ResponseWriter, r *http.Request) {
	content, success := dataFromRequest(r)
	if !success {
		panic(0)
	}
	content.x += 1
}

type ColorData struct {
	color string
	y     int
	x     int
}
