package main

import (
	"fmt"
	"testing"
)

var itemStageNames = [4]string{"small", "large", "big3", "forest:1-1"}

func BenchmarkSpawnItems(b *testing.B) {
	//fmt.Println("hi.")

	loadFromJson()

	//fmt.Println(len(areas))

	for _, stageName := range itemStageNames {
		b.Run(fmt.Sprintf("stage:%s Cores", stageName), func(b *testing.B) {
			b.StopTimer()

			testStage, _ := createStageByName(stageName)
			go drainChannel(testStage.updates)
			placeNPlayersOnStage(200, testStage) // This number has an impact

			b.StartTimer()

			for i := 0; i < b.N; i++ {
				testStage.spawnItems()
			}
		})
	}
}
