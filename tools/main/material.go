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
	ID         int    `json:"id"`
	CommonName string `json:"commonName"`
	CssColor   string `json:"cssColor"`
	Walkable   bool   `json:"walkable"`
	Layer1Css  string `json:"layer1css"`
	Layer2Css  string `json:"layer2css"`
}

type Color struct {
	CssClassName string `json:"cssClassName"`
	R            int    `json:"R"`
	G            int    `json:"G"`
	B            int    `json:"B"`
	A            string `json:"A"`
}

var R, G, B int

func getMaterialPage(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, materialPageHTML())
}

func materialPageHTML() string {
	output := ""
	output += divEditColorSelect()
	output += divEditMaterialSelect()
	output += `<br/>
				<div id="edit-ingredient-window">
				
				</div><br />
				<div id="output-ingredients">
					<button class="btn" hx-post="/outputIngredients" hx-target="#panel">Output changes</button>
				</div>`
	return output
}

func divEditColorSelect() string {
	output := `
	<div>
		<label>Colors</label>
		<select name="colorId" hx-get="/getEditColor" hx-target="#edit-ingredient-window">
			<option value="">--</option>			`
	for i, color := range colors {
		output += fmt.Sprintf(`<option value="%d">%s</option>`, i, color.CssClassName)
	}
	output += `		
		</select>
		<button class="btn" hx-get="/getNewColor" hx-target="#edit-ingredient-window">New</button>
	</div>`

	return output
}

func divEditMaterialSelect() string {
	output := `
	<div>
		<label>Materials</label>
		<select name="materialId" hx-get="/getEditMaterial" hx-target="#edit-ingredient-window">
			<option value="">--</option>			`
	for _, material := range materials {
		output += fmt.Sprintf(`<option value="%d">%s</option>`, material.ID, material.CommonName)
	}
	output += `		
		</select>
		<button class="btn" hx-get="/getNewMaterial" hx-target="#selectedMaterial">New</button>
	</div>`

	return output
}

func getEditColor(w http.ResponseWriter, r *http.Request) {
	queryValues := r.URL.Query()
	id, err := strconv.Atoi(queryValues.Get("colorId"))
	if err != nil {
		return
	}

	color := colors[id]
	R = color.R
	G = color.G
	B = color.B
	output := fmt.Sprintf(`<div id="exampleSquare" class="grid-row"><div class="grid-square" style="background-color:rgb(%d,%d,%d)"></div></div>`, R, G, B)

	editForm := `
	<form hx-put="/editMaterial" hx-target="#panel">
		<div>
			<label>Css Class</label>
			<input type="text" name="CommonName" value="%s">
		</div>
		<div>
			<label>R: </label>
			<input hx-get="/exampleSquare" hx-trigger="change" hx-target="#exampleSquare" type="text" name="R" value="%d">
			<label>G: </label>
			<input hx-get="/exampleSquare" hx-trigger="change" hx-target="#exampleSquare" type="text" name="G" value="%d">
			<label>B: </label>
			<input hx-get="/exampleSquare" hx-trigger="change" hx-target="#exampleSquare" type="text" name="B" value="%d">
			<label>A: </label>
			<inputtype="text" name="A" value="%s">
		</div>
		<input type="hidden" name="colorId" value="%d">
		<button class="btn">Save</button>
	</form>
	`
	output += fmt.Sprintf(editForm, color.CssClassName, color.R, color.G, color.B, color.A, id)

	io.WriteString(w, output)
}

func getEditMaterial(w http.ResponseWriter, r *http.Request) {
	queryValues := r.URL.Query()
	id, err := strconv.Atoi(queryValues.Get("materialId"))
	if err != nil {
		return
	}

	material := materials[id] //materialMap[name]
	color, ok := sliceToMap(colors, colorName)[material.CssColor]
	if !ok {
		fmt.Println("No Color")
		return
	}
	R = color.R
	G = color.G
	B = color.B
	output := fmt.Sprintf(`<div id="exampleSquare" class="grid-row"><div class="grid-square" style="background-color:rgb(%d,%d,%d)"></div></div>`, R, G, B)

	walkableIndicator := ""
	if material.Walkable {
		walkableIndicator = "checked"
	}

	editForm := `
	<form hx-put="/editMaterial" hx-target="#panel">
		<div>
			<label>Name: (ID: %d)</label>
			<input type="text" name="CommonName" value="%s">
		</div>
		<div>
			<label>Css Color Name: </label>
			<input type="text" name="CssColor" value="%s">
		</div>
		<div>
			<label>Layer 1 Css: </label>
			<input type="text" name="Layer1Css" value="">
		</div>
		<div>
			<label>Layer 2 Css: </label>
			<input type="text" name="Layer2Css" value="">
		</div>

		<input type="hidden" name="materialId" value="%d">
		<label>Walkable: </label>
		<input type="checkbox" name="walkable" %s>
		<button class="btn">Save</button>
	</form>
	`
	output += fmt.Sprintf(editForm, material.ID, material.CommonName, material.CssColor, material.ID, walkableIndicator)

	io.WriteString(w, output)
}

func editMaterial(w http.ResponseWriter, r *http.Request) {
	properties, _ := requestToProperties(r)
	materialId, _ := strconv.Atoi(properties["materialId"])
	commonName := properties["CommonName"]
	cssColor := properties["CssColor"]
	walkable := properties["walkable"]

	fmt.Printf("%d %s %s\n%s\n", materialId, commonName, cssColor, walkable)

	material := &materials[materialId]
	if material.ID != materialId {
		panic("Material IDs are corrupted")
	}
	material.CommonName = commonName
	material.CssColor = cssColor
	material.Walkable = (walkable == "on")
	material.Layer1Css = ""
	material.Layer2Css = ""
	io.WriteString(w, materialPageHTML())
}

func editColor(w http.ResponseWriter, r *http.Request) {
	properties, _ := requestToProperties(r)
	colorId, _ := strconv.Atoi(properties["colorId"])
	cssClassName := properties["CssClassName"]
	red, _ := strconv.Atoi(properties["R"])
	green, _ := strconv.Atoi(properties["G"])
	blue, _ := strconv.Atoi(properties["B"])
	alpha := properties["A"]

	fmt.Printf("%d %s %d %d %d %s\n", colorId, cssClassName, red, green, blue, alpha)

	color := &colors[colorId]
	color.CssClassName = cssClassName
	color.R = red
	color.G = green
	color.B = blue
	color.A = alpha

	io.WriteString(w, materialPageHTML())
}

func getNewMaterial(w http.ResponseWriter, r *http.Request) {
	newForm := `
	<form hx-post="/newMaterial" hx-target="#panel">
		<div>
			<label>Name: </label>
			<input type="text" name="CommonName" value="">
		</div>
		<div>
			<label>Class Name: </label>
			<input type="text" name="CssClassName" value="">
		</div>
		<div>
			<label>Css Color Name: </label>
			<input type="text" name="CssColor" value="%s">
		</div>
		<div>
			<label>Layer 1 Css: </label>
			<input type="text" name="Layer1Css" value="">
		</div>
		<div>
			<label>Layer 2 Css: </label>
			<input type="text" name="Layer2Css" value="">
		</div>
		<div>
			<label>Walkable: </label>
			<input type="checkbox" name="walkable" %s />
		</div>
	<button class="btn">Save</button>
	</form>
	`
	io.WriteString(w, newForm)
}

func newMaterial(w http.ResponseWriter, r *http.Request) {
	properties, _ := requestToProperties(r)
	materialId := len(materials)
	commonName := properties["CommonName"]
	walkable := (properties["walkable"] == "on")
	cssColor := properties["CssColor"]

	fmt.Printf("%s %s\n%s\n", commonName, cssColor, properties["walkable"])

	material := Material{ID: materialId, CommonName: commonName, CssColor: cssColor, Layer1Css: "", Layer2Css: "", Walkable: walkable}

	materialMap := sliceToMap(materials, materialName)
	_, ok := materialMap[commonName]
	if !ok {
		materials = append(materials, material)
	} else {
		panic("Duplicate name")
	}

	io.WriteString(w, materialPageHTML())
}

func getNewColor(w http.ResponseWriter, r *http.Request) {
	newForm := `
	<form hx-post="/newColor" hx-target="#panel">
		<div>
			<label>Css Class Name: </label>
			<input type="text" name="CssClassName" value="">
		</div>
		<div>
			<label>R: </label>
			<input type="text" name="R" value="">
			<label>G: </label>
			<input type="text" name="G" value="">
			<label>B: </label>
			<input type="text" name="B" value="">
			<label>A: </label>
			<input type="text" name="A" value="">
		</div>
		<button class="btn">Save</button>
	</form>
	`
	io.WriteString(w, newForm)
}

func newColor(w http.ResponseWriter, r *http.Request) {
	properties, _ := requestToProperties(r)
	colorIndex := len(colors)
	cssClassName := properties["CssClassName"]
	R, _ := strconv.Atoi(properties["R"])
	G, _ := strconv.Atoi(properties["G"])
	B, _ := strconv.Atoi(properties["B"])
	A := properties["A"]

	fmt.Printf("%d %s\nr%d g%d b%d a%s\n", colorIndex, cssClassName, R, G, B, A)

	color := Color{CssClassName: cssClassName, R: R, G: G, B: B, A: A}

	colorMap := sliceToMap(colors, colorName)
	_, ok := colorMap[cssClassName]
	if !ok {
		colors = append(colors, color)
	} else {
		panic("Duplicate name")
	}

	io.WriteString(w, materialPageHTML())
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

func outputIngredients(w http.ResponseWriter, r *http.Request) {
	err := WriteMaterialsToFile()
	if err != nil {
		panic(1)
	}

	err = WriteColorsToFile()
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

func WriteColorsToFile() error {
	data, err := json.Marshal(colors)
	if err != nil {
		return fmt.Errorf("error marshalling colorss: %w", err)
	}

	file, err := os.Create("./level/data/colors.json")
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
	cssFile, err := os.Create("./level/assets/colors.css")
	if err != nil {
		panic(err)
	}
	defer cssFile.Close()

	for _, color := range colors {
		rgbstring := fmt.Sprintf("rgb(%d, %d, %d)", color.R, color.G, color.B)
		if color.A != "" {
			rgbstring = fmt.Sprintf("rgba(%d, %d, %d, %s)", color.R, color.G, color.B, color.A)
		}
		cssRule := fmt.Sprintf(".%s { background-color: %s; }\n\n.%s-b { border-color: %s; }\n\n", color.CssClassName, rgbstring, color.CssClassName, rgbstring)
		_, err := cssFile.WriteString(cssRule)
		if err != nil {
			panic(err)
		}
	}
}
