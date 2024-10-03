package main

import (
	"fmt"
	"net/http"
	"strconv"
)

type Structure struct {
	ID                            string
	FragmentHeight, FragmentWidth int
	FragmentIds                   [][]string
}

func (c Context) structureHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		c.postStructure(w, r)
	}
}

func (c Context) postStructure(w http.ResponseWriter, r *http.Request) {
	properties, _ := requestToProperties(r)
	collectionName := properties["currentCollection"]
	groundName := properties["ground-name"]
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
	structureType := properties["structure-type"]

	// Panic if any required field is missing
	panicIfAnyEmpty("POST to /structure", collectionName, groundName, span, color1, color2, fuzz, strategy)

	// Process the data as necessary for ground generation
	// Add your logic for ground generation here, e.g., saving the structure
	if structureType == "ground" {
		fmt.Fprintf(w, "New ground generation initiated with the following details: Name: %s, Grid Size: %d, Colors: %s and %s, Fuzz: %f, Strategy: %s", groundName, spanI, color1, color2, fuzzF, strategy)
		col := c.Collections[collectionName]
		col.generateAndSaveGroundPattern(groundName, color1, color2, spanI, strategy, fuzzF)
	} else {

		fmt.Fprintf(w, "Invalid structure type (%s) with following details: Name: %s, Grid Size: %d, Colors: %s and %s, Fuzz: %f, Strategy: %s", structureType, groundName, spanI, color1, color2, fuzzF, strategy)
	}
}
