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
	CommonName   string `json:"commonName"`
	CssClassName string `json:"cssClassName"`
	Walkable     bool   `json:"walkable"`
	R            int    `json:"R"`
	G            int    `json:"G"`
	B            int    `json:"B"`
}

var R, G, B int

func getMaterialPage(w http.ResponseWriter, r *http.Request) {
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
	<button class="btn" hx-post="/submit" hx-target="#panel">Output changes</button>`
	return output
}

func exampleSquare(w http.ResponseWriter, r *http.Request) {
	red, err := strconv.Atoi(r.URL.Query().Get("R"))
	if err != nil {
		red = R
	} else {
		R = red
	}

	green, err := strconv.Atoi(r.URL.Query().Get("G"))
	if err != nil {
		green = G
	} else {
		G = green
	}

	blue, err := strconv.Atoi(r.URL.Query().Get("B"))
	if err != nil {
		blue = B
	} else {
		B = blue
	}

	output := fmt.Sprintf(`<div id="exampleSquare" class="grid-row"><div class="grid-square" style="background-color:rgb(%d,%d,%d)"></div></div>`, red, green, blue)
	io.WriteString(w, output)
}

func getMaterial(w http.ResponseWriter, r *http.Request) {
	queryValues := r.URL.Query()
	id, _ := strconv.Atoi(queryValues.Get("materialId"))

	fmt.Printf("Received id: %d", id)

	material := materials[id]
	R = material.R
	G = material.G
	B = material.B
	output := fmt.Sprintf(`<div id="exampleSquare" class="grid-row"><div class="grid-square" style="background-color:rgb(%d,%d,%d)"></div></div>`, R, G, B)

	walkableIndicator := ""
	if material.Walkable {
		walkableIndicator = "checked"
	}

	editForm := `
	<form hx-put="/materialEdit" hx-target="#panel">
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
		<input hx-get="/exampleSquare" hx-trigger="change" hx-target="#exampleSquare" type="text" name="R" value="%d">
		<label>G: </label>
		<input hx-get="/exampleSquare" hx-trigger="change" hx-target="#exampleSquare" type="text" name="G" value="%d">
		<label>B: </label>
		<input hx-get="/exampleSquare" hx-trigger="change" hx-target="#exampleSquare" type="text" name="B" value="%d">
	</div>
	<input type="hidden" name="materialId" value="%d">
	<label>Walkable: </label>
	<input type="checkbox" name="walkable" %s>
	<button class="btn">Save</button>
	</form>
	`
	output += fmt.Sprintf(editForm, id, material.CommonName, material.CssClassName,
		material.R, material.G, material.B, id, walkableIndicator)

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
	properties, _ := requestToProperties(r)
	materialId, _ := strconv.Atoi(properties["materialId"])
	commonName := properties["CommonName"]
	cssClass := properties["CssClassName"]
	walkable := properties["walkable"]
	red, _ := strconv.Atoi(properties["R"])
	green, _ := strconv.Atoi(properties["G"])
	blue, _ := strconv.Atoi(properties["B"])

	fmt.Printf("%d %s %s %d %d %d\n%s\n", materialId, commonName, cssClass, red, green, blue, walkable)

	for i := range materials {
		if materials[i].ID == materialId {
			materials[i].CommonName = commonName
			materials[i].CssClassName = cssClass
			materials[i].Walkable = (walkable == "on")
			materials[i].R = red
			materials[i].G = green
			materials[i].B = blue
		}
	}
	io.WriteString(w, materialPageHTML())
}

func materialNew(w http.ResponseWriter, r *http.Request) {
	properties, _ := requestToProperties(r)
	materialId := len(materials)
	commonName := properties["CommonName"]
	walkable := (properties["walkable"] == "on")
	cssClass := properties["CssClassName"]
	red, _ := strconv.Atoi(properties["R"])
	green, _ := strconv.Atoi(properties["G"])
	blue, _ := strconv.Atoi(properties["B"])

	fmt.Printf("%d %s %s %d %d %d\n%s\n", materialId, commonName, cssClass, red, green, blue, properties["walkable"])

	material := Material{ID: materialId, R: red, G: green, B: blue, CssClassName: cssClass, CommonName: commonName, Walkable: walkable}

	materials = append(materials, material)
	io.WriteString(w, materialPageHTML())
}

func newMaterialForm(w http.ResponseWriter, r *http.Request) {
	newForm := `
	<form hx-post="/materialNew" hx-target="#panel">
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
	data, err := json.Marshal(materials)
	if err != nil {
		return fmt.Errorf("error marshalling materials: %w", err)
	}

	file, err := os.Create("./level/data/materials.json")
	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}
	defer file.Close()

	_, err = file.Write(data)
	if err != nil {
		return fmt.Errorf("error writing to file: %w", err)
	}

	return nil
}

func createCSSFile() {
	cssFile, err := os.Create("./level/assets/materials.css")
	if err != nil {
		panic(err)
	}
	defer cssFile.Close()

	for _, material := range materials {
		cssRule := fmt.Sprintf(".%s {\n    background-color: rgb(%d, %d, %d);\n}\n\n", material.CssClassName, material.R, material.G, material.B)
		_, err := cssFile.WriteString(cssRule)
		if err != nil {
			panic(err)
		}
	}
}
