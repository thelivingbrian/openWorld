package main

import (
	"fmt"
	"testing"
)

var stageNames = [2]string{"clinic", "test-large"}
var playerCounts = [2]int{1, 100}

// /////////////////////////////////////////
// Movement

func BenchmarkMoveTwice(b *testing.B) {
	loadFromJson()

	for _, stageName := range stageNames {
		testStage := createStageByName(stageName)
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
		testStage := createStageByName(stageName)
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

// /////////////////////////////////////////
// Loading times

func BenchmarkCreateStage(b *testing.B) {
	loadFromJson()
	for _, stageName := range stageNames {
		//channelsToClose := make([]chan Update, 0)
		b.Run(fmt.Sprintf("stage:%s Cores", stageName), func(b *testing.B) {
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				testStage := createStageByName(stageName)
				go drainChannel(testStage.updates)
				//channelsToClose = append(channelsToClose, testStage.updates)
			}

			b.StopTimer()

			//for i := range channelsToClose {
			//	close(channelsToClose[i])
			//}
		})
	}
}

func BenchmarkLoadStage(b *testing.B) {
	loadFromJson()
	world := createGameWorld(testdb())
	for _, stageName := range stageNames {
		//channelsToClose := make([]chan Update, 0)
		b.Run(fmt.Sprintf("stage:%s Cores", stageName), func(b *testing.B) {
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				world.loadStageByName(stageName)
				//go drainChannel(testStage.updates)
				//channelsToClose = append(channelsToClose, testStage.updates)
			}

			b.StopTimer()
			//for i := range channelsToClose {
			//	close(channelsToClose[i])
			//}
		})
	}
}

func BenchmarkGetStage(b *testing.B) {
	loadFromJson()
	world := createGameWorld(testdb())
	for _, stageName := range stageNames {

		//channelsToClose := make([]chan Update, 0)
		s := world.loadStageByName(stageName)

		b.Run(fmt.Sprintf("stage:%s Cores", stageName), func(b *testing.B) {
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				world.getStageByName(stageName)
				//go drainChannel(testStage.updates)
				//channelsToClose = append(channelsToClose, testStage.updates)
			}

			b.StopTimer()

		})

		close(s.updates)
	}
}

// /////////////////////////////////////////
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

// General background channel spam test

// Put players on other stages to add additional go routunes/channels
