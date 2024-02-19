package main

import (
	"encoding/json"
	"fmt"
	"os"
)

var materials []Material
var areas []Area
var colors []Color

func populateFromJson() {
	colorFile := "./level/data/colors.json"
	colors = parseJsonFile[Color](colorFile)

	materialFile := "./level/data/materials.json"
	materials = parseJsonFile[Material](materialFile)

	areaFile := "./level/data/areas.json"
	areas = parseJsonFile[Area](areaFile)
}

func sliceToMap[T any](slice []T, f func(T) string) map[string]T {
	out := make(map[string]T)
	for _, entry := range slice {
		out[f(entry)] = entry
	}
	return out
}

func mapToSlice[T any](m map[string]T) []T {
	out := make([]T, len(m))
	i := 0
	for _, e := range m {
		out[i] = e
		i++
	}
	return out
}

func parseJsonFile[T any](filename string) []T {
	var out []T

	jsonData, err := os.ReadFile(filename)
	if err != nil {
		panic(err)
	}

	if err := json.Unmarshal(jsonData, &out); err != nil {
		panic(err)
	}

	fmt.Printf("Loaded %d entries.\n", len(out))

	return out
}

func colorName(c Color) string {
	return c.CssClassName
}

func materialName(m Material) string {
	return m.CommonName
}
