package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

type Transport struct {
	SourceY            int    `json:"sourceY"`
	SourceX            int    `json:"sourceX"`
	DestY              int    `json:"destY"`
	DestX              int    `json:"destX"`
	DestStage          string `json:"destStage"`
	Confirmation       bool   `json:"confirmation"`
	RejectInteractable bool   `json:"rejectInteractable"`
}

func (c *Context) getEditTransports(w http.ResponseWriter, r *http.Request) {
	queryValues := r.URL.Query()
	name := queryValues.Get("area-name")
	collectionName := queryValues.Get("currentCollection")
	collection := c.Collections[collectionName]
	if collection == nil {
		panic("ooo spooky")
	}
	spaceName := queryValues.Get("currentSpace")
	space := c.spaceFromNames(collectionName, spaceName)
	selectedArea := getAreaByName(space.Areas, name)
	if selectedArea == nil {
		io.WriteString(w, "<h2>no Area</h2>")
		return
	}

	err := tmpl.ExecuteTemplate(w, "transport-form", selectedArea)
	if err != nil {
		fmt.Println(err)
	}
	highlightSelects := collection.transportsAsOob(*selectedArea, spaceName)
	io.WriteString(w, highlightSelects)
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

	confirmation := (properties["confirmation"] == "on")
	rejectInteractable := (properties["reject-interactable"] == "on")

	collectionName := properties["currentCollection"]
	spaceName := properties["currentSpace"]
	space := c.spaceFromNames(collectionName, spaceName)
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
	currentTransport.Confirmation = confirmation
	currentTransport.RejectInteractable = rejectInteractable

	err := tmpl.ExecuteTemplate(w, "transport-form", selectedArea)
	if err != nil {
		fmt.Println(err)
	}
}

func (c Context) newTransport(w http.ResponseWriter, r *http.Request) {
	properties, _ := requestToProperties(r)
	areaName := properties["area-name"]
	fmt.Println(areaName)

	collectionName := properties["currentCollection"]
	spaceName := properties["currentSpace"]
	space := c.spaceFromNames(collectionName, spaceName)
	selectedArea := getAreaByName(space.Areas, areaName)
	if selectedArea == nil {
		io.WriteString(w, "<h2>no Area</h2>")
		return
	}

	selectedArea.Transports = append(selectedArea.Transports, Transport{})

	err := tmpl.ExecuteTemplate(w, "transport-form", selectedArea)
	if err != nil {
		fmt.Println(err)
	}
}

func (c Context) dupeTransport(w http.ResponseWriter, r *http.Request) {
	properties, _ := requestToProperties(r)
	transportId, _ := strconv.Atoi(properties["transport-id"])
	areaName := properties["transport-area-name"]

	collectionName := properties["currentCollection"]
	spaceName := properties["currentSpace"]
	space := c.spaceFromNames(collectionName, spaceName)
	selectedArea := getAreaByName(space.Areas, areaName)
	if selectedArea == nil {
		io.WriteString(w, "<h2>no Area</h2>")
		return
	}

	currentTransport := &selectedArea.Transports[transportId]
	newTransport := *currentTransport
	selectedArea.Transports = append(selectedArea.Transports, newTransport)

	err := tmpl.ExecuteTemplate(w, "transport-form", selectedArea)
	if err != nil {
		fmt.Println(err)
	}
}

func (c Context) deleteTransport(w http.ResponseWriter, r *http.Request) {
	properties, _ := requestToProperties(r)
	id, _ := strconv.Atoi(properties["transport-id"])
	areaName := properties["transport-area-name"]

	collectionName := properties["currentCollection"]
	spaceName := properties["currentSpace"]
	space := c.spaceFromNames(collectionName, spaceName)
	selectedArea := getAreaByName(space.Areas, areaName)
	if selectedArea == nil {
		io.WriteString(w, "<h2>no Area</h2>")
		return
	}

	selectedArea.Transports = append(selectedArea.Transports[:id], selectedArea.Transports[id+1:]...)
	fmt.Println(len(selectedArea.Transports))

	// Still need to remove highlights for deleted transports, and new?
	err := tmpl.ExecuteTemplate(w, "transport-form", selectedArea)
	if err != nil {
		fmt.Println(err)
	}
}

func (col *Collection) transportsAsOob(area AreaDescription, spacename string) string {
	var buf bytes.Buffer
	for _, transport := range area.Transports {
		tile := area.Blueprint.Tiles[transport.SourceY][transport.SourceX]
		event := GridClickDetails{
			Y:                transport.SourceY,
			X:                transport.SourceX,
			GridType:         "area",
			ScreenID:         "screen",
			DefaultTileColor: area.Blueprint.DefaultTileColor, // needs ground e.g (defaultBasedOnGroundCell func)
			Selected:         true,
			Location:         []string{spacename, area.Name},
		}
		col.executeGridSquareTemplate(&buf, event, tile)
	}
	return buf.String()
}

func (col *Collection) executeGridSquareTemplate(w io.Writer, event GridClickDetails, tile TileData) {
	var pageData = struct {
		Material     Material
		ClickEvent   GridClickDetails
		Interactable *InteractableDescription
	}{
		Material:     col.findPrototypeById(tile.PrototypeId).applyTransformForEditor(tile.Transformation),
		Interactable: col.findInteractableById(tile.InteractableId),
		ClickEvent:   event,
	}
	err := tmpl.ExecuteTemplate(w, "grid-square", pageData)
	if err != nil {
		fmt.Println(err)
	}
}
