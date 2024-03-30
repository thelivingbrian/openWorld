package main

import (
	"html/template"
	"net/http"
)

type Fragment struct {
	width  int
	height int
	tiles  [][]int
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
