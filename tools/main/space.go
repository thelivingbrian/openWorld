package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

type Space struct {
	CollectionName string
	Name           string
	Topology       string
	Latitude       int
	Longitude      int
	AreaHeight     int
	AreaWidth      int
	Areas          []AreaDescription
}

func (c Context) spacesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		c.getSpaces(w, r)
	}
	if r.Method == "POST" {
		c.postSpaces(w, r)
	}
}

func (c Context) getSpaces(w http.ResponseWriter, r *http.Request) {
	queryValues := r.URL.Query()
	collectionName := queryValues.Get("collectionName")
	if col, ok := c.Collections[collectionName]; ok {
		err := tmpl.ExecuteTemplate(w, "space-page", col)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func (c Context) postSpaces(w http.ResponseWriter, r *http.Request) {
	props, ok := requestToProperties(r)
	if !ok {
		fmt.Println("Invalid POST to /spaces. Properties are invalid.")
		io.WriteString(w, `<h3> Properties are invalid. </h3>`)
		return
	}
	cName, ok := props["currentCollection"]
	if !ok {
		fmt.Println("Invalid POST to /spaces. Collection not found.")
		io.WriteString(w, `<h3> Collection not found. </h3>`)
		return
	}
	if col, ok := c.Collections[cName]; ok {
		fmt.Println(col.Name)
		valid := true

		name, ok := props["newSpaceName"]
		valid = valid && ok

		lat, ok := props["latitude"]
		valid = valid && ok

		long, ok := props["longitude"]
		valid = valid && ok

		topology, ok := props["topology"]
		valid = valid && ok

		areaWidth, ok := props["areaWidth"]
		valid = valid && ok

		areaHeight, ok := props["areaHeight"]
		valid = valid && ok

		tileColor, ok := props["tileColor"]
		valid = valid && ok

		tileColor1, ok := props["tileColor1"]
		valid = valid && ok

		weather, ok := props["weather"]
		valid = valid && ok

		broadcastGroup, ok := props["broadcastGroup"]
		valid = valid && ok
		if !valid {
			fmt.Println("Invalid, failed to get properties by name.")
			io.WriteString(w, `<h3> Properties are invalid.</h3>`)
			return
		}

		latitude, err := strconv.Atoi(lat)
		valid = valid && (err == nil)

		longitude, err := strconv.Atoi(long)
		valid = valid && (err == nil)

		width, err := strconv.Atoi(areaWidth)
		valid = valid && (err == nil)

		height, err := strconv.Atoi(areaHeight)
		valid = valid && (err == nil)
		if !valid {
			fmt.Println(err)
			fmt.Println("Invalid, failed to cast lat/long/h/w.")
			io.WriteString(w, `<h3> Failed to cast. </h3>`)
			return
		}

		fmt.Printf("%s %s %s %s %s %d %d", name, topology, areaWidth, areaHeight, tileColor, latitude, longitude)

		space := createSpace(cName, name, latitude, longitude, topology, height, width, tileColor, tileColor1, weather, broadcastGroup)
		col.Spaces[name] = &space
		io.WriteString(w, `<h3>Success</h3>`)
		return
	}
	io.WriteString(w, `<h3>Invalid collection Name.</h3>`)
}

func createSpace(cName, name string, latitude, longitude int, topology string, height, width int, tileColor, tileColor1, weather, broadcastGroup string) Space {
	areas := make([][]AreaDescription, latitude)
	for y := 0; y < latitude; y++ {
		for x := 0; x < longitude; x++ {
			area := createBaseArea(height, width, tileColor, tileColor1, weather, broadcastGroup)

			if topology != "disconnected" {
				// This is consistent with Tiles
				area.Name = fmt.Sprintf("%s:%d-%d", name, y, x)
				area.North = fmt.Sprintf("%s:%d-%d", name, mod(y-1, latitude), x)
				area.South = fmt.Sprintf("%s:%d-%d", name, mod(y+1, latitude), x)
				area.East = fmt.Sprintf("%s:%d-%d", name, y, mod(x+1, longitude))
				area.West = fmt.Sprintf("%s:%d-%d", name, y, mod(x-1, longitude))
			}
			areas[y] = append(areas[y], area)
		}
	}
	// Remove edges if plane topology
	if topology == "plane" {
		for n := range areas[0] {
			areas[0][n].North = ""
		}
		for m := range areas[len(areas)-1] {
			areas[len(areas)-1][m].South = ""
		}
		for j := range areas {
			areas[j][0].West = ""
			areas[j][len(areas[j])-1].East = ""
		}
	}

	flatAreas := make([]AreaDescription, 0)
	for i := range areas {
		flatAreas = append(flatAreas, areas[i]...)
	}

	out := Space{CollectionName: cName, Name: name, Topology: topology, Latitude: latitude, Longitude: longitude, AreaHeight: height, AreaWidth: width, Areas: flatAreas}
	return out
}

func mod(i, n int) int {
	return ((i % n) + n) % n
}

func createBaseArea(height, width int, tileColor, tileColor1, weather, broadcastGroup string) AreaDescription {
	tiles := make([][]TileData, height)
	for i := range tiles {
		tiles[i] = make([]TileData, width)
	}

	blueprint := Blueprint{Tiles: tiles, DefaultTileColor: tileColor, DefaultTileColor1: tileColor1, Instructions: make([]Instruction, 0)}
	// safe is always false. Can be reset elsewhere.
	return AreaDescription{Name: "", Safe: false, Blueprint: &blueprint, Transports: make([]Transport, 0), Weather: weather, BroadcastGroup: broadcastGroup}
}

func getAreaByName(areas []AreaDescription, name string) *AreaDescription {
	for i, area := range areas {
		if name == area.Name {
			return &areas[i]
		}
	}
	return nil
}

/*
// Could also have getAreaByCoord
func (s *Space) getAreaByName(name string) *AreaDescription {
	return getAreaByName(s.Areas, name)
}

func (s *Space) coordToName(y, x int) string {
	return fmt.Sprintf("%s:%d-%d", s.Name, y, x)
}
*/

func getFragmentByName(fragments []Fragment, name string) *Fragment {
	for i, fragment := range fragments {
		if name == fragment.Name {
			return &fragments[i]
		}
	}
	return nil
}

func (col *Collection) getFragmentById(id string) *Fragment {
	for _, set := range col.Fragments {
		for i, fragment := range set {
			if id == fragment.ID {
				return &set[i]
			}
		}

	}
	return nil
}

//

func (c Context) spaceHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		c.getSpace(w, r)
	}
	if r.Method == "PUT" {
		c.putSpace(w, r)
	}
}

type AreaTile struct {
	ImgUriPath   string
	SelectedArea *AreaDescription
}

type SpaceEditPageData struct {
	//GridDetails   GridDetails
	SelectedSpace Space
	AreaTiles     [][]AreaTile // Should be some combo of area and url
}

func (c Context) getSpace(w http.ResponseWriter, r *http.Request) {
	queryValues := r.URL.Query()
	name := queryValues.Get("currentCollection")
	fmt.Println("Collection Name: " + name)
	space := queryValues.Get("currentSpace")
	fmt.Println("Space: " + space)

	if col, ok := c.Collections[name]; ok {
		if space, ok := col.Spaces[space]; ok {
			fmt.Println(space.Topology)

			var tiles [][]AreaTile
			if space.isSimplyTiled() {
				tiles = make([][]AreaTile, space.Latitude)
				for row := range tiles {
					tiles[row] = make([]AreaTile, space.Longitude)
					for column := range tiles[row] {
						areaName := fmt.Sprintf("%s:%d-%d", space.Name, row, column)
						path := fmt.Sprintf(`/images/make/%s/%s?currentCollection=%s`, space.Name, areaName, col.Name)
						tiles[row][column].ImgUriPath = path
						area := getAreaByName(space.Areas, areaName)
						if area == nil {
							panic("OH NO")
						}
						tiles[row][column].SelectedArea = area
						// Add the actual area
					}
				}
			}

			pagedata := SpaceEditPageData{
				SelectedSpace: *space,
				AreaTiles:     tiles,
			}
			err := tmpl.ExecuteTemplate(w, "space-edit", pagedata)
			if err != nil {
				fmt.Println(err)
			}
		}
	}
}

// Saves changes to space made in the editor
func (c Context) putSpace(w http.ResponseWriter, r *http.Request) {
	properties, _ := requestToProperties(r)
	collectionName := properties["currentCollection"]
	spaceName := properties["currentSpace"]
	panicIfAnyEmpty("PUT to /space", collectionName, spaceName)

	collection, ok := c.Collections[collectionName]
	if !ok {
		panic("no collection")
	}
	collection.saveSpace(spaceName)

	io.WriteString(w, `<h2>Success</h2>`)
}

//

func (c Context) newSpaceHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		queryValues := r.URL.Query()
		name := queryValues.Get("currentCollection")
		fmt.Println("Collection Name: " + name)
		if col, ok := c.Collections[name]; ok {
			col.getNewSpace(w, r)
		}
	}
}

func (col Collection) getNewSpace(w http.ResponseWriter, _ *http.Request) {
	err := tmpl.ExecuteTemplate(w, "space-new", col)
	if err != nil {
		fmt.Println(err)
	}
}

//

func (c Context) spaceDetailsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		c.getSpaceDetails(w, r)
	}
}

func (c Context) getSpaceDetails(w http.ResponseWriter, r *http.Request) {
	space := c.spaceFromGET(r)

	// Have default tile color change trigger redisplay screen
	err := tmpl.ExecuteTemplate(w, "space-details", space)
	if err != nil {
		fmt.Println(err)
	}
}

func (c Context) spaceStructuresHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		c.getSpaceStructures(w, r)
	}
}

func (c Context) getSpaceStructures(w http.ResponseWriter, _ *http.Request) {
	err := tmpl.ExecuteTemplate(w, "structure-select", nil)
	if err != nil {
		fmt.Println(err)
	}
}

func (c Context) spaceStructureHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		c.getSpaceStructure(w, r)
	}
	if r.Method == "POST" {
		c.postSpaceStructure(w, r)
	}
	if r.Method == "DELETE" {
		c.deleteSpaceStructure(w, r)
	}
}

func (c Context) spaceModifyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		c.getSpaceModify(w, r)
	}
	if r.Method == "POST" {
		c.postSpaceModify(w, r)
	}
}

func (c Context) postSpaceModify(w http.ResponseWriter, r *http.Request) {
	properties, _ := requestToProperties(r)
	collectionName := properties["currentCollection"]
	spaceName := properties["currentSpace"]
	space := c.spaceFromNames(collectionName, spaceName)
	if space == nil {
		panic("invalid space")
	}
	defaultColor, haveDefault := properties["default-tile-color"]
	defaultColor1, haveDefault1 := properties["default-tile-color1"]
	_, haveSafe := properties["safe-update"]
	safe := properties["safe"] == "on"
	weather, haveWeather := properties["weather"]
	broadcastGroup, haveBroadcast := properties["broadcast-group"]
	for i := range space.Areas {
		if haveDefault {
			space.Areas[i].Blueprint.DefaultTileColor = defaultColor
		}
		if haveDefault1 {
			space.Areas[i].Blueprint.DefaultTileColor1 = defaultColor1
		}
		if haveSafe {
			space.Areas[i].Safe = safe
		}
		if haveWeather {
			space.Areas[i].Weather = weather
		}
		if haveBroadcast {
			space.Areas[i].BroadcastGroup = broadcastGroup
		}
	}
	io.WriteString(w, "<h2>done</h2>")
}

func (c Context) getSpaceModify(w http.ResponseWriter, _ *http.Request) {
	err := tmpl.ExecuteTemplate(w, "space-modify-areas", nil)
	if err != nil {
		fmt.Println(err)
	}
}

func (c Context) getSpaceStructure(w http.ResponseWriter, r *http.Request) {
	queryValues := r.URL.Query()
	collectionName := queryValues.Get("currentCollection")
	structureType := queryValues.Get("structureType")

	if structureType == "ground" {
		col, ok := c.Collections[collectionName]
		if !ok {
			io.WriteString(w, "<h2>Invalid collection.</h2>")
		}
		grounds := col.StructureSets["ground"]
		err := tmpl.ExecuteTemplate(w, "structure-ground", grounds)
		if err != nil {
			fmt.Println(err)
		}
	} else {
		io.WriteString(w, "<h2>Sorry invalid structure selected.</h2>")
	}
}

func (c Context) postSpaceStructure(_ http.ResponseWriter, r *http.Request) {
	properties, _ := requestToProperties(r)
	collectionName := properties["currentCollection"]
	spaceName := properties["currentSpace"]
	lat := properties["lat"]
	latI, err := strconv.Atoi(lat)
	if err != nil {
		panic("Invalid latitude")
	}
	long := properties["long"]
	longI, err := strconv.Atoi(long)
	if err != nil {
		panic("Invalid longitude")
	}
	id := properties["structureId"]
	structureType := properties["structure-type"]
	panicIfAnyEmpty("PUT to /space/structure", collectionName, spaceName)

	space := c.spaceFromNames(collectionName, spaceName)
	if space == nil {
		panic("invalid space")
	}
	fmt.Printf("place %s on %s : %s - %s", id, space.Name, lat, long)

	// get each blueprint
	col, ok := c.Collections[collectionName]
	if !ok {
		panic("No collection")
	}
	structures, ok := col.StructureSets[structureType]
	if !ok {
		panic("No structures for: " + structureType)
	}
	structure, found := getStructureById(structures, id)
	if !found {
		panic("No Structure")
	}
	for i := 0; i < len(structure.FragmentIds); i++ {
		for j := 0; j < len(structure.FragmentIds[i]); j++ {
			if structure.FragmentIds[i][j] != "" {
				areaname := fmt.Sprintf("%s:%d-%d", space.Name, latI+i, longI+j)
				area := getAreaByName(space.Areas, areaname)
				if area == nil {
					continue
				}
				area.Blueprint.Instructions = append(area.Blueprint.Instructions, Instruction{
					ID:                 uuid.New().String(),
					X:                  0,
					Y:                  0,
					GridAssetId:        structure.FragmentIds[i][j],
					ClockwiseRotations: 0,
				})
				for _, instruction := range area.Blueprint.Instructions {
					col.applyInstruction(area.Blueprint.Tiles, instruction)
				}
			}
		}
	}
}

func (c Context) deleteSpaceStructure(_ http.ResponseWriter, r *http.Request) {
	fmt.Println("Attempting to delete.")
	properties, _ := requestToProperties(r)
	collectionName := properties["currentCollection"]
	spaceName := properties["currentSpace"]
	lat := properties["lat"]
	latI, err := strconv.Atoi(lat)
	if err != nil {
		panic("Invalid latitude")
	}
	long := properties["long"]
	longI, err := strconv.Atoi(long)
	if err != nil {
		panic("Invalid longitude")
	}
	id := properties["structureId"]
	structureType := properties["structure-type"]

	col, ok := c.Collections[collectionName]
	if !ok {
		panic("No collection")
	}
	space := c.spaceFromNames(collectionName, spaceName)
	if space == nil {
		panic("invalid space")
	}
	structures, ok := col.StructureSets[structureType]
	if !ok {
		panic("No structures for: " + structureType)
	}
	structure, found := getStructureById(structures, id)
	if !found {
		panic("No Structure")
	}
	for i := 0; i < len(structure.FragmentIds); i++ {
		for j := 0; j < len(structure.FragmentIds[i]); j++ {
			if structure.FragmentIds[i][j] != "" {
				areaname := fmt.Sprintf("%s:%d-%d", space.Name, latI+i, longI+j)
				area := getAreaByName(space.Areas, areaname)
				if area == nil {
					continue
				}
				new := make([]Instruction, 0)
				for index := range area.Blueprint.Instructions {
					if area.Blueprint.Instructions[index].GridAssetId == structure.FragmentIds[i][j] {
						// Remove old tiles
						currentRotations := area.Blueprint.Instructions[index].ClockwiseRotations
						grid := col.getTileGridByAssetId(area.Blueprint.Instructions[index].GridAssetId)
						if currentRotations%2 == 1 {
							clearTiles(area.Blueprint.Instructions[index].Y, area.Blueprint.Instructions[index].X, len(grid[0]), len(grid), area.Blueprint.Tiles)
						} else {
							clearTiles(area.Blueprint.Instructions[index].Y, area.Blueprint.Instructions[index].X, len(grid), len(grid[0]), area.Blueprint.Tiles)
						}
					} else {
						new = append(new, area.Blueprint.Instructions[index])
					}
				}
				area.Blueprint.Instructions = new
				for _, instruction := range area.Blueprint.Instructions {
					col.applyInstruction(area.Blueprint.Tiles, instruction)
				}
			}
		}
	}
}

//

func (c Context) spaceMapHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		properties, _ := requestToProperties(r)
		colName := properties["currentCollection"]
		spaceName := properties["currentSpace"]
		fmt.Println("Collection Name: " + colName)
		fmt.Println("Space Name Name: " + spaceName)
		if col, ok := c.Collections[colName]; ok {
			if space, ok := col.Spaces[spaceName]; ok {
				c.generateAllPNGs(space)
			}
		}
		io.WriteString(w, `<img src="/images/map/`+spaceName+`?currentCollection=`+colName+`" width="350" alt="map of space">`)
	}
}

func (c Context) generateAllPNGs(space *Space) { // Should probably return err
	if space.isSimplyTiled() {
		img := c.generateImageFromSpace(space)
		path := c.pathToMapsForSpace(space)
		os.MkdirAll(path, 0755)
		filename := fmt.Sprintf("%s.png", space.Name)
		fullPath := filepath.Join(path, filename)
		err := saveImageAsPNG(fullPath, img)
		if err != nil {
			panic(err)
		}
		c.generatePNGForEachArea(space, img)
	} else {
		fmt.Println("Only Simply tiled topologies are supported")
	}
}

func (c Context) generateImageFromSpace(space *Space) *image.RGBA {
	fmt.Println("Generating Png From space with simple tiling")
	latitude := space.Latitude
	areaHeight := space.AreaHeight
	heightInPixels := latitude * areaHeight
	longitude := space.Longitude
	areaWidth := space.AreaWidth
	widthInPixels := longitude * areaWidth
	col, ok := c.Collections[space.CollectionName]
	if !ok {
		panic("Invalid Collection Name on space: " + space.CollectionName)
	}

	img := image.NewRGBA(image.Rect(0, 0, widthInPixels, heightInPixels))
	for k := 0; k < latitude; k++ {
		for j := 0; j < longitude; j++ {
			area := getAreaByName(space.Areas, fmt.Sprintf("%s:%d-%d", space.Name, k, j))
			if area == nil {
				fmt.Println("no area" + fmt.Sprintf("%s:%d:%d", space.Name, k, j))
				continue
			}
			tinyImg := c.generateImgFromArea(area, *col)
			bounds := tinyImg.Bounds()
			for row := 0; row <= bounds.Dx(); row++ {
				for column := 0; column <= bounds.Dy(); column++ {
					img.Set((j*areaWidth)+column, (k*areaHeight)+row, tinyImg.RGBAAt(column, row))
				}
			}
		}
	}

	return img
}

// move?
func (c Context) generateImgFromArea(area *AreaDescription, col Collection) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, len(area.Blueprint.Tiles[0]), len(area.Blueprint.Tiles)))

	defaultColor := c.findColorByName(area.Blueprint.DefaultTileColor)
	defaultColor1 := c.findColorByName(area.Blueprint.DefaultTileColor1)
	for row := range area.Blueprint.Tiles {
		for column, tile := range area.Blueprint.Tiles[row] {
			outputColor := defaultColor
			ground := groundCellByCoord(area.Blueprint, row, column)
			if ground != nil && ground.Status == 1 {
				outputColor = defaultColor1
			}

			if tile.PrototypeId != "" {
				proto := col.findPrototypeById(tile.PrototypeId)
				if proto == nil {
					fmt.Println("WARN: PROTOTYPE MISSING: " + tile.PrototypeId)
					proto = &Prototype{MapColor: "red"}
				}
				colorString := c.getMapColorFromProto(*proto)
				protoColor := c.findColorByName(colorString) // will be invalid if proto has CommonName == empty
				if protoColor.CssClassName != "NONE" {
					outputColor = protoColor
				}

			}
			img.Set(column, row, color.RGBA{R: uint8(outputColor.R), G: uint8(outputColor.G), B: uint8(outputColor.B), A: 255})
		}
	}

	return img
}

func (c Context) generatePNGForEachArea(space *Space, img *image.RGBA) {
	for k := 0; k < space.Latitude; k++ {
		for j := 0; j < space.Longitude; j++ {
			area := getAreaByName(space.Areas, fmt.Sprintf("%s:%d-%d", space.Name, k, j))
			if area == nil {
				fmt.Println("no area" + fmt.Sprintf("%s:%d-%d", space.Name, k, j))
				continue
			}
			image := addRedSquare(img, k*space.AreaHeight, j*space.AreaWidth, space.AreaHeight, space.AreaWidth)
			filename := filepath.Join(c.pathToMapsForSpace(space), areaToFilename(area))
			err := saveImageAsPNG(filename, image)
			if err != nil {
				panic(err)
			}
		}
	}

}

func areaToFilename(area *AreaDescription) string {
	return strings.ReplaceAll(area.Name, ":", "-") + ".png"
}

func addRedSquare(img *image.RGBA, y0, x0, height, width int) *image.RGBA {
	copy := image.NewRGBA(img.Bounds())
	copy.Pix = append(copy.Pix[:0], img.Pix...)

	for deltaY := 0; deltaY < height; deltaY++ {
		copy.Set(x0, y0+deltaY, color.RGBA{R: 255, A: 255})
		copy.Set(x0+width-1, y0+deltaY, color.RGBA{R: 255, A: 255})
	}
	for deltaX := 0; deltaX < width; deltaX++ {
		copy.Set(x0+deltaX, y0, color.RGBA{R: 255, A: 255})
		copy.Set(x0+deltaX, y0+height-1, color.RGBA{R: 255, A: 255})
	}

	return copy
}

func (c Context) findColorByName(s string) Color {
	for _, color := range c.colors {
		if color.CssClassName == s {
			return color
		}
	}
	return Color{CssClassName: "NONE", R: 0, G: 0, B: 0}
}

func saveImageAsPNG(filename string, img image.Image) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	err = png.Encode(file, img)
	if err != nil {
		return err
	}

	return nil
}

// Utilities

func (space *Space) isSimplyTiled() bool {
	return space.Topology == "torus" || space.Topology == "plane"
}

/*
	// Fuse two spaces
	col, ok := c.Collections["bloop"]
	if !ok {
		panic("no bloop collection")
	}
	blue := col.Spaces["team-blue"]
	fusia := col.Spaces["team-fusia"]
	var areaNameMaker = func(base string) func(y, x int) string {
		return func(y, x int) string {
			return fmt.Sprintf("%s:%d-%d", base, y, x)
		}
	}
	coordToBlue := areaNameMaker("team-blue")
	coordToFusia := areaNameMaker("team-fusia")
	for i := 0; i < 8; i++ {
		// 1st side N/S
		blue.getAreaByName(coordToBlue(0, i)).North = coordToFusia(7, i)
		fusia.getAreaByName(coordToFusia(7, i)).South = coordToBlue(0, i)

		blue.getAreaByName(coordToBlue(7, i)).South = coordToFusia(0, i)
		fusia.getAreaByName(coordToFusia(0, i)).North = coordToBlue(7, i)

		//2nd side E/W
		blue.getAreaByName(coordToBlue(i, 0)).West = coordToFusia(i, 7)
		fusia.getAreaByName(coordToFusia(i, 7)).East = coordToBlue(i, 0)

		blue.getAreaByName(coordToBlue(i, 7)).East = coordToFusia(i, 0)
		fusia.getAreaByName(coordToFusia(i, 0)).West = coordToBlue(i, 7)
	}
*/
