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

	// races agressively with self unless on one -cpu=1 for reasons I do not understand
	// setParallelism has no impact

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

/*

// Tinfoil hat purposes

func BenchmarkDemoTryLock(b *testing.B) {
	loadFromJson()

	counts := []int{1, 100}

	//b.SetParallelism(1)

	for _, count := range counts {
		b.Run(fmt.Sprintf("stage:%s players:%d Cores", stageName, count), func(b *testing.B) {
			// b.StopTimer()

			// b.StartTimer()
			f := &Foo{}
			lock1, lock2 := sync.Mutex{}, sync.Mutex{}

			for i := 0; i < b.N; i++ {
				for index := 0; index < count; index++ {
					f.tryLockUnlock(&lock1, &lock2)
					f.tryLockUnlock(&lock2, &lock1)
				}
			}
		})
	}
}

type Foo struct {
}

func (*Foo) lockUnlock(lock1, lock2 *sync.Mutex) {
	lock1.Lock()
	defer lock1.Unlock()
	lock2.Lock()
	defer lock2.Unlock()
}

func (*Foo) tryLockUnlock(lock1, lock2 *sync.Mutex) {
	if !lock1.TryLock() {
		fmt.Println("lock1 already locked :( ")
	}
	defer lock1.Unlock()
	if !lock2.TryLock() {
		fmt.Println("lock2 already locked :( ")
	}
	defer lock2.Unlock()
}

*/

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
		bufferClearChannel := make(chan struct{})
		go drainChannel(bufferClearChannel)
		updatesForPlayer := make(chan []byte)
		go drainChannel(updatesForPlayer)
		players[i] = &Player{
			id:                fmt.Sprintf("tp%d", i),
			stage:             stage,
			actions:           createDefaultActions(),
			health:            100,
			updates:           updatesForPlayer,
			clearUpdateBuffer: bufferClearChannel,
			//world:             &World{worldStages: make(map[string]*Stage)},
			tangible: true,
		}
		players[i].placeOnStage(stage, 2, 2)
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
