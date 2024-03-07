package main

import (
	"testing"
)

func TestEnsureDefaultMaterialIsWalkable(t *testing.T) {
	loadFromJson()
	if len(materials) == 0 || materials[0].Walkable == false {
		t.Error("Default material is not walkable")
	}
}
