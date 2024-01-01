package main

import (
	"fmt"
	"testing"
)

func BenchmarkMarkAllDirty(b *testing.B) {
	loadFromJson()
	go func() {
		for {
			<-updates //  Should create a new channel and inject in, using this global one is bad
			// 	Will break after adding second test
		}
	}()

	stageNames := []string{"small", "large"}
	playerCounts := []int{1, 5, 10, 20, 40}

	for _, stageName := range stageNames {
		for _, playerCount := range playerCounts {
			b.Run(fmt.Sprintf("stage:%s players:%d Cores", stageName, playerCount), func(b *testing.B) {
				b.StopTimer() // Stop the timer while setting up the benchmark

				bigStage := createStageByName(stageName)
				players := make([]Player, playerCount)

				for i := range players {
					players[i] = Player{
						id:        fmt.Sprintf("tp%d", i),
						stage:     &bigStage,
						stageName: stageName,
						x:         2,
						y:         2,
						actions:   &Actions{false},
						health:    100,
					}
					placeOnStage(&players[i])
				}

				b.StartTimer() // Start the timer for the actual benchmarking

				for i := 0; i < b.N; i++ {
					bigStage.markAllDirty()
				}
			})
		}
	}
}
