package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type Context struct {
	Collections map[string]*Collection
	materials   []Material
	colors      []Color

	colorPath      string
	materialPath   string
	cssPath        string
	collectionPath string
}

const DEPLOY_materialPath = "../../server/main/data/materials.json"
const DEPLOY_areaPath = "../../server/main/data/areas.json"
const DEPLOY_cssPath = "../../server/main/assets/colors.css"

// Startup

func populateFromJson() Context {
	var c Context

	// I don't like this
	c.colorPath = "./data/colors/colors.json"
	c.materialPath = "./data/materials/materials.json"
	c.cssPath = "./assets/colors.css"
	c.collectionPath = "./data/collections/"

	c.colors = parseJsonFile[Color](c.colorPath)
	c.materials = parseJsonFile[Material](c.materialPath)
	c.Collections = getAllCollections(c.collectionPath)

	return c
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

	fmt.Printf("Loaded %d entries of %T.\n", len(out), *new(T))

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

func (c Context) writeMaterialsToLocalFile() error {
	return writeJsonFile(c.materialPath, c.materials)
}

func (c Context) writeColorsToLocalFile() error {
	return writeJsonFile(c.colorPath, c.colors)
}

// Combine with below
func (c Context) createLocalCSSFile() {
	c.createCSSFile(c.cssPath)
}

func (c Context) createCSSFile(path string) {
	cssFile, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	defer cssFile.Close()

	for _, color := range c.colors {
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

// Collections
func getAllCollections(collectionPath string) map[string]*Collection {
	dirs, err := os.ReadDir(collectionPath)
	if err != nil {
		fmt.Println(err)
	}

	collections := make(map[string]*Collection)
	for _, dir := range dirs {
		entry, _ := dir.Info()
		if entry.IsDir() {
			collection := Collection{Name: entry.Name()}

			pathToSpaces := filepath.Join(collectionPath, entry.Name(), "spaces")
			areaMap := make(map[string][]Area)
			// getListOfSubDirectorries
			// getListOfJSONFiles
			populateMaps(areaMap, pathToSpaces)
			collection.Spaces = areasToSpaces(areaMap, entry.Name())

			pathToFragments := filepath.Join(collectionPath, entry.Name(), "fragments")
			fragmentMap := make(map[string][]Fragment)
			populateMaps(fragmentMap, pathToFragments)
			collection.Fragments = addSetNamesToFragments(fragmentMap)

			pathToPrototypes := filepath.Join(collectionPath, entry.Name(), "prototypes")
			prototypeMap := make(map[string][]Prototype)
			populateMaps(prototypeMap, pathToPrototypes)
			collection.PrototypeSets = addSetNamesToProtypes(prototypeMap)
			collection.Prototypes = consolidate(collection.PrototypeSets) // consolidate prototypes

			collections[entry.Name()] = &collection

		}
	}
	return collections
}

func consolidate(prototypeSets map[string][]Prototype) map[string]*Prototype {
	out := make(map[string]*Prototype)
	for name, set := range prototypeSets {
		for i := range set {
			if out[set[i].ID] != nil {
				panic("Invalid: duplicate ID for prototypes: " + set[i].ID + " in: " + name)
			}
			out[set[i].ID] = &set[i]
		}
	}
	return out
}

func addSetNamesToFragments(fragmentMap map[string][]Fragment) map[string][]Fragment {
	for setName := range fragmentMap {
		for i := range fragmentMap[setName] {
			fragmentMap[setName][i].SetName = setName
		}
	}
	return fragmentMap
}

func addSetNamesToProtypes(protoMap map[string][]Prototype) map[string][]Prototype {
	out := make(map[string][]Prototype)
	for setName := range protoMap {
		arr := make([]Prototype, 0)
		for i := range protoMap[setName] {
			proto := protoMap[setName][i]
			proto.SetName = setName
			arr = append(arr, proto)
		}
		out[setName] = arr
	}
	return out
}

func areasToSpaces(areaMap map[string][]Area, collectionName string) map[string]*Space {
	out := make(map[string]*Space)
	for name, areas := range areaMap {
		out[name] = &Space{CollectionName: collectionName, Name: name, Areas: areas}
	}
	return out
}

func populateMaps[T any](m map[string][]T, pathToJsonDirectory string) {
	subEntries, err := os.ReadDir(pathToJsonDirectory)
	if err != nil {
		fmt.Println("Invalid directory: " + pathToJsonDirectory)
		return
	}

	for _, subEntry := range subEntries {
		if subEntry.IsDir() {
			fmt.Println("Ignoring misc directory: " + subEntry.Name())
			continue
		}
		parts := strings.Split(subEntry.Name(), ".")
		if len(parts) == 2 && strings.ToLower(parts[1]) == "json" {
			nameOfFile := strings.ToLower(parts[0])
			items := parseJsonFile[T](filepath.Join(pathToJsonDirectory, subEntry.Name()))
			m[nameOfFile] = items
		}
	}
}

func (c Context) getSpace(collectionName string, spaceName string) *Space {
	collection, ok := c.Collections[collectionName]
	if !ok {
		return nil
	}
	return collection.Spaces[spaceName]
}

// DEPLOYMENT

func (c Context) deploy(w http.ResponseWriter, r *http.Request) {
	queryValues := r.URL.Query()
	collectionName := queryValues.Get("currentCollection")
	c.deployLocalChanges(collectionName)
}

func (c Context) deployLocalChanges(collectionName string) {
	c.createCSSFile(DEPLOY_cssPath)
	flatAreas := collectionToAreas(c.Collections[collectionName])
	writeJsonFile(DEPLOY_areaPath, flatAreas)
	writeJsonFile(DEPLOY_materialPath, c.materials)
}

func collectionToAreas(collection *Collection) []Area {
	var out []Area
	for _, space := range collection.Spaces {
		out = append(out, space.Areas...)
	}
	return out
}
