package main

import (
	"fmt"
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
	} else {
		fmt.Fprintf(w, "Invalid structure type (%s) ", structureType)
	}
}
