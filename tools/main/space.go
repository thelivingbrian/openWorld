package main

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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

func createSpace(cName, name string, latitude, longitude int, topology string, height, width int, tileColor, tileColor1, weather, broadcastGroup string) Space {
	areas := make([][]AreaDescription, latitude)
	for y := 0; y < latitude; y++ {
		for x := 0; x < longitude; x++ {
			area := createBaseArea(height, width, tileColor, tileColor1, weather, broadcastGroup)

			if topology != "disconnected" {
				area.Name = fmt.Sprintf("%s:%d-%d", name, y, x)
				area.North = fmt.Sprintf("%s:%d-%d", name, mod(y-1, latitude), x)
				area.South = fmt.Sprintf("%s:%d-%d", name, mod(y+1, latitude), x)
				area.East = fmt.Sprintf("%s:%d-%d", name, y, mod(x+1, longitude))
				area.West = fmt.Sprintf("%s:%d-%d", name, y, mod(x-1, longitude))
			}
			areas[y] = append(areas[y], area)
		}
	}

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

	return Space{CollectionName: cName, Name: name, Topology: topology, Latitude: latitude, Longitude: longitude, AreaHeight: height, AreaWidth: width, Areas: flatAreas}
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

func (c Context) generateAllPNGs(space *Space) {
	if space.isSimplyTiled() {
		img := c.generateImageFromSpace(space)
		path := c.pathToMapsForSpace(space)
		_ = os.MkdirAll(path, 0755)
		filename := fmt.Sprintf("%s.png", space.Name)
		fullPath := filepath.Join(path, filename)
		err := saveImageAsPNG(fullPath, img)
		if err != nil {
			panic(err)
		}
		c.generatePNGForEachArea(space, img)
	}
}

func (c Context) generateImageFromSpace(space *Space) *image.RGBA {
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
					proto = &Prototype{MapColor: "red"}
				}
				colorString := c.getMapColorFromProto(*proto)
				protoColor := c.findColorByName(colorString)
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

	return png.Encode(file, img)
}

func (space *Space) isSimplyTiled() bool {
	return space.Topology == "torus" || space.Topology == "plane"
}

func Flatten(s Space) (Space, error) {
	if s.Topology != "plane" && s.Topology != "torus" {
		return s, errors.New("only simply tiled spaces may be flattened")
	}
	if len(s.Areas) == 0 {
		return s, nil
	}

	areaH, areaW := s.AreaHeight, s.AreaWidth
	if areaH <= 0 || areaW <= 0 {
		first := s.Areas[0]
		if first.Blueprint == nil || len(first.Blueprint.Tiles) == 0 || len(first.Blueprint.Tiles[0]) == 0 {
			return s, errors.New("cannot infer AreaHeight/AreaWidth: missing Space.AreaHeight/AreaWidth and first area blueprint is empty")
		}
		areaH = len(first.Blueprint.Tiles)
		areaW = len(first.Blueprint.Tiles[0])
	}

	lat, lon := s.Latitude, s.Longitude
	if lat <= 0 || lon <= 0 {
		maxY, maxX := 0, 0
		for _, a := range s.Areas {
			y, x, ok := parseAreaYX(a.Name)
			if !ok {
				return s, errors.New("cannot infer (Latitude, Longitude); provide them on Space or ensure area names are 'name:Y-X'")
			}
			if y > maxY {
				maxY = y
			}
			if x > maxX {
				maxX = x
			}
		}
		lat = maxY + 1
		lon = maxX + 1
	}

	totalH := areaH * lat
	totalW := areaW * lon

	bigTiles := make([][]TileData, totalH)
	for i := range bigTiles {
		bigTiles[i] = make([]TileData, totalW)
	}

	var bigGround [][]Cell
	var haveGround bool
	var instructions []Instruction
	var transports []Transport
	allSafe := true
	firstArea := s.Areas[0]
	defaultTileColor := ""
	defaultTileColor1 := ""

	initGround := func() {
		if haveGround {
			return
		}
		bigGround = make([][]Cell, totalH)
		for i := range bigGround {
			bigGround[i] = make([]Cell, totalW)
		}
		haveGround = true
	}

	for _, a := range s.Areas {
		if !a.Safe {
			allSafe = false
		}
		if a.Blueprint == nil {
			continue
		}

		if defaultTileColor == "" && a.Blueprint.DefaultTileColor != "" {
			defaultTileColor = a.Blueprint.DefaultTileColor
		}
		if defaultTileColor1 == "" && a.Blueprint.DefaultTileColor1 != "" {
			defaultTileColor1 = a.Blueprint.DefaultTileColor1
		}

		yIdx, xIdx, ok := parseAreaYX(a.Name)
		if !ok {
			return s, fmt.Errorf("area name %q doesn't match `...:Y-X`", a.Name)
		}

		yOff := yIdx * areaH
		xOff := xIdx * areaW

		for r := 0; r < len(a.Blueprint.Tiles); r++ {
			row := a.Blueprint.Tiles[r]
			for c := 0; c < len(row); c++ {
				globalR := yOff + r
				globalC := xOff + c
				if globalR < 0 || globalR >= totalH || globalC < 0 || globalC >= totalW {
					return s, fmt.Errorf("area %q tile (%d,%d) out of flattened bounds", a.Name, globalR, globalC)
				}
				bigTiles[globalR][globalC] = row[c]
			}
		}

		if len(a.Blueprint.Ground) > 0 {
			initGround()
			for r := 0; r < len(a.Blueprint.Ground); r++ {
				row := a.Blueprint.Ground[r]
				for c := 0; c < len(row); c++ {
					globalR := yOff + r
					globalC := xOff + c
					if globalR < 0 || globalR >= totalH || globalC < 0 || globalC >= totalW {
						return s, fmt.Errorf("area %q ground (%d,%d) out of flattened bounds", a.Name, globalR, globalC)
					}
					bigGround[globalR][globalC] = row[c]
				}
			}
		}

		for _, instr := range a.Blueprint.Instructions {
			instr.Y += yOff
			instr.X += xOff
			instructions = append(instructions, instr)
		}

		for _, tr := range a.Transports {
			tr.SourceY += yOff
			tr.SourceX += xOff
			transports = append(transports, tr)
		}
	}

	flatArea := AreaDescription{
		Name:          fmt.Sprintf("%s-flattened:0-0", s.Name),
		Safe:          allSafe,
		MapId:         firstArea.MapId,
		LoadStrategy:  firstArea.LoadStrategy,
		SpawnStrategy: firstArea.SpawnStrategy,
		Weather:       firstArea.Weather,
		Blueprint: &Blueprint{
			Tiles:             bigTiles,
			Instructions:      instructions,
			Ground:            bigGround,
			DefaultTileColor:  defaultTileColor,
			DefaultTileColor1: defaultTileColor1,
		},
		Transports: transports,
	}

	out := s
	out.Name = fmt.Sprintf("%s-flattened", s.Name)
	out.Latitude = 1
	out.Longitude = 1
	out.AreaHeight = totalH
	out.AreaWidth = totalW
	out.Areas = []AreaDescription{flatArea}

	if out.Topology == "torus" {
		out.Areas[0].North = out.Name
		out.Areas[0].South = out.Name
		out.Areas[0].East = out.Name
		out.Areas[0].West = out.Name
	}

	return out, nil
}

func parseAreaYX(name string) (y, x int, ok bool) {
	colon := strings.LastIndex(name, ":")
	if colon == -1 {
		return 0, 0, false
	}
	coord := name[colon+1:]
	parts := strings.Split(coord, "-")
	if len(parts) != 2 {
		return 0, 0, false
	}
	yy, err1 := strconv.Atoi(parts[0])
	xx, err2 := strconv.Atoi(parts[1])
	if err1 != nil || err2 != nil {
		return 0, 0, false
	}
	return yy, xx, true
}
