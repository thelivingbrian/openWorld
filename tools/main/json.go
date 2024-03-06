package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

var materials []Material
var areas []Area
var colors []Color

const colorPath = "./level/data/colors.json"
const materialPath = "./level/data/materials.json"
const areaPath = "./level/data/areas.json"
const cssPath = "./level/assets/colors.css"

const DEPLOY_materialPath = "../../server/main/data/materials.json"
const DEPLOY_areaPath = "../../server/main/data/areas.json"
const DEPLOY_cssPath = "../../server/main/assets/colors.css"

func populateFromJson() {
	colors = parseJsonFile[Color](colorPath)
	materials = parseJsonFile[Material](materialPath)
	areas = parseJsonFile[Area](areaPath)
}

func sliceToMap[T any](slice []T, f func(T) string) map[string]T {
	out := make(map[string]T)
	for _, entry := range slice {
		out[f(entry)] = entry
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

func writeJsonFile[T any](path string, entries []T) error {
	data, err := json.Marshal(entries)
	if err != nil {
		return fmt.Errorf("error marshalling materials: %w", err)
	}

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}
	defer file.Close()

	_, err = file.Write(data)
	if err != nil {
		return fmt.Errorf("error writing to file: %w", err)
	}

	return nil
}

func colorName(c Color) string {
	return c.CssClassName
}

func materialName(m Material) string {
	return m.CommonName
}

func writeMaterialsToLocalFile() error {
	return writeJsonFile(materialPath, materials)
}

func writeColorsToLocalFile() error {
	return writeJsonFile(colorPath, colors)
}

func createLocalCSSFile() {
	createCSSFile(cssPath)
}

func createCSSFile(path string) {
	cssFile, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	defer cssFile.Close()

	for _, color := range colors {
		rgbstring := fmt.Sprintf("rgb(%d, %d, %d)", color.R, color.G, color.B)
		if color.A != "" {
			rgbstring = fmt.Sprintf("rgba(%d, %d, %d, %s)", color.R, color.G, color.B, color.A)
		}
		cssRule := fmt.Sprintf(".%s { background-color: %s; }\n\n.%s-b { border-color: %s; }\n\n", color.CssClassName, rgbstring, color.CssClassName, rgbstring)
		_, err := cssFile.WriteString(cssRule)
		if err != nil {
			panic(err)
		}
	}
}

func deploy(w http.ResponseWriter, r *http.Request) {
	deployLocalChanges()
}

func deployLocalChanges() {
	populateFromJson()
	createCSSFile(DEPLOY_cssPath)
	writeJsonFile(DEPLOY_areaPath, areas)
	writeJsonFile(DEPLOY_materialPath, materials)
}
