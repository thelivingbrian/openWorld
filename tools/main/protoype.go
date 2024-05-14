package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/google/uuid"
)

type Prototype struct {
	ID          string `json:"id"`
	CommonName  string `json:"commonName"`
	CssColor    string `json:"cssColor"`
	Walkable    bool   `json:"walkable"` // More complex than bool? Or seperate OnStep Property? Transports?
	Floor1Css   string `json:"layer1css"`
	Floor2Css   string `json:"layer2css"`
	Ceiling1Css string `json:"ceiling1css"`
	Ceiling2Css string `json:"ceiling2css"`
	SetName     string `json:"setName"`
}

type Transformation struct {
	ClockwiseRotations int    `json:"clockwiseRotations,omitempty"`
	ColorPalette       string `json:"colorPalette,omitempty"`
}

type TileData struct {
	PrototypeId    string         `json:"prototypeId,omitempty"`
	Transformation Transformation `json:"transformation,omitempty"`
}

type PrototypeSelectPage struct {
	PrototypeSets []string
	CurrentSet    string
	Prototypes    []Prototype
}

func (c Context) prototypesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		c.getPrototypes(w, r)
	}
	if r.Method == "POST" {
		c.postPrototypes(w, r)
	}
}

func (c *Context) getPrototypes(w http.ResponseWriter, r *http.Request) {
	var PageData = c.prototypeSelectFromRequest(r)
	err := tmpl.ExecuteTemplate(w, "prototype-select", PageData)
	if err != nil {
		fmt.Println(err)
	}
}

func (c *Context) prototypeSelectFromRequest(r *http.Request) PrototypeSelectPage {
	queryValues := r.URL.Query()
	collectionName := queryValues.Get("currentCollection")
	setName := queryValues.Get("prototype-set")

	collection, ok := c.Collections[collectionName]
	if !ok {
		fmt.Println("Collection Name Invalid")
		return PrototypeSelectPage{}
	}

	var setOptions []string
	for key := range collection.PrototypeSets {
		setOptions = append(setOptions, key)
	}

	protos := collection.PrototypeSets[setName]
	transformedProtos := make([]Prototype, len(protos))
	for i := range protos {
		transformedProtos[i] = protos[i].peekTransform(Transformation{})
	}
	return PrototypeSelectPage{
		PrototypeSets: setOptions,
		CurrentSet:    setName,
		Prototypes:    transformedProtos,
	}
}

func (c Context) postPrototypes(w http.ResponseWriter, r *http.Request) {
	props, ok := requestToProperties(r)
	if !ok {
		fmt.Println("Invalid POST to /prototypes. Properties are invalid.")
		io.WriteString(w, `<h3> Properties are invalid. </h3>`)
		return
	}
	collectionName, ok := props["currentCollection"]
	if !ok {
		fmt.Println("Invalid POST to /prototypes. Collection not found.")
		io.WriteString(w, `<h3> Collection not found. </h3>`)
		return
	}
	setName, ok := props["prototype-set-name"]
	if !ok {
		fmt.Println("Invalid POST to prototypes. No Set Name.")
		io.WriteString(w, `<h3> No Set Name. </h3>`)
		return
	}
	fmt.Printf("%s %s \n", collectionName, setName)

	collection, ok := c.Collections[collectionName]
	if !ok {
		fmt.Println("Collection Name Invalid")
		return
	}

	collection.PrototypeSets[setName] = make([]Prototype, 0)

	// New Func
	outFile := c.collectionPath + collectionName + "/prototypes/" + setName + ".json"
	err := writeJsonFile(outFile, collection.Fragments[setName])
	if err != nil {
		panic(err)
	}

	io.WriteString(w, `<h2>Success</h2>`)
}

func (c Context) prototypesNewHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		getPrototypesNew(w, r)
	}
}

func getPrototypesNew(w http.ResponseWriter, _ *http.Request) {
	err := tmpl.ExecuteTemplate(w, "prototypes-new", nil)
	if err != nil {
		fmt.Println(err)
	}
}

func (c *Context) prototypeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		c.getPrototype(w, r)
	}
	if r.Method == "POST" {
		c.postPrototype(w, r)
	}
	if r.Method == "PUT" {
		c.putPrototype(w, r)
	}
}

func (c *Context) getPrototype(w http.ResponseWriter, r *http.Request) {
	queryValues := r.URL.Query()
	collectionName := queryValues.Get("currentCollection")
	id := queryValues.Get("prototype")
	collection, ok := c.Collections[collectionName]
	if !ok {
		panic("Help plz")
	}
	proto := collection.findPrototypeById(id)
	if proto == nil {
		panic("Invalid proto id")
	}

	err := tmpl.ExecuteTemplate(w, "prototype-edit", proto)
	if err != nil {
		fmt.Println(err)
	}
}

func (c Context) putPrototype(w http.ResponseWriter, r *http.Request) {
	fmt.Println("PUT for /prototype")

	properties, _ := requestToProperties(r)
	collectionName := properties["currentCollection"]
	setName := properties["prototype-set"]
	collection, ok := c.Collections[collectionName]
	if !ok {
		panic("no collection")
	}
	_, ok = collection.PrototypeSets[setName]
	if !ok {
		panic("no set")
	}

	id := properties["prototype-id"]
	commonName := properties["CommonName"]
	walkable := (properties["walkable"] == "on")
	cssColor := properties["CssColor"]
	floor1 := properties["Floor1Css"]
	floor2 := properties["Floor2Css"]
	ceiling1 := properties["Ceiling1Css"]
	ceiling2 := properties["Ceiling2Css"]
	fmt.Printf("%s | Floor: %s - %s Ceiling: %s - %s\n", commonName, floor1, floor2, ceiling1, ceiling2)
	panicIfAnyEmpty("Invalid prototype", id, commonName)

	proto := collection.findPrototypeById(id)
	if proto == nil {
		panic("no proto with that id")
	}
	proto.CommonName = commonName
	proto.Walkable = walkable
	proto.CssColor = cssColor
	proto.Floor1Css = floor1
	proto.Floor2Css = floor2
	proto.Ceiling1Css = ceiling1
	proto.Ceiling2Css = ceiling2

	fmt.Println(proto)

	outFile := c.collectionPath + collectionName + "/prototypes/" + setName + ".json"
	err := writeJsonFile(outFile, collection.PrototypeSets[setName])
	if err != nil {
		panic(err)
	}

	io.WriteString(w, "<h3>Done.</h3>")
}

func (c *Context) postPrototype(w http.ResponseWriter, r *http.Request) {
	fmt.Println("POST for /prototype")

	properties, _ := requestToProperties(r)
	collectionName := properties["currentCollection"]
	setName := properties["prototype-set"]
	collection, ok := c.Collections[collectionName]
	if !ok {
		panic("no collection")
	}
	set, ok := collection.PrototypeSets[setName]
	if !ok {
		panic("no set")
	}

	commonName := properties["CommonName"]
	walkable := (properties["walkable"] == "on")
	cssColor := properties["CssColor"]
	floor1 := properties["Floor1Css"]
	floor2 := properties["Floor2Css"]
	ceiling1 := properties["Ceiling1Css"]
	ceiling2 := properties["Ceiling2Css"]
	fmt.Printf("%s | Floor: %s - %s Ceiling: %s - %s\n", commonName, floor1, floor2, ceiling1, ceiling2)
	panicIfAnyEmpty("Invalid prototype", commonName) // The rest may be empty legitimately

	id := uuid.New().String()
	collection.PrototypeSets[setName] = append(set, Prototype{ID: id, SetName: setName, CommonName: commonName, Walkable: walkable, CssColor: cssColor, Floor1Css: floor1, Floor2Css: floor2, Ceiling1Css: ceiling1, Ceiling2Css: ceiling2})

	outFile := c.collectionPath + collectionName + "/prototypes/" + setName + ".json"
	err := writeJsonFile(outFile, collection.PrototypeSets[setName])
	if err != nil {
		panic(err)
	}
	io.WriteString(w, "<h3>Done.</h3>")
}

func panicIfAnyEmpty(errorMessage string, strings ...string) {
	for _, str := range strings {
		if str == "" {
			panic("panicIfAnyEmpty - caller provided error message: " + errorMessage)
		}
	}
}

func (c Context) prototypeNewHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		getPrototypeNew(w, r)
	}
}

func getPrototypeNew(w http.ResponseWriter, r *http.Request) {
	queryValues := r.URL.Query()
	setName := queryValues.Get("prototype-set")
	var pageData = struct {
		CurrentSet string
	}{
		CurrentSet: setName,
	}
	err := tmpl.ExecuteTemplate(w, "prototype-new", pageData)
	if err != nil {
		fmt.Println(err)
	}
}

func examplePrototype(w http.ResponseWriter, r *http.Request) {
	cssClass := r.URL.Query().Get("CssColor")
	floor1 := r.URL.Query().Get("Floor1Css")
	floor2 := r.URL.Query().Get("Floor2Css")
	ceiling1 := r.URL.Query().Get("Ceiling1Css")
	ceiling2 := r.URL.Query().Get("Ceiling2Css")
	layers := fmt.Sprintf(`<div class="box floor1 %s"> </div>
							<div class="box floor2 %s"> </div>
							<div class="box ceiling1 %s"></div>
							<div class="box ceiling2 %s"> </div>
							`, emptyTransformCss(floor1), emptyTransformCss(floor2), emptyTransformCss(ceiling1), emptyTransformCss(ceiling2))

	output := fmt.Sprintf(`<div id="exampleSquare" class="grid-row"><div class="grid-square %s">%s</div></div>`, cssClass, layers)
	io.WriteString(w, output)
}

// Utilities

func (proto *Prototype) applyTransform(transformation Transformation) Material {
	return Material{
		ID:          15793,
		CommonName:  proto.CommonName,
		CssColor:    proto.CssColor,
		Floor1Css:   transformCss(proto.Floor1Css, transformation),
		Floor2Css:   transformCss(proto.Floor2Css, transformation),
		Ceiling1Css: transformCss(proto.Ceiling1Css, transformation),
		Ceiling2Css: transformCss(proto.Ceiling2Css, transformation)}
}
func (proto *Prototype) applyTransformWithId(transformation Transformation, id int) Material {
	return Material{
		ID:          id,
		CommonName:  proto.CommonName,
		CssColor:    proto.CssColor,
		Floor1Css:   transformCss(proto.Floor1Css, transformation),
		Floor2Css:   transformCss(proto.Floor2Css, transformation),
		Ceiling1Css: transformCss(proto.Ceiling1Css, transformation),
		Ceiling2Css: transformCss(proto.Ceiling2Css, transformation)}
}

func (proto *Prototype) peekTransform(transformation Transformation) Prototype {
	return Prototype{
		ID:          proto.ID,
		SetName:     proto.SetName,
		CommonName:  proto.CommonName,
		CssColor:    proto.CssColor,
		Floor1Css:   transformCss(proto.Floor1Css, transformation),
		Floor2Css:   transformCss(proto.Floor2Css, transformation),
		Ceiling1Css: transformCss(proto.Ceiling1Css, transformation),
		Ceiling2Css: transformCss(proto.Ceiling2Css, transformation)}
}

// Template funcs
func (proto *Prototype) PeekFloor1() string {
	return transformCss(proto.Floor1Css, Transformation{})
}

func (proto *Prototype) PeekFloor2() string {
	return transformCss(proto.Floor2Css, Transformation{})
}

func (proto *Prototype) PeekCeiling1() string {
	return transformCss(proto.Ceiling1Css, Transformation{})
}

func (proto *Prototype) PeekCeiling2() string {
	return transformCss(proto.Ceiling2Css, Transformation{})
}
