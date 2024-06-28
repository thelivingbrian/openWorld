package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"net/http"
	"os"
	"strconv"

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
	// map?
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
	name := queryValues.Get("collectionName")
	if col, ok := c.Collections[name]; ok {
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

		space := createSpace(cName, name, latitude, longitude, topology, height, width, tileColor)
		col.Spaces[name] = &space
		io.WriteString(w, `<h3>Success</h3>`)
		return
	}
	io.WriteString(w, `<h3>Invalid collection Name.</h3>`)
}

func createSpace(cName, name string, latitude, longitude int, topology string, height, width int, tileColor string) Space {
	areas := make([][]AreaDescription, latitude)
	for y := 0; y < latitude; y++ {
		for x := 0; x < longitude; x++ {
			area := createBaseArea(height, width, tileColor)

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

func createBaseArea(height, width int, tileColor string) AreaDescription {
	tiles := make([][]TileData, height)
	for i := range tiles {
		tiles[i] = make([]TileData, width)
	}
	blueprint := Blueprint{Tiles: tiles, Instructions: make([]Instruction, 0)}
	return AreaDescription{Name: "", Safe: true, Blueprint: &blueprint, Transports: make([]Transport, 0), DefaultTileColor: tileColor}
}

func getAreaByName(areas []AreaDescription, name string) *AreaDescription {
	for i, area := range areas {
		if name == area.Name {
			return &areas[i]
		}
	}
	return nil
}

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

func (c Context) spaceMapHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		properties, _ := requestToProperties(r)
		colName := properties["currentCollection"]
		spaceName := properties["currentSpace"]
		fmt.Println("Collection Name: " + colName)
		fmt.Println("Space Name Name: " + spaceName)
		if col, ok := c.Collections[colName]; ok {
			if space, ok := col.Spaces[spaceName]; ok {
				fmt.Println("Space with name exists")
				simpleTiling := space.Topology == "torus" || space.Topology == "plane"
				if simpleTiling {
					img := c.generateImageFromSpace(space)
					path := "./data/collections/" + colName + "/spaces/maps/" + space.Name
					os.MkdirAll(path, 0755)
					err := saveImageAsPNG(path+"/"+space.Name+".png", img)
					if err != nil {
						panic(err)
					}
					c.generatePNGForEachArea(space, img, path)
				} else {
					fmt.Println("Only Simply tiled topologies are supported")
				}
			}
		}
	}
}

func (c Context) generateImageFromSpace(space *Space) *image.RGBA {
	fmt.Println("Generating Png From space with simple tiling")
	latitude := space.Latitude
	areaHeight := space.AreaHeight
	pixelHeight := latitude * areaHeight
	longitude := space.Longitude
	areaWidth := space.AreaWidth
	pixelWidth := longitude * areaWidth
	col, ok := c.Collections[space.CollectionName]
	if !ok {
		panic("Invalid Collection Name on space: " + space.CollectionName)
	}

	img := image.NewRGBA(image.Rect(0, 0, pixelWidth, pixelHeight))
	for k := 0; k < latitude; k++ {
		for j := 0; j < longitude; j++ {
			area := getAreaByName(space.Areas, fmt.Sprintf("%s:%d-%d", space.Name, k, j))
			if area == nil {
				fmt.Println("no area" + fmt.Sprintf("%s:%d:%d", space.Name, k, j))
				continue
			}
			areaColor := c.findColorByName(area.DefaultTileColor)
			//areaImg := *img
			//if &areaImg == img {
			//	fmt.Println("hit")
			//}
			for row := range area.Blueprint.Tiles {
				for column, tile := range area.Blueprint.Tiles[row] {
					proto := col.findPrototypeById(tile.PrototypeId)
					protoColor := c.findColorByName(proto.MapColor)
					if protoColor.CssClassName == "invalid" {
						protoColor = areaColor
					}
					img.Set((j*areaWidth)+column, (k*areaHeight)+row, color.RGBA{R: uint8(protoColor.R), G: uint8(protoColor.G), B: uint8(protoColor.B), A: 255})
				}
			}
		}
	}

	return img
	/*err := saveImageAsPNG("output.png", img)
	if err != nil {
		panic(err)
	}*/
}

func (c Context) generatePNGForEachArea(space *Space, img *image.RGBA, path string) {
	for k := 0; k < space.Latitude; k++ {
		for j := 0; j < space.Longitude; j++ {
			area := getAreaByName(space.Areas, fmt.Sprintf("%s:%d-%d", space.Name, k, j))
			if area == nil {
				fmt.Println("no area" + fmt.Sprintf("%s:%d:%d", space.Name, k, j))
				continue
			}
			image := addRedSquare(img, k*space.AreaHeight, j*space.AreaWidth, space.AreaHeight, space.AreaWidth)
			id := uuid.New().String()
			area.MapId = id
			filename := fmt.Sprintf("%s/%s.png", path, id)
			// Don't need this now, only on deploy?
			// If user generates pngs without saving the space the highlighted maps are effectively unfindable
			saveImageAsPNG(filename, image)
		}
	}

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
	return Color{CssClassName: "invalid"}
}

func saveImageAsPNG(filename string, img image.Image) error {
	fmt.Println("saving")

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
