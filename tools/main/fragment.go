package main

import (
	"fmt"
	"regexp"
)

type Fragment struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	SetName   string     `json:"setName"`
	Blueprint *Blueprint `json:"blueprint"`
}

func (col *Collection) generateMaterials(bp *Blueprint) [][]Material {
	tiles := bp.Tiles
	out := make([][]Material, len(tiles))
	for i := range tiles {
		out[i] = make([]Material, len(tiles[i]))
		for j := range tiles[i] {
			out[i][j] = col.createMaterial(bp, i, j)
		}
	}
	return out
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

func (col *Collection) createMaterial(bp *Blueprint, y, x int) Material {
	data := bp.Tiles[y][x]
	proto := col.findPrototypeById(data.PrototypeId)
	if proto == nil {
		proto = &Prototype{ID: "INVALID-", CssColor: "blue", Floor1Css: "green red-b thick"}
	}

	mat := proto.applyTransformForEditor(data.Transformation)
	ground := groundCellByCoord(bp, y, x)
	mat = addGroundToMaterial(mat, ground, bp.DefaultTileColor, bp.DefaultTileColor1)

	return mat
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

func emptyTransformCss(input string) string {
	emptyTransform := Transformation{}
	return transformCss(input, emptyTransform)
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

func panicIfAnyEmpty(errorMessage string, strings ...string) {
	for _, str := range strings {
		if str == "" {
			panic("panicIfAnyEmpty - caller provided error message: " + errorMessage)
		}
	}
}

func (col *Collection) getFragmentFromAssetId(fragmentID string) Fragment {
	fragment := col.getFragmentById(fragmentID)
	if fragment == nil {
		panic("No Fragment with ID: " + fragmentID)
	}
	return *fragment
}

func (col *Collection) getPrototypeOrCreateInvalid(protoId string) Prototype {
	proto := col.findPrototypeById(protoId)
	if proto == nil {
		fmt.Println("Requested invalid proto: " + protoId)
		return Prototype{ID: "INVALID-" + protoId, CssColor: "blue", Floor1Css: "green red-b thick"}
	}

	return *proto
}
