package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/google/uuid"
)

type InteractableDescription struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	SetName   string `json:"setName"`
	CssClass  string `json:"cssClass"`
	Pushable  bool   `json:"pushable"`
	Fragile   bool   `json:"fragile"`
	Reactions string `json:"reactions"`
}

func (c Context) interactablesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		c.getInteractables(w, r)
	}
	if r.Method == "POST" {
		c.postInteractables(w, r)
	}
}

func (c *Context) getInteractables(w http.ResponseWriter, r *http.Request) {
	queryValues := r.URL.Query()
	collectionName := queryValues.Get("currentCollection")
	setName := queryValues.Get("interactable-set")

	collection, ok := c.Collections[collectionName]
	if !ok {
		io.WriteString(w, `<h3> Collection not found. </h3>`)
	}

	var setOptions []string
	for key := range collection.InteractableSets {
		setOptions = append(setOptions, key)
	}

	interactables := collection.InteractableSets[setName]

	var pageData = struct {
		SetNames      []string
		CurrentSet    string
		Interactables []InteractableDescription
	}{
		SetNames:      setOptions,
		CurrentSet:    setName,
		Interactables: interactables,
	}

	err := tmpl.ExecuteTemplate(w, "interactable-select", pageData)
	if err != nil {
		fmt.Println(err)
	}
}

func (c Context) postInteractables(w http.ResponseWriter, r *http.Request) {
	props, ok := requestToProperties(r)
	if !ok {
		fmt.Println("Invalid POST to /interactables. Properties are invalid.")
		io.WriteString(w, `<h3> Properties are invalid. </h3>`)
		return
	}
	collectionName, ok := props["currentCollection"]
	if !ok {
		fmt.Println("Invalid POST to /interactables. Collection not found.")
		io.WriteString(w, `<h3> Collection not found. </h3>`)
		return
	}
	setName, ok := props["interactable-set-name"]
	if !ok {
		fmt.Println("Invalid POST to interactables. No Set Name.")
		io.WriteString(w, `<h3> No Set Name. </h3>`)
		return
	}
	fmt.Printf("%s %s \n", collectionName, setName)

	collection, ok := c.Collections[collectionName]
	if !ok {
		fmt.Println("Collection Name Invalid")
		return
	}

	collection.InteractableSets[setName] = make([]InteractableDescription, 0)
	collection.saveInteractableSet(setName)

	io.WriteString(w, `<h2>Success</h2>`)
}

func (c Context) interactablesNewHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		getInteractablesNew(w, r)
	}
}

func getInteractablesNew(w http.ResponseWriter, _ *http.Request) {
	err := tmpl.ExecuteTemplate(w, "interactables-new", nil)
	if err != nil {
		fmt.Println(err)
	}
}

func (c Context) interactableNewHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		getInteractableNew(w, r)
	}
}

func getInteractableNew(w http.ResponseWriter, r *http.Request) {
	queryValues := r.URL.Query()
	setName := queryValues.Get("interactable-set")
	var pageData = struct {
		SetName string
	}{
		SetName: setName,
	}
	err := tmpl.ExecuteTemplate(w, "interactable-new", pageData)
	if err != nil {
		fmt.Println(err)
	}
}

func (c *Context) interactableHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		c.getInteractable(w, r)
	}
	if r.Method == "POST" {
		c.postInteractable(w, r)
	}
	if r.Method == "PUT" {
		c.putInteractable(w, r)
	}
}

func (c *Context) getInteractable(w http.ResponseWriter, r *http.Request) {
	queryValues := r.URL.Query()
	collectionName := queryValues.Get("currentCollection")
	id := queryValues.Get("interactable")
	collection, ok := c.Collections[collectionName]
	if !ok {
		panic("Help plz")
	}
	interactable := collection.findInteractableById(id)
	if interactable == nil {
		panic("Invalid interactable id")
	}

	err := tmpl.ExecuteTemplate(w, "interactable-edit", interactable)
	if err != nil {
		fmt.Println(err)
	}
}

func (c Context) putInteractable(w http.ResponseWriter, r *http.Request) {
	fmt.Println("PUT for /interactable")

	properties, _ := requestToProperties(r)
	collectionName := properties["currentCollection"]
	setName := properties["interactable-set"]
	collection, ok := c.Collections[collectionName]
	if !ok {
		panic("no collection")
	}
	_, ok = collection.InteractableSets[setName]
	if !ok {
		panic("no set")
	}

	id := properties["interactable-id"]
	name := properties["Name"]
	cssClass := properties["CssClass"]
	pushable := (properties["pushable"] == "on")
	fragile := (properties["fragile"] == "on")
	reactions := properties["reactions"]

	fmt.Printf("%s | Pushable: %t - Fragile: %t - CSS: %s - Reactions %s\n", name, pushable, fragile, cssClass, reactions)
	panicIfAnyEmpty("Invalid interactable", id, name)

	interactable := collection.findInteractableById(id)
	if interactable == nil {
		panic("no proto with that id")
	}
	interactable.Name = name
	interactable.CssClass = cssClass
	interactable.Pushable = pushable
	interactable.Fragile = fragile
	interactable.Reactions = reactions

	//fmt.Println(interactable)
	collection.saveInteractableSet(setName)

	io.WriteString(w, "<h3>Done.</h3>")
}

func (c *Context) postInteractable(w http.ResponseWriter, r *http.Request) {
	fmt.Println("POST for /interactable")

	properties, _ := requestToProperties(r)
	collectionName := properties["currentCollection"]
	setName := properties["interactable-set"]
	collection, ok := c.Collections[collectionName]
	if !ok {
		panic("no collection")
	}
	set, ok := collection.InteractableSets[setName]
	if !ok {
		panic("no set")
	}

	name := properties["Name"]
	cssClass := properties["CssClass"]
	pushable := (properties["pushable"] == "on")
	fragile := (properties["fragile"] == "on")
	reactions := properties["reactions"]

	fmt.Printf("%s | Pushable: %t - Fragile: %t - %s\n", name, pushable, fragile, cssClass)
	panicIfAnyEmpty("Invalid interactable", name)

	id := uuid.New().String()
	collection.InteractableSets[setName] = append(set,
		InteractableDescription{
			ID:        id,
			SetName:   setName,
			Name:      name,
			CssClass:  cssClass,
			Pushable:  pushable,
			Fragile:   fragile,
			Reactions: reactions,
		})

	collection.saveInteractableSet(setName)

	io.WriteString(w, "<h3>Done.</h3>")
}

func exampleInteractable(w http.ResponseWriter, r *http.Request) {
	cssClass := r.URL.Query().Get("CssClass")

	output := fmt.Sprintf(`<div id="exampleSquare" class="grid-row"><div class="grid-square %s"></div></div>`, cssClass)
	io.WriteString(w, output)
}
