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

func WriteMaterialsToFile() error {
	data, err := json.Marshal(materials)
	if err != nil {
		return fmt.Errorf("error marshalling materials: %w", err)
	}

	file, err := os.Create("./level/data/materials.json")
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

func WriteColorsToFile() error {
	data, err := json.Marshal(colors)
	if err != nil {
		return fmt.Errorf("error marshalling colorss: %w", err)
	}

	file, err := os.Create("./level/data/colors.json")
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

func createCSSFile() {
	cssFile, err := os.Create("./level/assets/colors.css")
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
