package main

import (
	"testing"
)

func TestEnsureMaterialsLoad(t *testing.T) {
	loadFromJson()
	if len(materials) == 0 {
		t.Error("No Materials loaded on attempted start up")
	}
}
