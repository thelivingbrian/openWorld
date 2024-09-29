package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"testing"

	"github.com/google/uuid"
)

type Cell struct {
	status                                     int
	bottomRight, bottomLeft, topRight, topLeft bool
}

type Corner struct {
	a, b, c, d *Cell
}

func TestGenerateAllPrototypes(t *testing.T) {

	cells := GenerateCircle(16, 1.7)
	fragment, prototypes := makeAssetsFromCells(cells)

	fragments := make([]Fragment, 0)
	fragments = append(fragments, fragment)

	outFile := "./data/collections/bloop/fragments/ground-patterns.json"
	err := writeJsonFile(outFile, fragments)
	if err != nil {
		panic(err)
	}

	outFile2 := "./data/collections/bloop/proc/prototypes/floors.json"
	err = writeJsonFile(outFile2, prototypes)
	if err != nil {
		panic(err)
	}

}

func insertDiff(protoSource, protoQuery []Prototype) []Prototype {
	for i := range protoQuery {
		if !hasMatchingId(protoSource, protoQuery[i].ID) {
			protoSource = append(protoSource, protoQuery[i])
		}
	}
	return protoSource
}

func hasMatchingId(list []Prototype, id string) bool {
	for i := range list {
		if list[i].ID == id {
			return true
		}
	}
	return false
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

func makeAssetsFromCells(cells [][]Cell) (Fragment, []Prototype) {
	color1, color2 := "grass", "sand"

	color1OnTop := makePrototypeVariations(color1, color2)
	color2OnTop := makePrototypeVariations(color2, color1)

	tiles := make([][]TileData, len(cells))
	for i := range tiles {
		tiles[i] = make([]TileData, len(cells[i]))
		for j := range tiles[i] {
			id := "BLAH"
			cell := &cells[i][j]
			if cell.status == 1 {
				id = color2OnTop[roundednessToInt(cell.topLeft, cell.topRight, cell.bottomLeft, cell.bottomRight)].ID
			} else {
				id = color1OnTop[roundednessToInt(cell.topLeft, cell.topRight, cell.bottomLeft, cell.bottomRight)].ID
			}
			tiles[i][j] = TileData{PrototypeId: id}
		}
	}
	bp := Blueprint{Tiles: tiles}
	fragment := Fragment{ID: uuid.NewString(), Name: "test-frag", SetName: "ground-patterns", Blueprint: &bp}
	prototypes := insertDiff(color1OnTop, color2OnTop)
	return fragment, prototypes
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
	protos[0] = Prototype{ID: "", Floor1Css: top, MapColor: top, SetName: "floors"}
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

func (p *Prototype) assignMd5() {
	id, err := md5ForPrototype(*p)
	if err != nil {
		panic(err)
	}
	p.ID = id
}

func GenerateCircle(gridSize int, fuzz float64) [][]Cell {
	radius := float64(gridSize) / 3

	// Center of the circle
	cx, cy := float64(gridSize)/2, float64(gridSize)/2

	// Initialize the cell array
	cells := make([][]Cell, gridSize)
	for i := range cells {
		cells[i] = make([]Cell, gridSize)
	}

	// Fill the grid
	for i := 0; i < gridSize; i++ {
		for j := 0; j < gridSize; j++ {
			dx := float64(i) - cx
			dy := float64(j) - cy
			d := math.Hypot(dx, dy)
			p := logisticProbability(d, radius, fuzz)
			if rand.Float64() < p {
				cells[i][j].status = 1
			} else {
				cells[i][j].status = 0
			}
		}
	}

	printCells(cells)

	// Create the Corner array
	corners := make([][]*Corner, gridSize-1)
	for i := 0; i < gridSize-1; i++ {
		corners[i] = make([]*Corner, gridSize-1)
		for j := 0; j < gridSize-1; j++ {
			corners[i][j] = &Corner{
				a: &cells[i][j],
				b: &cells[i][j+1],
				c: &cells[i+1][j],
				d: &cells[i+1][j+1]}
		}
	}

	// gl future me
	// Find the roundness of each cell's corners
	for i := 0; i < gridSize-1; i++ {
		for j := 0; j < gridSize-1; j++ {
			corner := corners[i][j]
			count := corner.a.status + corner.b.status + corner.c.status + corner.d.status
			if count == 4 || count == 0 {
				corner.a.bottomRight = false
				corner.b.bottomLeft = false
				corner.c.topRight = false
				corner.d.topLeft = false
			} else if count == 3 {
				corner.a.bottomRight = !(corner.a.status == 1)
				corner.b.bottomLeft = !(corner.b.status == 1)
				corner.c.topRight = !(corner.c.status == 1)
				corner.d.topLeft = !(corner.d.status == 1)
			} else if count == 1 {
				corner.a.bottomRight = (corner.a.status == 1)
				corner.b.bottomLeft = (corner.b.status == 1)
				corner.c.topRight = (corner.c.status == 1)
				corner.d.topLeft = (corner.d.status == 1)
			} else if count == 2 {
				if corner.a.status == corner.b.status || corner.a.status == corner.c.status {
					corner.a.bottomRight = false
					corner.b.bottomLeft = false
					corner.c.topRight = false
					corner.d.topLeft = false
				} else { // corner.a.status is equal to corner.d status
					if rand.Float64() < .5 {
						corner.a.bottomRight = true
						corner.b.bottomLeft = false
						corner.c.topRight = false
						corner.d.topLeft = true
					} else {
						corner.a.bottomRight = false
						corner.b.bottomLeft = true
						corner.c.topRight = true
						corner.d.topLeft = false
					}
				}
			}
		}
	}

	printCellCorners(cells)

	return cells
}

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
