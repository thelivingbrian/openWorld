package main

import "testing"

func TestFindItemInMatrix(t *testing.T) {
	//matrix := [][]int{}
	matrix := make([][]int, 6)
	for i := range matrix {
		matrix[i] = make([]int, 6)
	}
	matrix[5][5] = 6

	DFS(matrix, 1, 0, 6)
	//DFSRecursive(matrix, nil, 2, 2, 6)
}
