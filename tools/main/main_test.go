package main

import "testing"

func TestFindItemInMatrix(t *testing.T) {
	matrix := [][]int{}
	matrix = make([][]int, 8)
	for i := range matrix {
		matrix[i] = make([]int, 5)
	}
	matrix[7][3] = 8

	findValueInMatrix(matrix, 0, 0, 8)
}
