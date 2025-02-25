package main

// [[ycoord, xcoord], ... ]

func jumpCross() [][2]int {
	return [][2]int{{2, 0}, {-2, 0}, {0, 2}, {0, -2}}
}

func cross() [][2]int {
	return [][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}}
}

func longCross(n int) [][2]int {
	var points [][2]int
	for i := 1; i <= n; i++ {
		points = append(points, [2]int{i, 0})  // right
		points = append(points, [2]int{-i, 0}) // left
		points = append(points, [2]int{0, i})  // up
		points = append(points, [2]int{0, -i}) // down
	}
	return points
}

func x() [][2]int {
	return [][2]int{{1, 1}, {-1, 1}, {1, -1}, {-1, -1}}
}

var grid3x3 [][2]int = createOddGrid(1) // Precompute others? var = [][]int{0,1, . . .}
var grid5x5 [][2]int = createOddGrid(2)
var grid7x7 [][2]int = createOddGrid(3)
var grid9x9 [][2]int = createOddGrid(4)

func createOddGrid(n int) [][2]int {
	var points [][2]int
	for x := -n; x <= n; x++ {
		for y := -n; y <= n; y++ {
			if x != 0 || y != 0 { // Exclude the center point (0, 0)
				points = append(points, [2]int{x, y})
			}
		}
	}
	return points
}

func findOffsetsGivenPowerUp(y int, x int, powerUp *PowerUp) [][2]int {
	output := make([][2]int, 0)
	if powerUp != nil {
		output = applyRelativeDistance(y, x, powerUp.areaOfInfluence)
	}
	return output
}

func applyRelativeDistance(y int, x int, offsets [][2]int) [][2]int {
	output := make([][2]int, len(offsets))
	for i := range offsets {
		output[i] = [2]int{y + offsets[i][0], x + offsets[i][1]}
	}
	return output
}
