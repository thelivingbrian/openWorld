package main

import (
	"regexp"
)

type Fragment struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	SetName   string     `json:"setName"`
	Blueprint *Blueprint `json:"blueprint"`
}

func groundCellByCoord(bp *Blueprint, y, x int) *Cell {
	if bp == nil || bp.Ground == nil {
		return nil
	}
	if y < 0 || x < 0 || y >= len(bp.Ground) || x >= len(bp.Ground[y]) {
		return nil
	}
	return &bp.Ground[y][x]
}

func (col *Collection) generateInteractables(tiles [][]TileData) [][]*InteractableDescription {
	out := make([][]*InteractableDescription, len(tiles))
	for i := range tiles {
		out[i] = make([]*InteractableDescription, len(tiles[i]))
		for j := range tiles[i] {
			out[i][j] = col.findInteractableById(tiles[i][j].InteractableId)
		}
	}
	return out
}

func transformCss(input string, transformation Transformation) string {
	pattern := regexp.MustCompile(`{([^:]*):([^}]*)}`)

	result := pattern.ReplaceAllStringFunc(input, func(s string) string {
		matches := pattern.FindStringSubmatch(s)
		if len(matches) == 3 {
			if matches[1] == "rotate" {
				return rotateCss(matches[2], transformation.ClockwiseRotations)
			}
			return matches[2]
		}
		panic("Have match " + s + " But submatch behavior is undefined (submatches != 3)")
	})
	return result
}

func rotateCss(input string, clockwiseRotations int) string {
	options := []string{"tr", "br", "bl", "tl"}
	currentIndex := findIndex(input, options)
	if currentIndex == -1 {
		panic("invalid rotation attempted")
	}
	return options[mod(currentIndex+clockwiseRotations, 4)]
}

func findIndex(s string, list []string) int {
	for i := range list {
		if list[i] == s {
			return i
		}
	}
	return -1
}
