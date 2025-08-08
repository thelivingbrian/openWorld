package main

import (
	"testing"
)

func TestMoveBoostWithHandlePress(t *testing.T) {
	loadFromJson()
	world, shutDown := createWorldForTesting()
	defer shutDown()
	player := createTestingPlayer(world, "")
	defer close(player.updates)

	testStage := createStageByName("test-walls-interactable-2")
	player.placeOnStage(testStage, 7, 12)

	if player.tile != testStage.tiles[7][12] {
		t.Error("Failed to place on stage")
	}
	if testStage.tiles[3][12].interactable == nil {
		t.Error("Expected interactable at y3 x12")
	}
	if testStage.tiles[3][12].interactable.pushable != true {
		t.Error("Expected interactable at y3 x12 to be pushable")
	}

	// Act
	player.addBoostsAndUpdate(5)
	player.handlePress(eventWithName("W"), "")

	if player.tile != testStage.tiles[5][12] {
		t.Errorf("Incorrect movement. Expected y5-x12 Player at Y:%d, X:%d ", player.tile.y, player.tile.x)
	}
	if testStage.tiles[1][12].interactable != nil {
		t.Error("Interactable should not have moved")
	}

	player.handlePress(eventWithName("W"), "")
	if testStage.tiles[1][12].interactable == nil {
		t.Error("Interactable should have moved")
	}
}

func TestJukeRightFromEveryDirection(t *testing.T) {
	loadFromJson()
	world, shutDown := createWorldForTesting()
	defer shutDown()
	player := createTestingPlayer(world, "")
	defer close(player.updates)

	testStage := createStageByName("test-walls-interactable-2")
	player.placeOnStage(testStage, 10, 11)

	if player.tile != testStage.tiles[10][11] {
		t.Error("Failed to place on stage")
	}
	if testStage.tiles[10][12].interactable == nil {
		t.Error("Expected interactable at y10 x12")
	}
	if testStage.tiles[10][12].interactable.pushable != true {
		t.Error("Expected interactable at y10 x12 to be pushable")
	}

	// Act
	player.handlePress(eventWithName("w"), "")
	if player.tile != testStage.tiles[9][11] {
		t.Errorf("Incorrect movement. Expected y9-x11 Player at Y:%d, X:%d ", player.tile.y, player.tile.x)
	}
	if testStage.tiles[8][11].interactable != nil {
		t.Error("Juke should not occur without previous - Interactable should not have moved")
	}

	player.handlePress(eventWithName("a"), "w")
	player.handlePress(eventWithName("s"), "a")
	if player.tile != testStage.tiles[10][10] {
		t.Errorf("Incorrect movement. Expected y10-x10 Player at Y:%d, X:%d ", player.tile.y, player.tile.x)
	}

	player.handlePress(eventWithName("d"), "s")
	player.handlePress(eventWithName("w"), "d")
	if player.tile != testStage.tiles[9][11] {
		t.Errorf("Incorrect movement. Expected y9-x11 Player at Y:%d, X:%d ", player.tile.y, player.tile.x)
	}
	if testStage.tiles[8][11].interactable == nil {
		t.Error("Interactable should have justed to y8:x10")
	}

	player.handlePress(eventWithName("a"), "w")
	if player.tile != testStage.tiles[9][10] {
		t.Errorf("Incorrect movement. Expected y9-x10 Player at Y:%d, X:%d ", player.tile.y, player.tile.x)
	}
	if testStage.tiles[9][9].interactable == nil {
		t.Error("Interactable should have justed to y9:x9")
	}

	player.handlePress(eventWithName("s"), "a")
	if player.tile != testStage.tiles[10][10] {
		t.Errorf("Incorrect movement. Expected y9-x11 Player at Y:%d, X:%d ", player.tile.y, player.tile.x)
	}
	if testStage.tiles[11][10].interactable == nil {
		t.Error("Interactable should have justed to y11:x10")
	}

	player.handlePress(eventWithName("d"), "s")
	if player.tile != testStage.tiles[10][11] {
		t.Errorf("Incorrect movement. Expected y9-x11 Player at Y:%d, X:%d ", player.tile.y, player.tile.x)
	}
	if testStage.tiles[10][12].interactable == nil {
		t.Error("Interactable should have justed to y10:x12")
	}

	player.handlePress(eventWithName("a"), "d")
	if player.tile != testStage.tiles[10][10] {
		t.Errorf("Incorrect movement. Expected y9-x11 Player at Y:%d, X:%d ", player.tile.y, player.tile.x)
	}
	if testStage.tiles[10][12].interactable == nil {
		t.Error("Juke should not occur when moving backwards")
	}
}

func TestJukeLeftFromEveryDirection(t *testing.T) {
	loadFromJson()
	world, shutDown := createWorldForTesting()
	defer shutDown()
	player := createTestingPlayer(world, "")
	defer close(player.updates)

	testStage := createStageByName("test-walls-interactable-2")
	player.placeOnStage(testStage, 10, 10)

	if player.tile != testStage.tiles[10][10] {
		t.Error("Failed to place on stage")
	}
	if testStage.tiles[10][12].interactable == nil {
		t.Error("Expected interactable at y10 x12")
	}
	if testStage.tiles[10][12].interactable.pushable != true {
		t.Error("Expected interactable at y10 x12 to be pushable")
	}

	// Act
	player.handlePress(eventWithName("d"), "")
	if player.tile != testStage.tiles[10][11] {
		t.Errorf("Incorrect movement. Expected y10-x11 Player at Y:%d, X:%d ", player.tile.y, player.tile.x)
	}

	player.handlePress(eventWithName("s"), "d")
	if player.tile != testStage.tiles[11][11] {
		t.Errorf("Incorrect movement. Expected y11-x11 Player at Y:%d, X:%d ", player.tile.y, player.tile.x)
	}
	if testStage.tiles[12][11].interactable == nil {
		t.Error("Interactable should have justed to y12:x11")
	}

	player.handlePress(eventWithName("a"), "s")
	if player.tile != testStage.tiles[11][10] {
		t.Errorf("Incorrect movement. Expected y11-x10 Player at Y:%d, X:%d ", player.tile.y, player.tile.x)
	}
	if testStage.tiles[11][9].interactable == nil {
		t.Error("Interactable should have justed to y11:x9")
	}

	player.handlePress(eventWithName("w"), "a")
	if player.tile != testStage.tiles[10][10] {
		t.Errorf("Incorrect movement. Expected y10-x10 Player at Y:%d, X:%d ", player.tile.y, player.tile.x)
	}
	if testStage.tiles[9][10].interactable == nil {
		t.Error("Interactable should have justed to y9:x10")
	}

	player.handlePress(eventWithName("d"), "w")
	if player.tile != testStage.tiles[10][11] {
		t.Errorf("Incorrect movement. Expected y10-x11 Player at Y:%d, X:%d ", player.tile.y, player.tile.x)
	}
	if testStage.tiles[10][12].interactable == nil {
		t.Error("Interactable should have justed to y10:x12")
	}

	player.handlePress(eventWithName("a"), "d")
	if player.tile != testStage.tiles[10][10] {
		t.Errorf("Incorrect movement. Expected y10-x10 Player at Y:%d, X:%d ", player.tile.y, player.tile.x)
	}
	if testStage.tiles[10][12].interactable == nil {
		t.Error("Juke should not occur when moving backwards")
	}
}

func TestNoJukeOnNonWalkable(t *testing.T) {
	loadFromJson()
	world, shutDown := createWorldForTesting()
	defer shutDown()
	player := createTestingPlayer(world, "")
	defer close(player.updates)

	testStage := createStageByName("test-walls-interactable-2")
	player.placeOnStage(testStage, 4, 5)

	if player.tile != testStage.tiles[4][5] {
		t.Error("Failed to place on stage")
	}
	if testStage.tiles[3][5].interactable == nil {
		t.Error("Expected interactable at y3 x5")
	}
	if testStage.tiles[3][5].interactable.pushable != true {
		t.Error("Expected interactable at y3 x5 to be pushable")
	}

	player.handlePress(eventWithName("d"), "w")
	if testStage.tiles[3][5].interactable != nil {
		t.Error("Expected interactable at y3 x5 not to have juked")
	}
	if testStage.tiles[4][6].interactable != nil {
		t.Error("Should not juke on walkable")
	}
	if testStage.tiles[4][5].interactable == nil {
		t.Error("Contrained juke should leave interactable under player")
	}
}

/////////////////////////////////////////////////////////
// Helpers

func eventWithName(name string) *PlayerSocketEvent {
	return &PlayerSocketEvent{
		Name: name,
	}
}
