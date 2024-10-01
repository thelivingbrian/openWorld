package main

import (
	"fmt"
	"reflect"
	"testing"
)

func TestGenerateAllPrototypes(t *testing.T) {

}

func TestSmoothCorners_1(t *testing.T) {
	gridSize := 4
	strategy := ""
	fuzz := 0.1
	var seed int64 = 314

	//  precomputed for seed 314
	//  ....
	//  ..XX
	//  .XXX
	//  ..X.
	expectedBefore := [][]Cell{
		{
			{0, false, false, false, false},
			{0, false, false, false, false},
			{0, false, false, false, false},
			{0, false, false, false, false},
		},
		{
			{0, false, false, false, false},
			{0, false, false, false, false},
			{1, false, false, false, false},
			{1, false, false, false, false},
		},
		{
			{0, false, false, false, false},
			{1, false, false, false, false},
			{1, false, false, false, false},
			{1, false, false, false, false},
		},
		{
			{0, false, false, false, false},
			{0, false, false, false, false},
			{1, false, false, false, false},
			{0, false, false, false, false},
		},
	}

	// Edges are always uncurved
	expectedAfter := [][]Cell{
		{
			{0, false, false, false, false},
			{0, false, false, false, false},
			{0, false, false, false, false},
			{0, false, false, false, false},
		},
		{
			{0, false, false, false, false},
			{0, true, false, false, false},
			{1, false, false, false, true},
			{1, false, false, false, false},
		},
		{
			{0, false, false, false, false},
			{1, false, true, false, true},
			{1, false, false, false, false},
			{1, false, false, false, false},
		},
		{
			{0, false, false, false, false},
			{0, false, false, true, false},
			{1, false, false, false, false},
			{0, false, false, false, true},
		},
	}

	cellResult := gridWithCircle(gridSize, strategy, fuzz, seed)

	if !reflect.DeepEqual(cellResult, expectedBefore) {
		t.Errorf("BEFORE RESULT:\n expected:\n\n %v \n\nHave:\n\n %v", expectedBefore, cellResult)
	}

	smoothCornerResult := smoothCorners(cellResult)

	if !reflect.DeepEqual(smoothCornerResult, expectedAfter) {
		t.Errorf("AFTER RESULT:\n expected:\n\n %v \n\nHave:\n\n %v", expectedAfter, smoothCornerResult)
	}
}

func TestSmoothCorners_2(t *testing.T) {
	gridSize := 3
	strategy := ""
	fuzz := 1.0
	var seed int64 = 42

	//  precomputed for seed 42
	//  ...
	//  .XX
	//  XXX
	//
	expectedBefore := [][]Cell{
		{
			{0, false, false, false, false},
			{1, false, false, false, false},
			{0, false, false, false, false},
		},
		{
			{1, false, false, false, false},
			{1, false, false, false, false},
			{1, false, false, false, false},
		},
		{
			{0, false, false, false, false},
			{1, false, false, false, false},
			{1, false, false, false, false},
		},
	}

	// Edges are always uncurved
	expectedAfter := [][]Cell{
		{
			{0, true, false, false, false},
			{1, false, false, false, false},
			{0, false, true, false, false},
		},
		{
			{1, false, false, false, false},
			{1, false, false, false, false},
			{1, false, false, false, false},
		},
		{
			{0, false, false, true, false},
			{1, false, false, false, false},
			{1, false, false, false, false},
		},
	}

	cellResult := gridWithCircle(gridSize, strategy, fuzz, seed)

	if !reflect.DeepEqual(cellResult, expectedBefore) {
		t.Errorf("BEFORE RESULT:\n expected:\n\n %v \n\nHave:\n\n %v", expectedBefore, cellResult)
	}

	smoothCornerResult := smoothCorners(cellResult)

	if !reflect.DeepEqual(smoothCornerResult, expectedAfter) {
		t.Errorf("AFTER RESULT:\n expected:\n\n %v \n\nHave:\n\n %v", expectedAfter, smoothCornerResult)
	}
}

func TestSmoothCorners_3(t *testing.T) {
	gridSize := 3
	strategy := "linear"
	fuzz := 1.0
	var seed int64 = 4

	//  precomputed for seed 42
	//  ...
	//  .XX
	//  XXX
	//
	expectedBefore := [][]Cell{
		{
			{0, false, false, false, false},
			{1, false, false, false, false},
			{0, false, false, false, false},
		},
		{
			{0, false, false, false, false},
			{1, false, false, false, false},
			{1, false, false, false, false},
		},
		{
			{0, false, false, false, false},
			{1, false, false, false, false},
			{0, false, false, false, false},
		},
	}

	// Edges are always uncurved
	expectedAfter := [][]Cell{
		{
			{0, false, false, false, false},
			{1, false, false, false, false},
			{0, false, true, false, false},
		},
		{
			{0, false, false, false, false},
			{1, false, false, false, false},
			{1, false, false, false, false},
		},
		{
			{0, false, false, false, false},
			{1, false, false, false, false},
			{0, false, false, false, true},
		},
	}

	cellResult := gridWithCircle(gridSize, strategy, fuzz, seed)

	if !reflect.DeepEqual(cellResult, expectedBefore) {
		t.Errorf("BEFORE RESULT:\n expected:\n\n %v \n\nHave:\n\n %v", expectedBefore, cellResult)
	}

	smoothCornerResult := smoothCorners(cellResult)

	if !reflect.DeepEqual(smoothCornerResult, expectedAfter) {
		t.Errorf("AFTER RESULT:\n expected:\n\n %v \n\nHave:\n\n %v", expectedAfter, smoothCornerResult)
	}
}

func TestSmoothCorners_Cross(t *testing.T) {

	before := [][]Cell{
		{
			{0, false, false, false, false},
			{1, false, false, false, false},
		},
		{
			{1, false, false, false, false},
			{0, false, false, false, false},
		},
	}

	after := smoothCorners(before)

	// Has two cases, must be one or the other. If neither throw.
	if !(after[0][0].bottomRight && after[1][1].topLeft) && !(after[0][1].bottomLeft && after[1][0].topRight) {
		fmt.Println(after[0][0].bottomRight && after[1][1].topLeft)
		fmt.Println(after[0][1].bottomLeft && after[1][0].topRight)
		t.Errorf("Incorrect smoothing of criss-cross pattern")
	}
}
