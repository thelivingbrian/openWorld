package main

import "testing"

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
		stageName:         testStage.name,
		x:                 4,
		y:                 1,
		actions:           createDefaultActions(),
		health:            100,
		updates:           updatesForPlayer,
		clearUpdateBuffer: bufferClearChannel,
		world:             world,
	}
	player.placeOnStage(testStage)

	// Act
	player.addBoosts(5)
	player.moveNorthBoost()

	// Assert
	if player.stageName != "hallway2" {
		t.Error("player stageName should be hallway2 but is: " + player.stageName)
	}

	if player.stage.name != "hallway2" {
		t.Error("player.stage.name should be hallway2")
	}

	if player.y != 7 || player.x != 4 {
		t.Error("Player should be at y:7 x:4")

	}
}

func (p *Player) placeOnStage(stage *Stage) {
	placePlayerOnStageAt(p, stage, p.y, p.x)
}
