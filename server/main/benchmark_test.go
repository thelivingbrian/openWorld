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

func BenchmarkDivByte(b *testing.B) {
	colors := []string{"black", "", "blue trsp20"}
	for _, color := range colors {
		b.Run(fmt.Sprintf("stage:%s Cores", color), func(b *testing.B) {
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				generateWeatherSolidBytes(color)
			}

			b.StopTimer()

		})
	}
}
func BenchmarkDivString(b *testing.B) {
	colors := []string{"black", "", "blue trsp20"}
	for _, color := range colors {
		b.Run(fmt.Sprintf("stage:%s Cores", color), func(b *testing.B) {
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				generateWeatherSolid(color)
			}

			b.StopTimer()

		})
	}
}
func BenchmarkDivDumb(b *testing.B) {
	colors := []string{"black", "", "blue trsp20"}
	for _, color := range colors {
		b.Run(fmt.Sprintf("stage:%s Cores", color), func(b *testing.B) {
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				generateWeatherDumb(color)
			}

			b.StopTimer()

		})
	}
}

// /////////////////////////////////////////
// Loading times

func BenchmarkCreateStage(b *testing.B) {
	loadFromJson()
	for _, stageName := range stageNames {
		area, success := areaFromName(stageName)
		if !success {
			panic("invalid area.")
		}
		b.Run(fmt.Sprintf("stage:%s Cores", stageName), func(b *testing.B) {
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				createStageFromArea(area)
			}

			b.StopTimer()

		})
	}
}

func BenchmarkLoadStage(b *testing.B) {
	loadFromJson()
	world := createGameWorld(testdb())
	for _, stageName := range stageNames {

		b.Run(fmt.Sprintf("stage:%s Cores", stageName), func(b *testing.B) {
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				world.loadStageByName(stageName)
			}

			b.StopTimer()
		})
	}
}

func BenchmarkGetStage(b *testing.B) {
	loadFromJson()
	world := createGameWorld(testdb())
	for _, stageName := range stageNames {

		world.loadStageByName(stageName)

		b.Run(fmt.Sprintf("stage:%s Cores", stageName), func(b *testing.B) {
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				world.getStageByName(stageName)
			}

			b.StopTimer()

		})
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
		updatesForPlayer := make(chan []byte)
		players[i] = &Player{
			id:        fmt.Sprintf("tp%d", i),
			stage:     stage,
			stageName: stage.name,
			x:         2,
			y:         2,
			actions:   createDefaultActions(),
			health:    100,
			updates:   updatesForPlayer,
		}
		go drainChannel(players[i].updates)
		players[i].placeOnStage(stage)
	}
	return players
}

// Seems strange this is only a test helper?
func createStageByName(name string) *Stage {
	area, success := areaFromName(name)
	if !success {
		panic("invalid area.")
	}
	return createStageFromArea(area)
}

// Add a func to go routine a player moving in backgroud

// General background channel spam test

// Put players on other stages to add additional go routunes/channels
