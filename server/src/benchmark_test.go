package main

import (
	"fmt"
	"testing"
)

func BenchmarkMarkAllDirty(b *testing.B) {
	loadFromJson()

	//updates = make(chan Update) // Should add back!
	go func() {
		for {
			<-updates
		}
	}()

	sName := "large"
	bigStage := createStageByName(sName)
	testPlayer := Player{
		id:        "testToken",
		stage:     &bigStage,
		stageName: sName,
		x:         2,
		y:         2,
		actions:   &Actions{false},
		health:    100,
	}
	placeOnStage(&testPlayer)

	// Add additional players to stage
	for i := 0; i < 9; i++ {
		newPlayer := testPlayer
		newPlayer.id = fmt.Sprintf("tp%d", i)
		placeOnStage(&newPlayer)
	}

	for i := 0; i < b.N; i++ {
		testPlayer.stage.markAllDirty()
	}
}
