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
	Name      string
	Spaces    map[string][]Area
	Fragments map[string][]Fragment
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

		</form>
	</div>
</div>
`

var collectionsTmpl = template.Must(template.New("CollectionPage").Parse(collectionsPage))

func (c Context) collectionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		c.getCollections(w, r)
	}
}

func (c Context) getCollections(w http.ResponseWriter, r *http.Request) {
	collectionsTmpl.Execute(w, c.Collections)
}

/// Spaces

var divSpacePage = `
<div id="space_page">
	<div id="space_select">
		<input type="hidden" name="currentCollection" value="{{.Name}}" />
		<span>
			<b>Collection:</b>  {{.Name}}  
			<b>(<a hx-get="/deploy" hx-include="[name='currentCollection']" hx-target="#panel" href="#">Deploy</a>)</b>
		</span>
		<br />
	</div>
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
	<input type="hidden" name="currentCollection" value="INSERT" />
	
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

func (c Context) getSpace(collectionName string, spaceName string) *Space {
	return &Space{CollectionName: collectionName, Name: spaceName, Areas: c.Collections[collectionName].Spaces[spaceName]}
}
