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
					move(players[0], -1, 0)
					move(players[0], 1, 0)
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
						move(players[index], -1, 0)
					}
					for index := range players {
						move(players[index], 1, 0)
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

func BenchmarkDivByteBuffer(b *testing.B) {
	colors := []string{"black", "", "blue trsp20"}
	for _, color := range colors {
		b.Run(fmt.Sprintf("stage:%s Cores", color), func(b *testing.B) {
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				generateWeatherSolidByteBuffer(color)
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

func BenchmarkFetchStage(b *testing.B) {
	loadFromJson()
	world, shutDown := createWorldForTesting()
	defer shutDown()
	for _, stagename := range stageNames {

		stage := loadStageByName(world, stagename)
		if stage == nil {
			b.Error("Invalid stage: " + stagename)
		}
		players := placeNPlayersOnStage(1, stage)

		b.Run(fmt.Sprintf("stage:%s Cores", stagename), func(b *testing.B) {
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				// Going to be pointlessly fast because second hit is a map access
				// clear stages after ?
				players[0].fetchStageSync(stagename)
			}

			b.StopTimer()

		})
	}
}

// /////////////////////////////////////////
// Helpers
func drainChannel[T any](c chan T) {
	for {
		_, ok := <-c
		if !ok {
			break
		}
	}
}

func placeNPlayersOnStage(n int, stage *Stage) []*Player {
	players := make([]*Player, n)
	for i := range players {
		updatesForPlayer := make(chan []byte)
		go drainChannel(updatesForPlayer)
		players[i] = &Player{
			id: fmt.Sprintf("tp%d", i),
			//stage:        stage,
			actions: createDefaultActions(),
			//health:       100,
			updates:      updatesForPlayer,
			world:        &World{worldStages: make(map[string]*Stage)},
			tangible:     true,
			playerStages: map[string]*Stage{},
		}
		players[i].health.Store(100)
		players[i].placeOnStage(stage, 2, 2)
	}
	return players
}

// Seems strange this is only a test helper?
func createStageByName(name string) *Stage {
	area, success := areaFromName(name)
	if !success {
		panic("invalid area: " + name)
	}
	return createStageFromArea(area)
}

// Add a func to go routine a player moving in backgroud

// General background channel spam test

// Put players on other stages to add additional go routunes/channels
