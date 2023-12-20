package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
)

type Material struct {
	ID           int    `json:"id"`
	R            int    `json:"R"`
	G            int    `json:"G"`
	B            int    `json:"B"`
	CssClassName string `json:"cssClassName"`
	CommonName   string `json:"commonName"`
}

var materials []Material

func getMaterialPage(w http.ResponseWriter, r *http.Request) {
	jsonData, err := os.ReadFile("./tools/level/data/materials.json")
	if err != nil {
		panic(err)
	}

	// Parse the JSON data.
	if err := json.Unmarshal(jsonData, &materials); err != nil {
		panic(err)
	}

	io.WriteString(w, materialPageHTML())
}

func materialPageHTML() string {
	output := `
	<div>
		<label>Materials</label>
		<select name="materialId" hx-get="/material" hx-target="#selectedMaterial">
			
`
	for _, material := range materials {
		output += fmt.Sprintf(`<option value="%d">%s</option>`, material.ID, material.CommonName)
	}
	output += `		
		</select>
	</div>
	<div>
		<label>Material</label>
		<div id="selectedMaterial">
		</div>
	</div>`
	return output
}

func getMaterial(w http.ResponseWriter, r *http.Request) {
	queryValues := r.URL.Query()
	id := queryValues.Get("materialId")

	fmt.Println("Received id: %s", id)

	output := ""
	for _, material := range materials {
		if id, _ := strconv.Atoi(id); id == material.ID {
			output += fmt.Sprintf(`<h3>%s</h3><br /><h4>%s</h4></br /><p>R:%dG:%dB:%d</p>`,
				material.CommonName, material.CssClassName, material.R, material.G, material.B)
		}
	}
	io.WriteString(w, output)
}

func createCSSFile() {
	// Read the JSON file.
	jsonData, err := os.ReadFile("materials.json")
	if err != nil {
		panic(err)
	}

	// Parse the JSON data.
	//var materials []Material
	if err := json.Unmarshal(jsonData, &materials); err != nil {
		panic(err)
	}

	// Open the CSS file for writing.
	cssFile, err := os.Create("materials.css")
	if err != nil {
		panic(err)
	}
	defer cssFile.Close()

	// Write CSS rules for each material.
	for _, material := range materials {
		cssRule := fmt.Sprintf(".%s {\n    background-color: rgb(%d, %d, %d);\n}\n\n", material.CssClassName, material.R, material.G, material.B)
		_, err := cssFile.WriteString(cssRule)
		if err != nil {
			panic(err)
		}
	}
}
