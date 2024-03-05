package main

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
)

type Material struct {
	ID          int    `json:"id"`
	CommonName  string `json:"commonName"`
	CssColor    string `json:"cssColor"`
	Walkable    bool   `json:"walkable"`
	Floor1Css   string `json:"layer1css"`
	Floor2Css   string `json:"layer2css"`
	Ceiling1Css string `json:"ceiling1css"`
	Ceiling2Css string `json:"ceiling2css"`
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
		<button class="btn" hx-get="/getNewMaterial" hx-target="#edit-ingredient-window">New</button>
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
			<input type="text" name="A" value="%s">
		</div>
		<input type="hidden" name="colorId" value="%d">
		<button class="btn">Save</button>
	</form>
	`
	output += fmt.Sprintf(editForm, color.CssClassName, color.R, color.G, color.B, color.A, id)

	io.WriteString(w, output)
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

func getEditMaterial(w http.ResponseWriter, r *http.Request) {
	queryValues := r.URL.Query()
	id, err := strconv.Atoi(queryValues.Get("materialId"))
	if err != nil {
		return
	}

	material := materials[id]
	color, ok := sliceToMap(colors, colorName)[material.CssColor]
	A := "1.0"
	if !ok {
		fmt.Println("No Color")
		color = colors[0]
		A = "0"
	}
	R = color.R
	G = color.G
	B = color.B

	overlay := fmt.Sprintf(`<div class="box floor1 %s"></div><div class="box floor2 %s"></div><div class="box ceiling1 %s"></div><div class="box ceiling2 %s"></div>`, material.Floor1Css, material.Floor2Css, material.Ceiling1Css, material.Ceiling2Css)
	output := fmt.Sprintf(`<div id="exampleSquare" class="grid-row"><div class="grid-square" style="background-color:rgba(%d,%d,%d,%s)">%s</div></div>`, R, G, B, A, overlay)

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
			<input hx-get="/exampleMaterial" hx-trigger="change" hx-target="#exampleSquare" hx-include="[name='Floor1Css'],[name='Floor2Css'],[name='Ceiling1Css'],[name='Ceiling2Css']" type="text" name="CssColor" value="%s">
		</div>
		<div>
			<label>Floor 1 Css: </label>
			<input hx-get="/exampleMaterial" hx-trigger="change" hx-target="#exampleSquare" hx-include="[name='Floor2Css'],[name='Ceiling1Css'],[name='Ceiling2Css'],[name='CssColor']" type="text" name="Floor1Css" value="%s">
		</div>
		<div>
			<label>Floor 2 Css: </label>
			<input hx-get="/exampleMaterial" hx-trigger="change" hx-target="#exampleSquare" hx-include="[name='Floor1Css'],[name='Ceiling1Css'],[name='Ceiling2Css'],[name='CssColor']" type="text" name="Floor2Css" value="%s">
		</div>
		<div>
			<label>Ceiling 1 Css: </label>
			<input hx-get="/exampleMaterial" hx-trigger="change" hx-target="#exampleSquare" hx-include="[name='Ceiling2Css'],[name='Floor1Css'],[name='Floor2Css'],[name='CssColor']" type="text" name="Ceiling1Css" value="%s">
		</div>
		<div>
			<label>Ceiling 2 Css: </label>
			<input hx-get="/exampleMaterial" hx-trigger="change" hx-target="#exampleSquare" hx-include="[name='Ceiling1Css'],[name='Floor1Css'],[name='Floor2Css'],[name='CssColor']" type="text" name="Ceiling2Css" value="%s">
		</div>

		<input type="hidden" name="materialId" value="%d">
		<label>Walkable: </label>
		<input type="checkbox" name="walkable" %s>
		<button class="btn">Save</button>
	</form>
	`
	output += fmt.Sprintf(editForm, material.ID, material.CommonName, material.CssColor, material.Floor1Css, material.Floor2Css, material.Ceiling1Css, material.Ceiling2Css, material.ID, walkableIndicator)

	io.WriteString(w, output)
}

func editMaterial(w http.ResponseWriter, r *http.Request) {
	properties, _ := requestToProperties(r)
	materialId, _ := strconv.Atoi(properties["materialId"])
	commonName := properties["CommonName"]
	cssColor := properties["CssColor"]
	walkable := properties["walkable"]
	floor1 := properties["Floor1Css"]
	floor2 := properties["Floor2Css"]
	ceiling1 := properties["Ceiling1Css"]
	ceiling2 := properties["Ceiling2Css"]

	fmt.Printf("%d %s %s\n%s\n", materialId, commonName, cssColor, walkable)

	material := &materials[materialId]
	if material.ID != materialId {
		panic("Material IDs are corrupted")
	}
	material.CommonName = commonName
	material.CssColor = cssColor
	material.Walkable = (walkable == "on")
	material.Floor1Css = floor1
	material.Floor2Css = floor2
	material.Ceiling1Css = ceiling1
	material.Ceiling2Css = ceiling2
	io.WriteString(w, materialPageHTML())
}

func getNewMaterial(w http.ResponseWriter, r *http.Request) {
	newForm := `
	<form hx-post="/newMaterial" hx-target="#panel">
		<div id="exampleSquare" class="grid-row">
			<div class="grid-square"></div>
		</div>
		<div>
			<label>Name: </label>
			<input type="text" name="CommonName" value="">
		</div>
		<div>
			<label>Css Color Name: </label>
			<input hx-get="/exampleMaterial" hx-trigger="change" hx-target="#exampleSquare" hx-include="[name='Floor1Css'],[name='Floor2Css'],[name='Ceiling1Css'],[name='Ceiling2Css']" type="text" name="CssColor" value="">
		</div>
		<div>
			<label>Floor 1 Css: </label>
			<input hx-get="/exampleMaterial" hx-trigger="change" hx-target="#exampleSquare" hx-include="[name='CssColor'],[name='Floor2Css'],[name='Ceiling1Css'],[name='Ceiling2Css']" type="text" name="Floor1Css" value="">
		</div>
		<div>
			<label>Floor 2 Css: </label>
			<input hx-get="/exampleMaterial" hx-trigger="change" hx-target="#exampleSquare" hx-include="[name='Floor1Css'],[name='CssColor'],[name='Ceiling1Css'],[name='Ceiling2Css']" type="text" name="Floor2Css" value="">
		</div>
		<div>
			<label>Ceiling 1 Css: </label>
			<input hx-get="/exampleMaterial" hx-trigger="change" hx-target="#exampleSquare" hx-include="[name='Floor1Css'],[name='Floor2Css'],[name='CssColor'],[name='Ceiling2Css']" type="text" name="Ceiling1Css" value="">
		</div>
		<div>
			<label>Ceiling 2 Css: </label>
			<input hx-get="/exampleMaterial" hx-trigger="change" hx-target="#exampleSquare" hx-include="[name='Floor1Css'],[name='Floor2Css'],[name='Ceiling1Css'],[name='CssColor']" type="text" name="Ceiling2Css" value="">
		</div>
		<div>
			<label>Walkable: </label>
			<input type="checkbox" name="walkable" />
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
	floor1 := properties["Floor1Css"]
	floor2 := properties["Floor2Css"]
	ceiling1 := properties["Ceiling1Css"]
	ceiling2 := properties["Ceiling2Css"]

	fmt.Printf("%s | Floor: %s - %s Ceiling: %s - %s\n", commonName, floor1, floor2, ceiling1, ceiling2)

	material := Material{ID: materialId, CommonName: commonName, CssColor: cssColor, Floor1Css: floor1, Floor2Css: floor2, Ceiling1Css: ceiling1, Ceiling2Css: ceiling2, Walkable: walkable}

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
	cssClassName := properties["CssClassName"]
	R, _ := strconv.Atoi(properties["R"])
	G, _ := strconv.Atoi(properties["G"])
	B, _ := strconv.Atoi(properties["B"])
	A := properties["A"]

	//fmt.Printf("%d %s\nr%d g%d b%d a%s\n", colorIndex, cssClassName, R, G, B, A)

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

func exampleMaterial(w http.ResponseWriter, r *http.Request) {
	cssClass := r.URL.Query().Get("CssColor")
	floor1 := r.URL.Query().Get("Floor1Css")
	floor2 := r.URL.Query().Get("Floor2Css")
	ceiling1 := r.URL.Query().Get("Ceiling1Css")
	ceiling2 := r.URL.Query().Get("Ceiling2Css")
	layers := fmt.Sprintf(`<div class="box floor1 %s"> </div>
							<div class="box floor2 %s"> </div>
							<div class="box ceiling1 %s"></div>
							<div class="box ceiling2 %s"> </div>
							`, floor1, floor2, ceiling1, ceiling2)

	output := fmt.Sprintf(`<div id="exampleSquare" class="grid-row"><div class="grid-square %s">%s</div></div>`, cssClass, layers)
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
