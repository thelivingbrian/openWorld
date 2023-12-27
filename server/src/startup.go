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
	MaterialID int    `json:"materialId"`
	DestY      int    `json:"destY"`
	DestX      int    `json:"destX"`
	DestStage  string `json:"destStage"`
}

type Area struct {
	Name      string      `json:"name"`
	Safe      bool        `json:"safe"`
	Tiles     [][]int     `json:"tiles"`
	Transport []Transport `json:"transport"`
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
	for i, _ := range tiles {
		tiles[i] = make([]Tile, len(area.Tiles[i]))
		for j, _ := range tiles[i] {
			tiles[i][j] = newTile(materials[area.Tiles[i][j]])
		}
	}
	return Stage{tiles: tiles, playerMap: make(map[string]*Player), playerMutex: sync.Mutex{}}
}
