package main

import (
	"encoding/json"
	"fmt"
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

type Color struct {
	CssClassName string `json:"cssClassName"`
	R            int    `json:"R"`
	G            int    `json:"G"`
	B            int    `json:"B"`
	A            string `json:"A"`
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

func writeJsonFile[T any](path string, entries T, pretty bool) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}
	defer file.Close()

	enc := json.NewEncoder(file)
	if pretty {
		enc.SetIndent("", "  ")
	}
	if err := enc.Encode(entries); err != nil {
		return fmt.Errorf("error encoding JSON: %w", err)
	}

	return nil
}

func (c Context) writeColorsToLocalFile() error {
	return writeJsonFile(COLOR_PATH, c.colors, true)
}

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
			}

			pathToSpaces := filepath.Join(collectionPath, entry.Name(), "spaces")
			populateMaps(collection.Spaces, pathToSpaces)

			pathToFragments := filepath.Join(collectionPath, entry.Name(), "fragments")
			populateMaps(collection.Fragments, pathToFragments)

			pathToPrototypes := filepath.Join(collectionPath, entry.Name(), "prototypes")
			populateMaps(collection.PrototypeSets, pathToPrototypes)

			pathToInteractables := filepath.Join(collectionPath, entry.Name(), "interactables")
			populateMaps(collection.InteractableSets, pathToInteractables)

			collections[entry.Name()] = &collection

		}
	}
	return collections
}

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

// DEPLOYMENT

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

func (c Context) compileCollectionByName(collectionName string) {
	os.RemoveAll(COMPILE_basePath)
	os.MkdirAll(COMPILE_imagePath, 0755)
	if c.Collections[collectionName] == nil {
		panic("invalid collection")
	}
	c.compileCollection(c.Collections[collectionName])
}

func (c Context) compileCollection(collection *Collection) {
	areas := make([]AreaOutput, 0)

	for _, space := range collection.Spaces {
		c.generateAllPNGs(space)
		for _, desc := range space.Areas {

			mapid := ""
			if space.isSimplyTiled() {
				mapid = c.copyMapPNG(space, &desc)
			}
			// Add maps for all individual areas as well

			areas = append(areas, collection.areaOutputFromDescription(desc, mapid))
		}
	}
	fmt.Printf("Writing (%d) Areas", len(areas))
	writeJsonFile(filepath.Join(COMPILE_basePath, AREA_FILENAME), areas, false)
}

func (col Collection) areaOutputFromDescription(desc AreaDescription, mapid string) AreaOutput {
	outputTiles, err := col.compileMaterialsFromBlueprint(desc.Blueprint)
	if err != nil {
		panic(desc.Name + ": has compile error: " + err.Error())
	}

	outputInteractables := col.generateInteractables(desc.Blueprint.Tiles)

	return AreaOutput{
		Name:           desc.Name,
		Safe:           desc.Safe,
		Tiles:          outputTiles,
		Interactables:  outputInteractables,
		Transports:     desc.Transports,
		North:          desc.North,
		South:          desc.South,
		East:           desc.East,
		West:           desc.West,
		MapId:          mapid,
		LoadStrategy:   desc.LoadStrategy,
		SpawnStrategy:  desc.SpawnStrategy,
		Weather:        desc.Weather,
		BroadcastGroup: desc.BroadcastGroup,
	}
}

func (col *Collection) generateInteractables(tiles [][]TileData) [][]*InteractableDescription {
	out := make([][]*InteractableDescription, len(tiles))
	for i := range tiles {
		out[i] = make([]*InteractableDescription, len(tiles[i]))
		for j := range tiles[i] {
			out[i][j] = col.resolveInteractableByTile(tiles[i][j])
		}
	}
	return out
}

func (col *Collection) resolveInteractableByTile(tile TileData) *InteractableDescription {
	base := col.findInteractableById(tile.InteractableId)
	if base == nil {
		return nil
	}

	out := *base
	if base.States != nil {
		out.States = make(map[string]InteractableStateDescription, len(base.States))
		for key, value := range base.States {
			out.States[key] = value
		}
	}

	stateName := tile.InteractableState
	if stateName == "" {
		stateName = out.DefaultState
	}
	if stateName == "" {
		stateName = "default"
	}

	if out.DefaultState == "" {
		out.DefaultState = "default"
	}
	if out.States == nil {
		out.States = map[string]InteractableStateDescription{}
	}
	if _, ok := out.States[out.DefaultState]; !ok {
		out.States[out.DefaultState] = InteractableStateDescription{
			CssClass:       out.CssClass,
			Pushable:       out.Pushable,
			Walkable:       out.Walkable,
			Fragile:        out.Fragile,
			Sticky:         out.Sticky,
			RejectTeleport: out.RejectTeleport,
			Reactions:      out.Reactions,
			ReactionRules:  out.ReactionRules,
		}
	}

	selected := out.States[stateName]
	if _, ok := out.States[stateName]; !ok {
		stateName = out.DefaultState
		selected = out.States[stateName]
	}

	out.State = stateName
	out.CssClass = selected.CssClass
	out.Pushable = selected.Pushable
	out.Walkable = selected.Walkable
	out.Fragile = selected.Fragile
	out.Sticky = selected.Sticky
	out.RejectTeleport = selected.RejectTeleport
	out.Reactions = selected.Reactions
	out.ReactionRules = append([]ReactionRule(nil), selected.ReactionRules...)

	return &out
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

func (collection *Collection) compileMaterialsFromBlueprint(bp *Blueprint) ([][]Material, error) {
	outputTiles := make([][]Material, len(bp.Tiles))
	for y := range bp.Tiles {
		outputTiles[y] = make([]Material, len(bp.Tiles[y]))
		for x, tile := range bp.Tiles[y] {
			// Find proto
			proto := collection.findPrototypeById(tile.PrototypeId)
			if proto == nil {
				return nil, fmt.Errorf("Prototype with id: %s Not found. y:%d x:%d", bp.Tiles[y][x].PrototypeId, y, x)
			}

			// Apply transform
			mat := proto.applyTransform(tile.Transformation)

			// Apply ground
			ground := groundCellByCoord(bp, y, x)
			mat = addGroundToMaterial(mat, ground, bp.DefaultTileColor, bp.DefaultTileColor1)

			// Assign
			outputTiles[y][x] = mat
		}
	}
	return outputTiles, nil
}

func addGroundToMaterial(material Material, cell *Cell, color0, color1 string) Material {
	if cell == nil {
		material.Ground1Css = color0
		return material
	}
	if material.Ground2Css != "" {
		return material
	}
	primary, secondary := color0, color1
	if cell.Status != 0 {
		primary, secondary = color1, color0
	}
	material.Ground2Css = primary
	if cell.TopLeft || cell.TopRight || cell.BottomLeft || cell.BottomRight {
		material.Ground1Css = secondary
	}
	if cell.TopLeft {
		material.Ground2Css += " r0-tl"
	}
	if cell.TopRight {
		material.Ground2Css += " r0-tr"
	}
	if cell.BottomLeft {
		material.Ground2Css += " r0-bl"
	}
	if cell.BottomRight {
		material.Ground2Css += " r0-br"
	}
	return material
}
