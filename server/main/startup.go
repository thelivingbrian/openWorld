package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type Material struct {
	ID           int    `json:"id"`
	CommonName   string `json:"commonName"`
	CssClassName string `json:"cssClassName"`
	Walkable     bool   `json:"walkable"`
	R            int    `json:"R"`
	G            int    `json:"G"`
	B            int    `json:"B"`
}

type Transport struct {
	SourceY   int    `json:"sourceY"`
	SourceX   int    `json:"sourceX"`
	DestY     int    `json:"destY"`
	DestX     int    `json:"destX"`
	DestStage string `json:"destStage"`
}

type Area struct {
	Name       string      `json:"name"`
	Safe       bool        `json:"safe"`
	Tiles      [][]int     `json:"tiles"`
	Transports []Transport `json:"transports"`
}

var (
	materials []Material
	areas     []Area
)

func populateStructUsingFileName[T any](ptr *T, fn string) {
	jsonData, err := os.ReadFile(fmt.Sprintf("./data/%s.json", fn))
	if err != nil {
		panic(err)
	}

	if err := json.Unmarshal(jsonData, ptr); err != nil {
		panic(err)
	}
}

func loadFromJson() {
	populateStructUsingFileName[[]Material](&materials, "materials")
	populateStructUsingFileName[[]Area](&areas, "areas")

	fmt.Printf("Loaded %d materials.", len(materials))
	fmt.Printf("Loaded %d areas.", len(areas))
}

func areaFromName(s string) (area Area, success bool) {
	for _, area := range areas {
		if area.Name == s {
			return area, true
		}
	}
	return Area{}, false
}
