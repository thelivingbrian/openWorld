package main

import (
	"fmt"
	"testing"
)

var stageNames = [2]string{"small", "large"}
var playerCounts = [2]int{1, 100}

func BenchmarkMoveTwice(b *testing.B) {
	loadFromJson()

	for _, stageName := range stageNames {
		testStage, _ := createStageByName(stageName)
		go drainChannel(testStage.updates)
		for _, playerCount := range playerCounts {
			b.Run(fmt.Sprintf("stage:%s players:%d Cores", stageName, playerCount), func(b *testing.B) {
				b.StopTimer() // Stop the timer while setting up the benchmark
				players := placeNPlayersOnStage(playerCount, testStage)
				b.StartTimer() // Start the timer for the actual benchmarking

				for i := 0; i < b.N; i++ {
					//testStage.updateAll("")
					//players[0].nextPower() // This uses a pickup reducing the highlights decreasing time for one player
					players[0].move(-1, 0)
					players[0].move(1, 0)
				}
			})
		}
	}
}

func BenchmarkMoveAllTwice(b *testing.B) {
	loadFromJson()
	for _, stageName := range stageNames {
		testStage, _ := createStageByName(stageName)
		go drainChannel(testStage.updates)
		for _, playerCount := range playerCounts {
			b.Run(fmt.Sprintf("stage:%s players:%d Cores", stageName, playerCount), func(b *testing.B) {
				b.StopTimer()
				players := placeNPlayersOnStage(playerCount, testStage)

				b.StartTimer()
				for i := 0; i < b.N; i++ {
					for index := range players {
						players[index].move(-1, 0)
					}
					for index := range players {
						players[index].move(1, 0)
					}
				}
			})
		}
	}
}

// Move in circles test

// Teleport test

// Helpers
func drainChannel[T any](c chan T) {
	for {
		<-c
	}
}

func placeNPlayersOnStage(n int, stage *Stage) []*Player {
	players := make([]*Player, n)
	for i := range players {
		players[i] = &Player{
			id:        fmt.Sprintf("tp%d", i),
			stage:     stage,
			stageName: stage.name,
			x:         2,
			y:         2,
			actions:   createDefaultActions(),
			health:    100,
		}
		players[i].placeOnStage()
	}
	return players
}

// Add a func to go routine a player moving in backgroud

// Put players on other stages to add additional go routunes/channels
