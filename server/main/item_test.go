package main

import (
	"fmt"
	"testing"
)

// Pick one and have this specify area.SpawnStrategy
var itemStageNames = [2]string{"clinic", "test-large"}

func BenchmarkSpawnItems(b *testing.B) {
	loadFromJson()

	for _, stageName := range itemStageNames {
		b.Run(fmt.Sprintf("stage:%s Cores", stageName), func(b *testing.B) {
			b.StopTimer()

			testStage := createStageByName(stageName)
			players := placeNPlayersOnStage(200, testStage) // This number has an impact - if true why?

			b.StartTimer()

			for i := 0; i < b.N; i++ {
				spawnItemsFor(players[0], testStage)
			}
		})
	}
}
