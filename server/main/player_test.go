package main

import (
	"testing"
)

func TestMoveNorthBoostWithValidNorthernNeighbor(t *testing.T) {
	loadFromJson()
	world := createGameWorld(testdb())
	testStage := createStageByName("hallway")
	updatesForPlayer := make(chan []byte)
	go drainChannel(updatesForPlayer)
	bufferClearChannel := make(chan struct{})
	go drainChannel(bufferClearChannel)

	player := &Player{
		id:                "tp",
		stage:             testStage,
		actions:           createDefaultActions(),
		health:            100,
		updates:           updatesForPlayer,
		clearUpdateBuffer: bufferClearChannel,
		world:             world,
		tangible:          true,
	}
	player.placeOnStage(testStage, 1, 4)

	// Act
	player.addBoosts(5)
	player.moveNorthBoost()

	if player.stage.name != "hallway2" {
		t.Error("player.stage.name should be hallway2")
	}

	if player.getTileSync().y != 7 || player.getTileSync().x != 4 {
		t.Error("Player should be at y:7 x:4")

	}
}

func (p *Player) placeOnStage(stage *Stage, y, x int) {
	placePlayerOnStageAt(p, stage, y, x)
}
