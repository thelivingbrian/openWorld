package main

import (
	"testing"
)

func TestEnsureNoMaterialsLoad(t *testing.T) {
	loadFromJson()
	if len(materials) != 0 {
		t.Error("No Materials loaded on attempted start up")
	}
}

func TestEnsureMaterialsLoad(t *testing.T) {
	loadFromJson()
	if len(areas) == 0 {
		t.Error("No Materials loaded on attempted start up")
	}
}
