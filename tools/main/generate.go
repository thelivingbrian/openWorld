package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
)

type Cell struct {
	Status                                     int
	BottomRight, BottomLeft, TopRight, TopLeft bool
}

// Still needed?
func (col *Collection) generateAndSaveGroundPattern(config GroundConfig) {
	cells := GenerateCircle(config.Span, config.Strategy, config.Fuzz)
	prototypes, fragments, structure := makeAssetsForGround(cells, config)

	col.PrototypeSets["floors"] = merge(col.PrototypeSets["floors"], prototypes, IdsMatchProto)
	outFileProto := "./data/collections/bloop/prototypes/floors.json"
	err := writeJsonFile(outFileProto, col.PrototypeSets["floors"], true)
	if err != nil {
		panic(err)
	}

	col.Fragments["ground-patterns"] = merge(fragments, col.Fragments["ground-patterns"], IdsMatchFragment)
	outFileFragment := "./data/collections/bloop/fragments/ground-patterns.json"
	err = writeJsonFile(outFileFragment, col.Fragments["ground-patterns"], false)
	if err != nil {
		panic(err)
	}

	col.StructureSets["ground"] = merge(col.StructureSets["ground"], append(make([]Structure, 0), structure), IdsMatchStructure)
	outFileStruct := "./data/collections/bloop/structures/ground.json"
	err = writeJsonFile(outFileStruct, col.StructureSets["ground"], false)
	if err != nil {
		panic(err)
	}
}

///////////////////////////////////////////////////////
// Proceedural Prototype Management

func (p *Prototype) assignMd5() {
	id, err := md5ForPrototype(*p)
	if err != nil {
		panic(err)
	}
	p.ID = id
}

func md5ForPrototype(p Prototype) (string, error) {
	p.ID = "" // This prevents recursive match prevention
	jsonData, err := json.Marshal(p)
	if err != nil {
		return "", err
	}

	// Generate MD5 hash and convert to hex
	hash := md5.Sum(jsonData)
	return hex.EncodeToString(hash[:]), nil
}

func merge[T any](tSource, tQuery []T, equal func(T, T) bool) []T {
	out := append(make([]T, 0), tSource...)
	for i := range tQuery {
		if !contains(tSource, tQuery[i], equal) {
			out = append(out, tQuery[i])
		}
	}
	return out
}

func contains[T any](tList []T, tItem T, equal func(T, T) bool) bool {
	for i := range tList {
		if equal(tList[i], tItem) {
			return true
		}
	}
	return false
}

func IdsMatchProto(p1, p2 Prototype) bool {
	return p1.ID == p2.ID
}

func IdsMatchFragment(f1, f2 Fragment) bool {
	return f1.ID == f2.ID
}

func IdsMatchStructure(s1, s2 Structure) bool {
	return s1.ID == s2.ID
}

///////////////////////////////////////////////////////
// Ground Pattern Generation

type Corner struct {
	a, b, c, d *Cell
}

func GenerateCircle(span int, strategy string, fuzz float64) [][]Cell {
	cells := smoothCorners(gridWithCircle(span*16, strategy, fuzz, 0))
	return cells
}

func gridWithCircle(gridSize int, strategy string, fuzz float64, seed int64) [][]Cell {
	if seed == 0 {
		seed = rand.Int63()
	}
	r := rand.New(rand.NewSource(seed))

	cells := make([][]Cell, gridSize)
	for i := range cells {
		cells[i] = make([]Cell, gridSize)
	}
	radius := float64(gridSize) * .4
	probability := logisticProbability
	if strategy == "linear" {
		probability = linearProbability
	}
	cx, cy := float64(gridSize)/2, float64(gridSize)/2
	for i := 0; i < gridSize; i++ {
		for j := 0; j < len(cells[i]); j++ {
			dx := float64(i) - cx
			dy := float64(j) - cy
			d := math.Hypot(dx, dy)
			p := probability(d, radius, fuzz)
			if r.Float64() < p {
				cells[i][j].Status = 1
			} else {
				cells[i][j].Status = 0
			}
		}
	}
	//printCells(cells)
	return cells
}

func smoothCorners(cells [][]Cell) [][]Cell {
	gridHeight := len(cells)

	// Create the Corner array
	corners := make([][]*Corner, gridHeight-1)
	for i := 0; i < gridHeight-1; i++ {
		gridWidth := len(cells[i])
		corners[i] = make([]*Corner, gridWidth-1)
		for j := 0; j < gridWidth-1; j++ {
			corners[i][j] = &Corner{
				a: &cells[i][j],
				b: &cells[i][j+1],
				c: &cells[i+1][j],
				d: &cells[i+1][j+1]}
		}
	}

	// Find the roundness of each cell's corners
	for i := 0; i < len(corners); i++ {
		for j := 0; j < len(corners[i]); j++ {
			corner := corners[i][j]
			count := corner.a.Status + corner.b.Status + corner.c.Status + corner.d.Status
			if count == 4 || count == 0 {
				corner.a.BottomRight = false
				corner.b.BottomLeft = false
				corner.c.TopRight = false
				corner.d.TopLeft = false
			} else if count == 3 {
				corner.a.BottomRight = !(corner.a.Status == 1)
				corner.b.BottomLeft = !(corner.b.Status == 1)
				corner.c.TopRight = !(corner.c.Status == 1)
				corner.d.TopLeft = !(corner.d.Status == 1)
			} else if count == 1 {
				corner.a.BottomRight = (corner.a.Status == 1)
				corner.b.BottomLeft = (corner.b.Status == 1)
				corner.c.TopRight = (corner.c.Status == 1)
				corner.d.TopLeft = (corner.d.Status == 1)
			} else if count == 2 {
				if corner.a.Status == corner.b.Status || corner.a.Status == corner.c.Status {
					corner.a.BottomRight = false
					corner.b.BottomLeft = false
					corner.c.TopRight = false
					corner.d.TopLeft = false
				} else { // corner.a.status is equal to corner.d status
					if rand.Float64() < .5 {
						corner.a.BottomRight = true
						corner.b.BottomLeft = false
						corner.c.TopRight = false
						corner.d.TopLeft = true
					} else {
						corner.a.BottomRight = false
						corner.b.BottomLeft = true
						corner.c.TopRight = true
						corner.d.TopLeft = false
					}
				}
			}
		}
	}

	//printCellCorners(cells)
	return cells
}

/*
// Debugging Print Logs

func printCellCorners(cells [][]Cell) {
	for i := range cells {
		top, bottom := "", ""
		for _, cell := range cells[i] {
			tl := boolToChar(cell.topLeft, "██", "  ")
			tr := boolToChar(cell.topRight, "██", "  ")
			bl := boolToChar(cell.bottomLeft, "██", "  ")
			br := boolToChar(cell.bottomRight, "██", "  ")

			top += tl + tr
			bottom += bl + br
		}
		fmt.Println(top)
		fmt.Println(bottom)
	}
}

func boolToChar(condition bool, trueChar, falseChar string) string {
	if condition {
		return trueChar
	}
	return falseChar
}

func printCells(grid [][]Cell) {
	n := len(grid)
	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			if grid[i][j].status == 1 {
				fmt.Print("██")
			} else {
				fmt.Print("  ")
			}
		}
		fmt.Println()
	}
}
*/

// //////////////////////////////////////////////////
//  Inclusion probability functions

// Inclusion function using the logistic function
var logisticProbability = func(d, r, fuzz float64) float64 {
	// If fuzz is positive, probability follows sigmoid graph from 1 to 0
	// as (d - r) e.g. signed distance goes from very negative to very far away
	return 1 / (1 + math.Exp((d-r)/fuzz))
}

// Inclusion function with linear dropoff
var linearProbability = func(d, r, fuzz float64) float64 {
	if d < r-fuzz {
		return 1.0
	} else if d > r+fuzz {
		return 0.0
	} else {
		return 1 - (d-(r-fuzz))/(2*fuzz)
	}
}

func makeAssetsForGround(cells [][]Cell, config GroundConfig) ([]Prototype, []Fragment, Structure) {
	color1OnTop := makePrototypeVariations(config.Color1, config.Color2)
	color2OnTop := makePrototypeVariations(config.Color2, config.Color1)

	tiles := make([][]TileData, len(cells))
	for i := range tiles {
		tiles[i] = make([]TileData, len(cells[i]))
		for j := range tiles[i] {
			id := "BLAH"
			cell := &cells[i][j]
			if cell.Status == 1 {
				id = color2OnTop[roundednessToInt(cell.TopLeft, cell.TopRight, cell.BottomLeft, cell.BottomRight)].ID
			} else {
				id = color1OnTop[roundednessToInt(cell.TopLeft, cell.TopRight, cell.BottomLeft, cell.BottomRight)].ID
			}
			tiles[i][j] = TileData{PrototypeId: id}
		}
	}

	size := len(cells) / 16                 // assumes a square
	blueprints := make([][]Blueprint, size) // only supports 16x16 extra will be ignored
	fragments := make([]Fragment, size*size)
	structure := Structure{ID: config.Name, FragmentIds: make([][]string, size), GroundConfig: &config}
	for a := 0; a < size; a++ {
		blueprints[a] = make([]Blueprint, size)
		structure.FragmentIds[a] = make([]string, size)
		for b := 0; b < size; b++ {
			blueprints[a][b] = Blueprint{Tiles: make([][]TileData, 16)}
			name := fmt.Sprintf("%s-%d-%d", config.Name, a, b)
			hash := md5.Sum([]byte(name))
			id := hex.EncodeToString(hash[:])
			for i := 0; i < 16; i++ {
				blueprints[a][b].Tiles[i] = tiles[(a*16)+i][b*16 : (b+1)*16]
			}
			fragments[(a*size)+b] = Fragment{ID: id, Name: name, SetName: "ground-patterns", Blueprint: &blueprints[a][b]}
			structure.FragmentIds[a][b] = id
		}
	}

	prototypes := merge(color1OnTop, color2OnTop, IdsMatchProto)
	return prototypes, fragments, structure
}

func roundednessToInt(tl, tr, bl, br bool) int {
	return (boolToInt(tl) << 3) | (boolToInt(tr) << 2) | (boolToInt(bl) << 1) | boolToInt(br)
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func makePrototypeVariations(top, bottom string) []Prototype {
	protos := make([]Prototype, 16)
	protos[0] = Prototype{ID: "", Floor1Css: top, MapColor: top, SetName: "floors", Walkable: true}
	protos[0].assignMd5()
	for i := 1; i < 16; i++ {
		tl, tr, bl, br := "", "", "", ""
		if i&8 != 0 {
			tl = " r0-{rotate:tl}"
		}
		if i&4 != 0 {
			tr = " r0-{rotate:tr}"
		}
		if i&2 != 0 {
			bl = " r0-{rotate:bl}"
		}
		if i&1 != 0 {
			br = " r0-{rotate:br}"
		}
		protos[i] = Prototype{
			ID:        "",
			Floor1Css: bottom,
			Floor2Css: top + tl + tr + bl + br,
			MapColor:  top,
			Walkable:  true,
			SetName:   "proc-floors",
		}
		protos[i].assignMd5()
	}
	return protos
}
