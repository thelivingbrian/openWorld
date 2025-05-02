package main

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
)

type Structure struct {
	ID           string        `json:"id"`
	FragmentIds  [][]string    `json:"fragmentIds"`
	GroundConfig *GroundConfig `json:"groundConfig,omitempty"`
}

type GroundConfig struct {
	Name     string  `json:"name"`
	Span     int     `json:"span"`
	Color1   string  `json:"color1"`
	Color2   string  `json:"color2"`
	Fuzz     float64 `json:"fuzz"`
	Strategy string  `json:"strategy"`
}

func (c Context) structureHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		c.postStructure(w, r)
	}
	if r.Method == "DELETE" {
		c.deleteStructure(w, r)
	}
	if r.Method == "PUT" {
		c.putStructure(w, r)
	}
}

func (c Context) postStructure(w http.ResponseWriter, r *http.Request) {
	properties, _ := requestToProperties(r)
	structureType := properties["structure-type"]

	panicIfAnyEmpty("POST to /structure", structureType)

	if structureType == "ground" {
		collectionName := properties["currentCollection"]
		structureName := properties["structure-name"]
		span := properties["span"]
		spanI, err := strconv.Atoi(span)
		if err != nil {
			panic("Invalid grid size")
		}
		color1 := properties["color1"]
		color2 := properties["color2"]
		fuzz := properties["fuzz"]
		fuzzF, err := strconv.ParseFloat(fuzz, 64)
		if err != nil {
			panic("Invalid fuzz value")
		}
		strategy := properties["strategy"]

		col, ok := c.Collections[collectionName]
		if !ok {
			panic("Invalid collection.")
		}
		fmt.Fprintf(w, "New ground generation initiated with the following details: Name: %s, Grid Size: %d, Colors: %s and %s, Fuzz: %f, Strategy: %s", structureName, spanI, color1, color2, fuzzF, strategy)
		config := GroundConfig{Name: structureName, Span: spanI, Color1: color1, Color2: color2, Fuzz: fuzzF, Strategy: strategy}
		col.generateAndSaveGroundPattern(config) // Two methods?
	} else if structureType == "pattern" {

	} else {
		fmt.Fprintf(w, "Invalid structure type (%s) ", structureType)
	}
}

func (c Context) deleteStructure(w http.ResponseWriter, r *http.Request) {
	properties, _ := requestToProperties(r)
	collectionName := properties["currentCollection"]
	structureId := properties["structureId"]
	structureType := properties["structure-type"]

	col, ok := c.Collections[collectionName]
	if !ok {
		io.WriteString(w, "<div><h2>Invalid collection.</h2></div>")
		return
	}
	grounds, ok := col.StructureSets[structureType]
	if !ok {
		io.WriteString(w, "<div><h2>No structure type.</h2></div>")
		return
	}

	fmt.Println("Deleting: " + structureId)

	col.StructureSets[structureType] = removeStructuresById(grounds, structureId)

	// bad
	outFileStruct := "./data/collections/bloop/structures/ground.json"
	err := writeJsonFile(outFileStruct, col.StructureSets["ground"], false)
	if err != nil {
		panic(err)
	}
}

func (c Context) putStructure(w http.ResponseWriter, r *http.Request) {
	properties, _ := requestToProperties(r)
	collectionName := properties["currentCollection"]
	structureId := properties["structureId"]
	structureType := properties["structure-type"]

	col, ok := c.Collections[collectionName]
	if !ok {
		io.WriteString(w, "<div><h2>Invalid collection.</h2></div>")
		return
	}
	structures, ok := col.StructureSets[structureType]
	if !ok {
		io.WriteString(w, "<div><h2>No structure type.</h2></div>")
		return
	}

	structure, found := getStructureById(structures, structureId)
	if !found {
		io.WriteString(w, "<div><h2 class=\"dangerous\">no structure with that id</h2></div>")
		return
	}
	if structureType == "ground" {
		if structure.GroundConfig == nil {
			panic("unable to regenerate without config")
		}
		col.generateAndSaveGroundPattern(*structure.GroundConfig)
	}
}

func removeStructuresById(structures []Structure, id string) []Structure {
	out := make([]Structure, 0)
	for i := range structures {
		if structures[i].ID != id {
			out = append(out, structures[i])
		}
	}
	return out
}

func getStructureById(structures []Structure, id string) (Structure, bool) {
	for i := range structures {
		if structures[i].ID == id {
			return structures[i], true
		}
	}
	return Structure{}, false
}
