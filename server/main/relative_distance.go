package main

// [[ycoord, xcoord], ... ]

func jumpCross() [][2]int {
	return [][2]int{{2, 0}, {-2, 0}, {0, 2}, {0, -2}}
}

func cross() [][2]int {
	return [][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}}
}

func x() [][2]int {
	return [][2]int{{1, 1}, {-1, 1}, {1, -1}, {-1, -1}}
}

func applyRelativeDistance(y int, x int, offsets [][2]int) [][2]int {
	output := make([][2]int, len(offsets))
	for i := range offsets {
		output[i] = [2]int{y + offsets[i][0], x + offsets[i][1]}
	}
	return output
}
