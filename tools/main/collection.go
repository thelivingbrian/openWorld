package main

import (
	"fmt"
	"html/template"
	"net/http"
)

type Fragment struct {
	width  int
	height int
	tiles  [][]int
}

type Space struct {
	CollectionName string
	Name           string
	Areas          []Area
}

type Collection struct {
	Name       string
	SpaceNames []string
	Spaces     map[string][]Area
	Fragments  map[string][]Fragment
	//hyttakenSpaceNames map[string]bool
	// last modified etc
}

var collectionsPage = `
<div id="collection_page"> 
	<div id="collection_select">
		<h3>Select Existing Collection: </h3>
		<select name="collectionName" hx-get="/spaces" hx-target="#collection_page">
		<option value=""></option>
		{{range  $key, $value := .}}
			<option value="{{$key}}">{{$key}}</option>
		{{end}}
		</select>
	</div> 
	<div id="collection_new">
		<h3>Create New Colletion</h3>
		<form>
			<label>Collection Name: <label>
			<input type="text" name="newCollectionName" /><br />
			
			<label>Space Name: <label>
			<input type="text" name="newSpaceName" /><br />
			
			<label>Initial Space Latitude: <label>
			<input type="text" name="latitude" /><br />
			
			<label>Intitial Space Longitude: <label>
			<input type="text" name="longitude" /><br />
			
			<label>Topology: <label>
			<input type="radio" name="Plane" />	
			<input type="radio" name="Torus" />
			<br />

			</label>Area Dimensions</label><br />
			<label>Width : </label><input type="text" name="areaWidth" />
			<label>Height : </label><input type="text" name="areaHeight" /><br />
			
			<label>Default Tile Color : <label>
			<input type="text" name="tileColor" /><br />
		</form>
	</div>
</div>
`

var collectionsTmpl = template.Must(template.New("CollectionPage").Parse(collectionsPage))

// Rename collectionHandler handler
func (c Context) collectionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		c.getCollections(w, r)
	}

}

func (c Context) getCollections(w http.ResponseWriter, r *http.Request) {
	collectionsTmpl.Execute(w, c.Collections)
}

// Remove
/*
func planesMake(w http.ResponseWriter, r *http.Request) {
	err := os.Mkdir(filepath.Join("./planes", "new"), 0755)
	if err != nil {
		fmt.Println(err)
	}

	dirs, err := os.ReadDir("./planes")
	if err != nil {
		fmt.Println(err)
	}

	for _, dir := range dirs {
		x, _ := dir.Info()
		fmt.Println(x.Name())
		fmt.Println(dir.IsDir())
	}
}
*/
/// Spaces

var divSpacePage = `
<div id="space_page">
	<input type="hidden" name="currentCollection" value="{{.Name}}" />

	<div id="space_select">
		<label><b>Select Space: </b></label>
		<select name="spaceName" hx-get="/areas" hx-include="[name='currentCollection']" hx-target="#area_select">
		<option value=""></option>
		{{range  $key, $value := .Spaces}}
			<option value="{{$key}}">{{$key}}</option>
		{{end}}
		</select>
		<br />
		<div id="area_select">

		</div>
	</div> 
</div>`

var newSpace = `
<div id="space_new">
<h3>Create New Space</h3>
<form>
	<label>Space Name: <label>
	<input type="text" name="newSpaceName" /><br />
	
	<label>Initial Space Latitude: <label>
	<input type="text" name="latitude" /><br />
	
	<label>Intitial Space Longitude: <label>
	<input type="text" name="longitude" /><br />
	
	<label>Topology: <label>
	<input type="radio" name="Plane" />	
	<input type="radio" name="Torus" />
	<br />

	</label>Area Dimensions</label><br />
	<label>Width : </label><input type="text" name="areaWidth" />
	<label>Height : </label><input type="text" name="areaHeight" /><br />
	
	<label>Default Tile Color : <label>
	<input type="text" name="tileColor" /><br />
</form>
</div>`

var spaceTmpl = template.Must(template.New("SpacePage").Parse(divSpacePage))

func (c Context) spacesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		queryValues := r.URL.Query()
		name := queryValues.Get("collectionName")
		fmt.Println(name)
		if col, ok := c.Collections[name]; ok {
			fmt.Println(col.Name + "a")
			col.getSpaces(w, r)
		}
	}

}

func (col Collection) getSpaces(w http.ResponseWriter, r *http.Request) {
	err := spaceTmpl.Execute(w, col)
	if err != nil {
		fmt.Println(err)
	}
}

/*
var divAreaSelect = `
	<h3>Select Area: </h3>
	<input type="hidden" name="currentSpace" value="{{.Name}}" />
	<select name="areaName" hx-get="/area" hx-target="#collection_page">
		<option value=""></option>
	{{range  $i, $area := .Areas}}
		<option value="{{$area.Name}}">{{$area.Name}}</option>
	{{end}}
	</select>

`

var areaSelectTmpl = template.Must(template.New("AreaSelect").Parse(divAreaSelect))
*/

func (c Context) areasHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		queryValues := r.URL.Query()
		collectionName := queryValues.Get("currentCollection")
		spaceName := queryValues.Get("spaceName")
		s := c.getSpace(collectionName, spaceName)
		//fmt.Println("Final space: " + s.Name)
		//fmt.Println(len(s.Areas))
		getAreaPage(w, r, s)
	}
}

func (c Context) getSpace(collectionName string, spaceName string) *Space {
	fmt.Println("Spaces to choose:  ")
	fmt.Println(len(c.Collections[collectionName].Spaces))
	fmt.Println("Areas found in selected space: ")
	fmt.Println(len(c.Collections[collectionName].Spaces[spaceName]))
	return &Space{CollectionName: collectionName, Name: spaceName, Areas: c.Collections[collectionName].Spaces[spaceName]}
}

/*
func (c Context) areaHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		queryValues := r.URL.Query()
		collectionName := queryValues.Get("currentCollection")
		spaceName := queryValues.Get("currentSpace")
		areaName := queryValues.Get("areaName")
		fmt.Println("area collection name: " + collectionName)
		fmt.Println("area space name: " + spaceName)
		fmt.Println("area name: " + areaName)
	}
}
*/
