package main

import (
	"fmt"
	"testing"
)

// Break this out somehow.
var itemStageNames = [2]string{"clinic", "test-large"}

func BenchmarkSpawnItems(b *testing.B) {
	loadFromJson()

	for _, stageName := range itemStageNames {
		b.Run(fmt.Sprintf("stage:%s Cores", stageName), func(b *testing.B) {
			b.StopTimer()

			testStage := createStageByName(stageName)
			placeNPlayersOnStage(200, testStage) // This number has an impact

			b.StartTimer()

			for i := 0; i < b.N; i++ {
				testStage.spawnItems()
			}
		})
	}
}
