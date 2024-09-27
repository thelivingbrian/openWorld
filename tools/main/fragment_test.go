package main

import (
	"fmt"
	"math"
	"math/rand"
	"testing"
)

// Inclusion function using the logistic function
func inclusionProbability(d, r, fuzz float64) float64 {
	return 1 / (1 + math.Exp((d-r)/fuzz))
	/*
		if d < r-fuzz {
			return 1.0
		} else if d > r+fuzz {
			return 0.0
		} else {
			return 1 - (d-(r-fuzz))/(2*fuzz)
		}
	*/
}

type Cell struct {
	status                                     int
	bottomRight, bottomLeft, topRight, topLeft bool
}

type Corner struct {
	a, b, c, d *Cell
}

func TestCircleGeneration(t *testing.T) {
	n := 32             // Size of the grid
	r := float64(n) / 3 // Radius of the circle
	fuzz := 1.7         // Fuzz factor; adjust this to vary sharpness

	// Center of the circle
	cx, cy := float64(n)/2, float64(n)/2

	// Initialize the cell array
	cells := make([][]Cell, n)
	for i := range cells {
		cells[i] = make([]Cell, n)
	}

	// Fill the grid
	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			dx := float64(i) - cx
			dy := float64(j) - cy
			d := math.Hypot(dx, dy)
			p := inclusionProbability(d, r, fuzz)
			if rand.Float64() < p {
				cells[i][j].status = 1
			} else {
				cells[i][j].status = 0
			}
		}
	}

	printCells(cells)

	// Create the Corner array
	corners := make([][]*Corner, n-1)
	for i := 0; i < n-1; i++ {
		corners[i] = make([]*Corner, n-1)
		for j := 0; j < n-1; j++ {
			corners[i][j] = &Corner{
				a: &cells[i][j],
				b: &cells[i][j+1],
				c: &cells[i+1][j],
				d: &cells[i+1][j+1]}
		}
	}

	// gl future me
	// Find the roundness of each cell's corners
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-1; j++ {
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

	for i := 0; i < n; i++ {
		PrintCells(cells[i])
	}
}

func PrintCells(cells []Cell) {
	top, bottom := "", ""
	for _, cell := range cells {
		var tl, tr, bl, br string

		// Determine the corner characters based on status
		//if cell.status == 1 {
		tl = boolToChar(cell.topLeft, "██", "  ")
		tr = boolToChar(cell.topRight, "██", "  ")
		bl = boolToChar(cell.bottomLeft, "██", "  ")
		br = boolToChar(cell.bottomRight, "██", "  ")
		//} else {
		//	tl = boolToChar(cell.topLeft, "  ", "██")
		//	tr = boolToChar(cell.topRight, "  ", "██")
		//	bl = boolToChar(cell.bottomLeft, "  ", "██")
		//	br = boolToChar(cell.bottomRight, "  ", "██")
		//}

		// Print the cell as a 2x2 grid
		//fmt.Println(tl + tr)
		//fmt.Println(bl + br)
		top += tl + tr
		bottom += bl + br
	}
	fmt.Println(top)
	fmt.Println(bottom)

}

// boolToChar returns the first string if condition is true, otherwise the second string
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
