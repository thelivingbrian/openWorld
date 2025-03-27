package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

// How to expand to non-local collection source?
type Context struct {
	Collections map[string]*Collection
	colors      []Color
}

// Break everything out for compile (using funcs)
// Deploy should only need base path because it is just a copy of compile ?
const COMPILE_basePath = "./data/out"
const COMPILE_imagePath = COMPILE_basePath + "/images"

const DEPLOY_basePath = "../../server/main/data"
const DEPLOY_imagePath = DEPLOY_basePath + "/images"
const DEPLOY_cssPath = "../../server/main/assets/colors.css"

const AREA_FILENAME = "areas.json"
const MATERIAL_FILENAME = "materials.json"

const COLOR_PATH string = "./data/colors/colors.json"
const CSS_PATH string = "./assets/colors.css"
const COLLECTION_PATH string = "./data/collections/"

// Startup
func populateFromJson() Context {
	var c Context

	c.colors = parseJsonFile[[]Color](COLOR_PATH)
	c.Collections = c.getAllCollections(COLLECTION_PATH)

	return c
}

func parseJsonFile[T any](filename string) T {
	var out T

	jsonData, err := os.ReadFile(filename)
	if err != nil {
		panic(err)
	}

	if err := json.Unmarshal(jsonData, &out); err != nil {
		panic(err)
	}

	fmt.Printf("Loaded %s. Contents: %T.\n", filename, *new(T))

	return out
}

func writeJsonFile[T any](path string, entries T) error {
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

func (c Context) writeColorsToLocalFile() error {
	return writeJsonFile(COLOR_PATH, c.colors)
}

// Combine with below
func (c Context) createLocalCSSFile() {
	c.createCSSFile(CSS_PATH)
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
		cssRule := fmt.Sprintf(".%s { background-color: %s; }\n", color.CssClassName, rgbstring)
		cssRule += fmt.Sprintf(".%s-b { border-color: %s; }\n", color.CssClassName, rgbstring)
		cssRule += fmt.Sprintf(".%s-t { color: %s; }\n\n", color.CssClassName, rgbstring)
		_, err := cssFile.WriteString(cssRule)
		if err != nil {
			panic(err)
		}
	}
}

// Helper
func (c Context) pathToMapsForSpace(space *Space) string {
	return COLLECTION_PATH + space.CollectionName + "/spaces/maps/" + space.Name + "/"
}

// Collections
func (c Context) getAllCollections(collectionPath string) map[string]*Collection {
	dirs, err := os.ReadDir(collectionPath)
	if err != nil {
		fmt.Println(err)
	}

	collections := make(map[string]*Collection)
	for _, dir := range dirs {
		entry, _ := dir.Info()
		if entry.IsDir() {
			collection := Collection{
				Name:             entry.Name(),
				Spaces:           make(map[string]*Space),
				Fragments:        make(map[string][]Fragment),
				PrototypeSets:    make(map[string][]Prototype),
				InteractableSets: make(map[string][]InteractableDescription),
				StructureSets:    make(map[string][]Structure),
			}

			pathToSpaces := filepath.Join(collectionPath, entry.Name(), "spaces")
			populateMaps(collection.Spaces, pathToSpaces)

			pathToFragments := filepath.Join(collectionPath, entry.Name(), "fragments")
			populateMaps(collection.Fragments, pathToFragments)

			pathToPrototypes := filepath.Join(collectionPath, entry.Name(), "prototypes")
			populateMaps(collection.PrototypeSets, pathToPrototypes)

			pathToInteractables := filepath.Join(collectionPath, entry.Name(), "interactables")
			populateMaps(collection.InteractableSets, pathToInteractables)

			pathToStructures := filepath.Join(collectionPath, entry.Name(), "structures")
			populateMaps(collection.StructureSets, pathToStructures)

			collections[entry.Name()] = &collection

		}
	}
	return collections
}

/*
func addSetNamesToFragments(fragmentMap map[string][]Fragment) map[string][]Fragment {
	for setName := range fragmentMap {
		for i := range fragmentMap[setName] {
			fragmentMap[setName][i].SetName = setName
		}
	}
	return fragmentMap
}

func (c Context) addSetNamesToProtypes(protoMap map[string][]Prototype) map[string][]Prototype {
	out := make(map[string][]Prototype)
	for setName := range protoMap {
		arr := make([]Prototype, 0)
		for i := range protoMap[setName] {
			proto := protoMap[setName][i]
			proto.SetName = setName

			// Add map color for old protos
			// proto.MapColor = c.getMapColorFromProto(proto)

			arr = append(arr, proto)
		}
		out[setName] = arr
	}
	return out
}
*/

func populateMaps[T any](m map[string]T, pathToJsonDirectory string) {
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

func (c Context) spaceFromNames(collectionName string, spaceName string) *Space {
	collection, ok := c.Collections[collectionName]
	if !ok {
		return nil
	}
	return collection.Spaces[spaceName]
}

// DEPLOYMENT

func (c Context) deployHandler(w http.ResponseWriter, r *http.Request) {
	queryValues := r.URL.Query()
	collectionName := queryValues.Get("currentCollection")
	c.deploy(collectionName)
}

func (c Context) deploy(collectionName string) {
	c.createCSSFile(DEPLOY_cssPath)
	c.compileCollectionByName(collectionName)
	os.RemoveAll(DEPLOY_basePath)
	os.MkdirAll(DEPLOY_imagePath, 0755)
	err := copyDir(COMPILE_basePath, DEPLOY_basePath)
	if err != nil {
		panic(err)
	}
}

func (c Context) compile(w http.ResponseWriter, r *http.Request) {
	queryValues := r.URL.Query()
	collectionName := queryValues.Get("currentCollection")
	c.createCSSFile(CSS_PATH)
	c.compileCollectionByName(collectionName)
}

func (c Context) compileCollectionByName(collectionName string) {
	os.RemoveAll(COMPILE_basePath)
	os.MkdirAll(COMPILE_imagePath, 0755)
	if c.Collections[collectionName] == nil {
		panic("invalid collection")
	}
	c.compileCollection(c.Collections[collectionName])
}

func (c Context) compileCollection(collection *Collection) {
	//mapToMaterials := make(map[Transformation]map[string]Material)
	//materials := make([]Material, 0)
	areas := make([]AreaOutput, 0)

	for _, space := range collection.Spaces {
		c.generateAllPNGs(space)
		for _, desc := range space.Areas {
			//var outputTiles [][]int
			outputTiles := collection.compileTileDataAndAccumulateMaterials(desc)

			mapid := ""
			if space.isSimplyTiled() {
				mapid = c.copyMapPNG(space, &desc)
			}
			// Add maps for all individual areas as well

			areas = append(areas, AreaOutput{
				Name:             desc.Name,
				Safe:             desc.Safe,
				Tiles:            outputTiles,
				Interactables:    collection.generateInteractables(desc.Blueprint.Tiles),
				Transports:       desc.Transports,
				DefaultTileColor: desc.Blueprint.DefaultTileColor, // used?
				North:            desc.North,
				South:            desc.South,
				East:             desc.East,
				West:             desc.West,
				MapId:            mapid,
				LoadStrategy:     desc.LoadStrategy,
				SpawnStrategy:    desc.SpawnStrategy,
			})
		}
	}
	fmt.Printf("Writing (%d) Areas", len(areas))
	writeJsonFile(filepath.Join(COMPILE_basePath, AREA_FILENAME), areas)
	//fmt.Printf("Writing (%d) Materials", len(materials))
	//writeJsonFile(filepath.Join(COMPILE_basePath, MATERIAL_FILENAME), materials)

}

func (c Context) copyMapPNG(space *Space, area *AreaDescription) string {
	src := filepath.Join(c.pathToMapsForSpace(space), areaToFilename(area))
	id := uuid.New().String()
	filename := fmt.Sprintf("%s.png", id)

	dest := filepath.Join("./data/out/images", filename)
	err := copyFile(src, dest)
	if err != nil {
		panic(err)
	}
	return id
}

func (collection *Collection) compileTileDataAndAccumulateMaterials(desc AreaDescription) [][]Material {
	outputTiles := make([][]Material, len(desc.Blueprint.Tiles))
	for y := range desc.Blueprint.Tiles {
		outputTiles[y] = make([]Material, len(desc.Blueprint.Tiles[y]))
		for x := range desc.Blueprint.Tiles[y] {
			//var id int
			tile := desc.Blueprint.Tiles[y][x]
			proto := collection.findPrototypeById(tile.PrototypeId)
			if proto == nil {
				errMsg := fmt.Sprintf("Prototype with id: %s Not found. Area: %s | y:%d x:%d", desc.Blueprint.Tiles[y][x].PrototypeId, desc.Name, y, x)
				panic("PROTO NOT FOUND. error - " + errMsg)
			}

			/*
				protoToMat, found := mapToMaterials[desc.Blueprint.Tiles[y][x].Transformation]
				if found {
					_, found = protoToMat[desc.Blueprint.Tiles[y][x].PrototypeId]
					if !found {
						id = len(materials)
						newMaterial := prototype.applyTransform(desc.Blueprint.Tiles[y][x].Transformation, len(materials))
						protoToMat[desc.Blueprint.Tiles[y][x].PrototypeId] = newMaterial
						materials = append(materials, newMaterial)
					} else {
						id = 7 //protoToMat[desc.Blueprint.Tiles[y][x].PrototypeId].ID
					}
				} else {
					protoToMat = make(map[string]Material)
					id = len(materials)
					newMaterial := prototype.applyTransform(desc.Blueprint.Tiles[y][x].Transformation, len(materials))
					protoToMat[desc.Blueprint.Tiles[y][x].PrototypeId] = newMaterial
					materials = append(materials, newMaterial)
					mapToMaterials[desc.Blueprint.Tiles[y][x].Transformation] = protoToMat
				}
				// Is added step worth it or should server areas have materials by value?
			*/

			// getProto -> transform -> add ground -> description includes [][]Material

			// Apply transform
			mat := proto.applyTransformForEditor(tile.Transformation)

			// Apply ground
			ground := groundCellByCoord(desc.Blueprint, y, x)
			mat = addGroundToMaterial(mat, ground, desc.Blueprint.DefaultTileColor, desc.Blueprint.DefaultTileColor1)

			// // Css Overrides
			// if proto.CssColor != "" {
			// 	mat.Ground2Css = proto.CssColor
			// }

			outputTiles[y][x] = mat
		}
	}
	return outputTiles
}
