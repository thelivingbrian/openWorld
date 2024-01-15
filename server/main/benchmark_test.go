package main

import (
	"fmt"
	"testing"
)

func drainChannel[T any](c chan T) {
	for {
		<-c
	}
}

func BenchmarkMarkAllDirty(b *testing.B) {
	loadFromJson()

	stageNames := []string{"small", "large"}
	playerCounts := []int{1, 5, 10, 100}

	for _, stageName := range stageNames {
		testStage, _ := createStageByName(stageName)
		go drainChannel(testStage.updates)
		for _, playerCount := range playerCounts {
			b.Run(fmt.Sprintf("stage:%s players:%d Cores", stageName, playerCount), func(b *testing.B) {
				b.StopTimer() // Stop the timer while setting up the benchmark

				players := make([]Player, playerCount)

				for i := range players {
					players[i] = Player{
						id:        fmt.Sprintf("tp%d", i),
						stage:     testStage,
						stageName: stageName,
						x:         2,
						y:         2,
						actions:   createDefaultActions(),
						health:    100,
					}
					players[i].placeOnStage()
				}

				b.StartTimer() // Start the timer for the actual benchmarking
				for i := 0; i < b.N; i++ {
					testStage.markAllDirty() // Test this with player.action.space on
				}
			})
		}
	}
}

// Move in circles test

// Teleport test
