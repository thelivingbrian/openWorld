package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
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
		<option value="">--</option>
			
`
	for _, material := range materials {
		output += fmt.Sprintf(`<option value="%d">%s</option>`, material.ID, material.CommonName)
	}
	output += `		
		</select>
		<button class="btn" hx-get="/newMaterialForm" hx-target="#selectedMaterial">New</button>
	</div>
	<div>
		<label><b>Material: </b></label>
		<div id="selectedMaterial">
		</div>
	</div>
	<button class="btn" hx-post="/submit" hx-target="#page">Output changes</button>`
	return output
}

func getMaterial(w http.ResponseWriter, r *http.Request) {
	queryValues := r.URL.Query()
	id := queryValues.Get("materialId")

	fmt.Printf("Received id: %s", id)

	output := ""
	for _, material := range materials {
		if id, _ := strconv.Atoi(id); id == material.ID {
			editForm := `
			<form hx-put="/materialEdit" hx-target="#page">
			<div>

				<label>Name: (ID: %d)</label>
				<input type="text" name="CommonName" value="%s">
			</div>
			<div>
				<label>Class Name: </label>
				<input type="text" name="CssClassName" value="%s">
			</div>
			<div>
				<label>R: </label>
				<input type="text" name="R" value="%d">
				<label>G: </label>
				<input type="text" name="G" value="%d">
				<label>B: </label>
				<input type="text" name="B" value="%d">
			</div>
			<input type="hidden" name="materialId" value="%d">
			<button class="btn">Save</button>
			</form>
			`
			output += fmt.Sprintf(editForm, id, material.CommonName, material.CssClassName,
				material.R, material.G, material.B, id)
		}
	}
	io.WriteString(w, output)
}

func stringToPropertyMap(body string) map[string]string {
	propMap := make(map[string]string)
	props := strings.Split(body, "&")
	for _, prop := range props {
		keyValue := strings.Split(prop, "=")
		propMap[keyValue[0]] = keyValue[1]
	}
	return propMap
}

func materialEdit(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("Error reading body: %v", err)
		return
	}

	bodyS := string(body[:])
	properties := stringToPropertyMap(bodyS)
	materialId, _ := strconv.Atoi(properties["materialId"])
	commonName := properties["CommonName"]
	cssClass := properties["CssClassName"]
	red, _ := strconv.Atoi(properties["R"])
	green, _ := strconv.Atoi(properties["G"])
	blue, _ := strconv.Atoi(properties["B"])

	fmt.Printf("%d %s %s %d %d %d\n\n", materialId, commonName, cssClass, red, green, blue)

	for i, _ := range materials {
		if materials[i].ID == materialId {
			materials[i].CommonName = commonName
			materials[i].CssClassName = cssClass
			materials[i].R = red
			materials[i].G = green
			materials[i].B = blue
		}
	}
	io.WriteString(w, materialPageHTML())
}

func materialNew(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("Error reading body: %v", err)
		return
	}

	bodyS := string(body[:])
	fmt.Println(bodyS)
	properties := stringToPropertyMap(bodyS)
	materialId, _ := strconv.Atoi(properties["materialId"])
	commonName := properties["CommonName"]
	cssClass := properties["CssClassName"]
	red, _ := strconv.Atoi(properties["R"])
	green, _ := strconv.Atoi(properties["G"])
	blue, _ := strconv.Atoi(properties["B"])

	fmt.Printf("%d %s %s %d %d %d\n\n", materialId, commonName, cssClass, red, green, blue)

	material := Material{ID: materialId, R: red, G: green, B: blue, CssClassName: cssClass, CommonName: commonName}

	materials = append(materials, material)
	io.WriteString(w, materialPageHTML())
}

func newMaterialForm(w http.ResponseWriter, r *http.Request) {
	newForm := `
	<form hx-post="/materialNew" hx-target="#page">
	<div>
		<label>ID: </label>
		<input type="text" name="materialId" value="">
	</div>
	<div>
		<label>Name: </label>
		<input type="text" name="CommonName" value="">
	</div>
	<div>
		<label>Class Name: </label>
		<input type="text" name="CssClassName" value="">
	</div>
	<div>
		<label>R: </label>
		<input type="text" name="R" value="">
		<label>G: </label>
		<input type="text" name="G" value="">
		<label>B: </label>
		<input type="text" name="B" value="">
	</div>
	<button class="btn">Save</button>
	</form>
	`
	io.WriteString(w, newForm)
}

func submit(w http.ResponseWriter, r *http.Request) {
	err := WriteMaterialsToFile()

	if err != nil {
		panic(1)
	}

	createCSSFile()

	getMaterialPage(w, r)
}

func WriteMaterialsToFile() error {
	// Serialize the materials slice to JSON
	data, err := json.Marshal(materials)
	if err != nil {
		return fmt.Errorf("error marshalling materials: %w", err)
	}

	// Create and open a file
	file, err := os.Create("./tools/level/data/materials.json")
	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}
	defer file.Close()

	// Write the JSON data to the file
	_, err = file.Write(data)
	if err != nil {
		return fmt.Errorf("error writing to file: %w", err)
	}

	return nil
}

func createCSSFile() {
	// Read the JSON file.
	/*jsonData, err := os.ReadFile("materials.json")
	if err != nil {
		panic(err)
	}*/

	// Parse the JSON data.
	//var materials []Material
	/*if err := json.Unmarshal(jsonData, &materials); err != nil {
		panic(err)
	}*/

	// Open the CSS file for writing.
	cssFile, err := os.Create("./tools/level/data/materials.css")
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
