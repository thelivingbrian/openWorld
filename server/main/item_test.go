package main

import (
	"fmt"
	"math"
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
				testStage.spawn.Should = nil // Or Always etc
				testStage.spawnItems()
				//calc1() // ~23ns
				//calc2()  // <1 ns
				//fmt.Println("hi") // One fmt.Println adds ~150-200us time
				//fmt.Println("hi")
			}
		})
	}
}

func calc1() float64 {
	return 1.0 / math.Pow(4.0, float64(7+-1))
}

func calc2() float64 {
	denominator := 1 << (2 * (7 + -1))
	return 1.0 / float64(denominator)
}
