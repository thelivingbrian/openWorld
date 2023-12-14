package main

// [[ycoord, xcoord], ... ]

func cross() [][2]int {
	return [][2]int{{2, 0}, {-2, 0}, {0, 2}, {0, -2}}
}

func applyRelativeDistance(y int, x int, offsets [][2]int) [][2]int {
	output := make([][2]int, len(offsets))
	for i := range offsets {
		output[i] = [2]int{y + offsets[i][0], x + offsets[i][1]}
	}
	return output
}
