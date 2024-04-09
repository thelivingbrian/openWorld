package main

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"strconv"
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
		<label>Space: <a hx-get="/spaces/new" hx-include="[name='currentCollection']" hx-target="#space_select" href="#">New</a></label>
		<select name="currentSpace" hx-get="/areas" hx-include="[name='currentCollection']" hx-target="#area_select">
			<option value=""></option>
			{{range  $key, $value := .Spaces}}
				<option value="{{$key}}">{{$key}}</option>
			{{end}}
		</select>
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
		<label><b>Space Name: </b></label>
		<input type="text" name="newSpaceName" /><br />
		<br />
		<label><b>Latitude: </b></label>
		<input type="text" name="latitude" /><br />
		
		<label><b>Longitude: </b></label>
		<input type="text" name="longitude" /><br />

		<label><b>Topology: </b></label><br />
		<span><input type="radio" name="topology" value="plane" checked />Plane</span><br />
		<span><input type="radio" name="topology" value="Torus" />Torus</span><br />
		<br /> 

		</label><b>Area Dimensions</b></label><br />
		<label>Width : </label><input type="text" name="areaWidth" value="16"/>
		<label>Height : </label><input type="text" name="areaHeight" value="16" /><br />
		
		<label>Default Tile Color : </label>
		<input type="text" name="tileColor" /><br />

		<input type="submit" />
	</form>
</div>`

var spaceTmpl = template.Must(template.New("SpacePage").Parse(divNewSpace))

func (c Context) spacesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		c.getSpaces(w, r)
	}
	if r.Method == "POST" {
		c.postSpaces(w, r)
	}
}

func (c Context) getSpaces(w http.ResponseWriter, r *http.Request) {
	queryValues := r.URL.Query()
	name := queryValues.Get("collectionName")
	if col, ok := c.Collections[name]; ok {
		err := spacesTmpl.Execute(w, col)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func (c Context) postSpaces(w http.ResponseWriter, r *http.Request) {
	props, ok := requestToProperties(r)
	if !ok {
		fmt.Println("Invalid POST to /spaces. Properties are invalid.")
		io.WriteString(w, `<h3> Properties are invalid. </h3>`)
		return
	}
	cName, ok := props["currentCollection"]
	if !ok {
		fmt.Println("Invalid POST to /spaces. Collection not found.")
		io.WriteString(w, `<h3> Collection not found. </h3>`)
		return
	}
	if col, ok := c.Collections[cName]; ok {
		fmt.Println(col.Name)
		valid := true

		name, ok := props["newSpaceName"]
		valid = valid && ok

		lat, ok := props["latitude"]
		valid = valid && ok

		long, ok := props["longitude"]
		valid = valid && ok

		topology, ok := props["topology"]
		valid = valid && ok

		areaWidth, ok := props["areaWidth"]
		valid = valid && ok

		areaHeight, ok := props["areaHeight"]
		valid = valid && ok

		tileColor, ok := props["tileColor"]
		valid = valid && ok
		if !valid {
			fmt.Println("Invalid, failed to get properties by name.")
			io.WriteString(w, `<h3> Properties are invalid.</h3>`)
			return
		}

		latitude, err := strconv.Atoi(lat)
		valid = valid && (err == nil)

		longitude, err := strconv.Atoi(long)
		valid = valid && (err == nil)

		width, err := strconv.Atoi(areaWidth)
		valid = valid && (err == nil)

		height, err := strconv.Atoi(areaHeight)
		valid = valid && (err == nil)
		if !valid {
			fmt.Println(err)
			fmt.Println("Invalid, failed to cast lat/long/h/w.")
			io.WriteString(w, `<h3> Failed to cast. </h3>`)
			return
		}

		fmt.Printf("%s %s %s %s %s %d %d", name, topology, areaWidth, areaHeight, tileColor, latitude, longitude)

		area := createBaseArea(height, width, tileColor)
		space := createSpace(cName, name, latitude, longitude, topology, area)
		col.Spaces[name] = &space
		io.WriteString(w, `<h3>Success</h3>`)
		return
	}
	io.WriteString(w, `<h3>Invalid collection Name.</h3>`)
}

func createSpace(cName, name string, latitude, longitude int, topology string, area Area) Space {
	areas := make([][]Area, latitude)
	for y := 0; y < latitude; y++ {
		for x := 0; x < longitude; x++ {
			// This is consistent with Tiles
			area.Name = fmt.Sprintf("%s:%d-%d", name, y, x)
			area.North = fmt.Sprintf("%s:%d-%d", name, mod(y-1, latitude), x)
			area.South = fmt.Sprintf("%s:%d-%d", name, mod(y+1, latitude), x)
			area.East = fmt.Sprintf("%s:%d-%d", name, y, mod(x+1, longitude))
			area.West = fmt.Sprintf("%s:%d-%d", name, y, mod(x-1, longitude))
			areas[y] = append(areas[y], area)
		}
	}
	// Remove edges if plane topology

	flatAreas := make([]Area, 0)
	for i := range areas {
		flatAreas = append(flatAreas, areas[i]...)
	}

	out := Space{CollectionName: cName, Name: name, Areas: flatAreas}
	return out
}

func mod(i, n int) int {
	return ((i % n) + n) % n
}

func createBaseArea(height, width int, tileColor string) Area {
	tiles := make([][]int, height)
	for i := range tiles {
		tiles[i] = make([]int, width)
	}
	return Area{Name: "", Safe: true, Tiles: tiles, Transports: make([]Transport, 0), DefaultTileColor: tileColor}
}

func getAreaByName(areas []Area, name string) *Area {
	for i, area := range areas {
		if name == area.Name {
			return &areas[i]
		}
	}
	return nil
}

func getFragmentByName(fragments []Fragment, name string) *Fragment {
	for i, fragment := range fragments {
		if name == fragment.Name {
			return &fragments[i]
		}
	}
	return nil
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
