package main

import (
	"fmt"
	"html/template"
	"net/http"
)

type Space struct {
	CollectionName string
	Name           string
	Areas          []Area
}

var divSpacePage = `
<div id="space_page">
	<div id="collection_header">
		<input type="hidden" name="currentCollection" value="{{.Name}}" />
		<span>
			<b>Collection:</b>  {{.Name}}  
			<b>(<a hx-get="/deploy" hx-include="[name='currentCollection']" hx-target="#panel" href="#">Deploy</a>)</b>
		</span>
	</div>
	<br />
	<div id="space_select">
		<label><b>Select Space: </b></label>
		<select name="spaceName" hx-get="/areas" hx-include="[name='currentCollection']" hx-target="#area_select">
			<option value=""></option>
			{{range  $key, $value := .Spaces}}
				<option value="{{$key}}">{{$key}}</option>
			{{end}}
		</select>
		<button class="btn" hx-get="/spaces/new" hx-include="[name='currentCollection']" hx-target="#space_select">New</button>
		<br />
		<div id="area_select">

		</div>
	</div>
	<br />
</div>`
var spacesTmpl = template.Must(template.New("SpacesPage").Parse(divSpacePage))

var divNewSpace = `
<div id="space_new">
	<form hx-post="/spaces">
		<input type="hidden" name="currentCollection" value="{{.Name}}" />
		
		<h3>Create New Space</h3>
		<label>Space Name: <label>
		<input type="text" name="newSpaceName" /><br />
		
		<label>Latitude: <label>
		<input type="text" name="latitude" /><br />
		
		<label>Longitude: <label>
		<input type="text" name="longitude" /><br />
		
		<label>Topology: <label><br />
		<span><input type="radio" name="topology" value="plane" />Plane</span><br />
		<span><input type="radio" name="topology" value="Torus" />Torus</span><br />
		<br />

		</label>Area Dimensions</label><br />
		<label>Width : </label><input type="text" name="areaWidth" value="16"/>
		<label>Height : </label><input type="text" name="areaHeight" value="16" /><br />
		
		<label>Default Tile Color : <label>
		<input type="text" name="tileColor" /><br />

		<input type="submit" />
	</form>
</div>`

var spaceTmpl = template.Must(template.New("SpacePage").Parse(divNewSpace))

func (c Context) spacesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		queryValues := r.URL.Query()
		name := queryValues.Get("collectionName")
		if col, ok := c.Collections[name]; ok {
			//fmt.Println(col.Name + "a")
			col.getSpaces(w, r)
		}
	}
	if r.Method == "POST" {
		props, ok := requestToProperties(r)
		if ok {
			fmt.Println(len(props))
		}
	}
}

func (col Collection) getSpaces(w http.ResponseWriter, r *http.Request) {
	err := spacesTmpl.Execute(w, col)
	if err != nil {
		fmt.Println(err)
	}
}

//

func (c Context) newSpaceHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		queryValues := r.URL.Query()
		name := queryValues.Get("currentCollection")
		fmt.Println("Collection Name: " + name)
		if col, ok := c.Collections[name]; ok {
			col.getNewSpace(w, r)
		}
	}
}

func (col Collection) getNewSpace(w http.ResponseWriter, r *http.Request) {
	err := spaceTmpl.Execute(w, col)
	if err != nil {
		fmt.Println(err)
	}
}

// Where does this belong?
func (c Context) getSpace(collectionName string, spaceName string) *Space {
	return &Space{CollectionName: collectionName, Name: spaceName, Areas: c.Collections[collectionName].Spaces[spaceName]}
}
