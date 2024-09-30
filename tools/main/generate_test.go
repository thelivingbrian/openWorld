package main

import (
	"reflect"
	"testing"
)

func TestGenerateAllPrototypes(t *testing.T) {

}

func TestGridWithCircle(t *testing.T) {
	gridSize := 4
	strategy := ""
	fuzz := 0.1
	result := gridWithCircle(gridSize, strategy, fuzz, 314)

	//  precomputed for seed 314
	expected := [][]Cell{
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

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}
