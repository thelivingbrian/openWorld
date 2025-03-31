package main

import (
	"testing"
)

func TestEnsureAreasLoad(t *testing.T) {
	loadFromJson()
	if len(areas) == 0 {
		t.Error("No Materials loaded on attempted start up")
	}
}
