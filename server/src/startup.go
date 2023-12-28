package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
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
	jsonData, err := os.ReadFile(fmt.Sprintf("./server/src/data/%s.json", fn))
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

	fmt.Println(len(materials))
	fmt.Println(len(areas))
}

func areaFromName(s string) Area {
	for _, area := range areas {
		if area.Name == s {
			return area
		}
	}
	panic("Area not found")
}

func stageFromArea(s string) Stage {
	area := areaFromName(s)
	tiles := make([][]Tile, len(area.Tiles))
	for y := range tiles {
		tiles[y] = make([]Tile, len(area.Tiles[y]))
		for x := range tiles[y] {
			tiles[y][x] = newTile(materials[area.Tiles[y][x]])
		}
	}
	for _, transport := range area.Transports {
		tiles[transport.SourceY][transport.SourceX].Teleport = &Teleport{transport.DestStage, transport.DestY, transport.DestX}
	}
	return Stage{tiles: tiles, playerMap: make(map[string]*Player), playerMutex: sync.Mutex{}}
}
