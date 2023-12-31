package main

import (
	"fmt"
	"testing"
)

func addSomeStrings(s string) string {
	output := ""
	for i := 0; i < 5; i++ {
		output += s
	}
	return output
}

func BenchmarkUpdateFullScreen(b *testing.B) {
	loadFromJson()
	bigStage := createStageByName("big")
	testPlayer := &Player{
		id:        "testToken",
		stage:     &bigStage,
		stageName: "big",
		x:         2,
		y:         2,
		actions:   &Actions{false},
		health:    100,
	}
	testPlayer.x = 7

	placeOnStage(testPlayer)

	for i := 0; i < b.N; i++ {
		//addSomeStrings("hi")
		fmt.Println(i)
		updateFullScreen(testPlayer) // Calling SomeFunction from src package
	}
	//updateFullScreen()
}
