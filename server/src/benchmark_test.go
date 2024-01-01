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
	fmt.Println("Running test")

	loadFromJson()
	//updates = make(chan Update) // Should add back!
	go func() {
		for {
			<-updates
		}
	}()

	bigStage := createStageByName("big")
	testPlayer := Player{
		id:        "testToken",
		stage:     &bigStage,
		stageName: "big",
		x:         2,
		y:         2,
		actions:   &Actions{false},
		health:    100,
	}

	placeOnStage(&testPlayer)
	for i := 0; i < 50; i++ {
		newPlayer := testPlayer
		newPlayer.id = fmt.Sprintf("tp%d", i)
		placeOnStage(&newPlayer)
	}

	for i := 0; i < b.N; i++ {
		//fmt.Println(i)
		testPlayer.stage.markAllDirty()
	}
}
