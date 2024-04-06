package main

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
)

func (c Context) getEditTransports(w http.ResponseWriter, r *http.Request) {
	queryValues := r.URL.Query()
	name := queryValues.Get("area-name")
	collectionName := queryValues.Get("currentCollection")
	spaceName := queryValues.Get("currentSpace")
	space := c.getSpace(collectionName, spaceName)
	selectedArea := getAreaByName(space.Areas, name)
	if selectedArea == nil {
		io.WriteString(w, "<h2>no Area</h2>")
		return
	}

	output := transportFormHtml(*selectedArea)
	output += transportsAsOob(*selectedArea)
	io.WriteString(w, output)
}

func (c Context) editTransport(w http.ResponseWriter, r *http.Request) {
	properties, _ := requestToProperties(r)
	transportId, _ := strconv.Atoi(properties["transport-id"])
	destStage := properties["transport-stage-name"]
	destY, _ := strconv.Atoi(properties["transport-dest-y"])
	destX, _ := strconv.Atoi(properties["transport-dest-x"])
	sourceY, _ := strconv.Atoi(properties["transport-source-y"])
	sourceX, _ := strconv.Atoi(properties["transport-source-x"])
	areaName := properties["transport-area-name"]

	collectionName := properties["currentCollection"]
	spaceName := properties["currentSpace"]
	space := c.getSpace(collectionName, spaceName)
	selectedArea := getAreaByName(space.Areas, areaName)
	if selectedArea == nil {
		io.WriteString(w, "<h2>no Area</h2>")
		return
	}

	currentTransport := &selectedArea.Transports[transportId]
	currentTransport.DestY = destY
	currentTransport.DestX = destX
	currentTransport.SourceY = sourceY
	currentTransport.SourceX = sourceX
	currentTransport.DestStage = destStage

	output := transportFormHtml(*selectedArea)
	io.WriteString(w, output)
}

func (c Context) newTransport(w http.ResponseWriter, r *http.Request) {
	properties, _ := requestToProperties(r)
	areaName := properties["area-name"]
	fmt.Println(areaName)

	collectionName := properties["currentCollection"]
	spaceName := properties["currentSpace"]
	space := c.getSpace(collectionName, spaceName)
	selectedArea := getAreaByName(space.Areas, areaName)
	if selectedArea == nil {
		io.WriteString(w, "<h2>no Area</h2>")
		return
	}

	selectedArea.Transports = append(selectedArea.Transports, Transport{})

	output := transportFormHtml(*selectedArea)
	io.WriteString(w, output)

}

func editTransportForm(i int, t Transport, sourceName string) string {
	output := fmt.Sprintf(`
	<form hx-post="/editTransport" hx-target="#edit_transports" hx-swap="outerHTML">
		<input type="hidden" name="transport-id" value="%d" />
		<input type="hidden" name="transport-area-name" value="%s" />
		<table>
			<tr>
				<td align="right">Dest stage-name:</td>
				<td align="left">
					<input type="text" name="transport-stage-name" value="%s" />
				</td>
			</tr>
			<tr>
				<td align="right">Dest y</td>
				<td align="left">
					<input type="text" name="transport-dest-y" value="%d" />
				</td>
				<td align="right">x</td>
				<td align="left">
					<input type="text" name="transport-dest-x" value="%d" />
				</td>
			</tr>
			<tr>
				<td align="right">Source y</td>
				<td align="left">
					<input type="text" name="transport-source-y" value="%d" />
				</td>
				<td align="right">x</td>
				<td align="left">
					<input type="text" name="transport-source-x" value="%d" />
				</td>
			</tr>
			<tr>
				<td align="right">Css-class:</td>
				<td align="left">
					<input type="text" name="transport-css-class" value="%s" />
				</td>
			<tr />
		</table>

		<button class="btn">Submit</button>
		<button class="btn" hx-post="/dupeTransport" hx-include="[name='area-name'],[name='currentCollection'],[name='currentSpace']">Duplicate</button>
		<button class="btn" hx-post="/deleteTransport" hx-include="[name='area-name'],[name='currentCollection'],[name='currentSpace']">Delete</button>
	</form>`, i, sourceName, t.DestStage, t.DestY, t.DestX, t.SourceY, t.SourceX, "pink")
	return output
}

func transportFormHtml(area Area) string {
	output := `<div id="edit_transports">
					<h4>Transports: </h4>
					<a hx-post="/newTransport" hx-include="[name='area-name'],[name='currentCollection'],[name='currentSpace']" hx-target="#edit_transports" href="#"> New </a><br />`
	for i, transport := range area.Transports {
		output += editTransportForm(i, transport, area.Name)
	}
	output += `</div>`
	return output
}

func (c Context) dupeTransport(w http.ResponseWriter, r *http.Request) {
	properties, _ := requestToProperties(r)
	transportId, _ := strconv.Atoi(properties["transport-id"])
	areaName := properties["transport-area-name"]

	collectionName := properties["currentCollection"]
	spaceName := properties["currentSpace"]
	space := c.getSpace(collectionName, spaceName)
	selectedArea := getAreaByName(space.Areas, areaName)
	if selectedArea == nil {
		io.WriteString(w, "<h2>no Area</h2>")
		return
	}

	currentTransport := &selectedArea.Transports[transportId]
	newTransport := *currentTransport
	selectedArea.Transports = append(selectedArea.Transports, newTransport)

	output := transportFormHtml(*selectedArea)
	io.WriteString(w, output)
}

func (c Context) deleteTransport(w http.ResponseWriter, r *http.Request) {
	properties, _ := requestToProperties(r)
	id, _ := strconv.Atoi(properties["transport-id"])
	areaName := properties["transport-area-name"]

	collectionName := properties["currentCollection"]
	spaceName := properties["currentSpace"]
	space := c.getSpace(collectionName, spaceName)
	selectedArea := getAreaByName(space.Areas, areaName)
	if selectedArea == nil {
		io.WriteString(w, "<h2>no Area</h2>")
		return
	}

	selectedArea.Transports = append(selectedArea.Transports[:id], selectedArea.Transports[id+1:]...)
	fmt.Println(len(selectedArea.Transports))

	output := transportFormHtml(*selectedArea)
	// Remove highlight for deleted transport
	io.WriteString(w, output)
}

func transportsAsOob(area Area) string {
	output := ``
	for _, transport := range area.Transports {
		var yStr = strconv.Itoa(transport.SourceY)
		var xStr = strconv.Itoa(transport.SourceX)
		output += `<div hx-swap-oob="true" hx-post="/clickOnSquare" hx-trigger="click" hx-include="[name='radio-tool'],[name='selected-material']" hx-headers='{"y": "` + yStr + `", "x": "` + xStr + `"}' class="grid-square ` + modifications[transport.SourceY][transport.SourceX].CssColor + `" id="c` + yStr + `-` + xStr + `"><div class="box top med red-b"></div></div></div>`
	}
	output += ``
	return output
}
