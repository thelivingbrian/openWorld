package main

import "testing"

func TestEnsureCharacterCanMoveOnWalkable(t *testing.T) {
	loadFromJson()
	testStage := createStageByName("hallway")
	startingTile := testStage.tiles[3][11]
	if walkable(testStage.tiles[3][13]) {
		t.Error("This Hallway tile (y:3 x:13) should not be walkable")
	}

	npc := &NonPlayer{
		id: "test-npc",
	}

	addNPCAndNotifyOthers(npc, startingTile)
	if npc.tile != startingTile {
		t.Error("NPC did not spawn at correct location")
	}

	moveEast(npc)
	if npc.tile != testStage.tiles[3][12] {
		t.Error("NPC did not move east")
	}

	moveEast(npc)
	if npc.tile != testStage.tiles[3][12] {
		t.Error("NPC should not walk on walls")
	}

	moveSouth(npc)
	if npc.tile != testStage.tiles[4][12] {
		t.Error("NPC did not move south")
	}
}

func TestEnsureCharacterCanRotateInteractables(t *testing.T) {
	loadFromJson()
	testStage := createStageByName("hallway")
	startingTile := testStage.tiles[3][7]
	npc := &NonPlayer{
		id: "test-npc",
	}
	addNPCAndNotifyOthers(npc, startingTile)
	north := testStage.tiles[2][7]
	south := testStage.tiles[4][7]
	east := testStage.tiles[3][8]
	west := testStage.tiles[3][6]

	int0 := &Interactable{
		name:     "test-interactable-0",
		pushable: true,
		cssClass: "red",
	}
	int1 := &Interactable{
		name:     "test-interactable-1",
		pushable: true,
		cssClass: "blue",
	}
	int2 := &Interactable{
		name:     "test-interactable-2",
		pushable: true,
		cssClass: "green",
	}
	int3 := &Interactable{
		name:     "test-interactable-3",
		pushable: true,
		cssClass: "yellow",
	}
	int4 := &Interactable{
		name:     "test-interactable-4",
		pushable: false,
		cssClass: "black",
	}
	int5 := &Interactable{
		name:     "test-interactable-5",
		pushable: false,
		cssClass: "white",
	}

	north.interactable = int0

	rotate(npc, true)
	if north.interactable != nil && east.interactable != int0 {
		t.Error("NPC should have rotated the interactable clockwise")
	}

	rotate(npc, true)
	if east.interactable != nil && south.interactable != int0 {
		t.Error("NPC should have rotated the interactable clockwise twice")
	}

	rotate(npc, true)
	if south.interactable != nil && west.interactable != int0 {
		t.Error("NPC should have rotated the interactable clockwise x3")
	}

	rotate(npc, true)
	if west.interactable != nil && north.interactable != int0 {
		t.Error("NPC should have rotated the interactable clockwise x4")
	}

	south.interactable = int2
	rotate(npc, false)
	if south.interactable != nil && north.interactable != nil && east.interactable != int0 && west.interactable != int2 {
		t.Error("NPC should have rotated both interactables counter-clockwise")
	}

	rotate(npc, false)
	if south.interactable != int0 && north.interactable != int2 && east.interactable != nil && west.interactable != nil {
		t.Error("NPC should have rotated both interactables counter-clockwise again")
	}

	west.interactable = int1
	rotate(npc, true)
	rotate(npc, true)

	if north.interactable != int0 && east.interactable != int1 && south.interactable != int2 && west.interactable != nil {
		t.Error("NPC should have rotated all interactables clockwise")
	}

	west.material.Walkable = false
	rotate(npc, true)
	if north.interactable != int0 && east.interactable != int1 && south.interactable != int2 && west.interactable != nil {
		t.Error("Should not rotate through walls")
	}

	west.material.Walkable = true
	west.interactable = int3
	rotate(npc, true)
	if north.interactable != int3 && east.interactable != int0 && south.interactable != int1 && west.interactable != int2 {
		t.Error("NPC should have rotated all four interactables clockwise")
	}

	rotate(npc, false)
	rotate(npc, false)
	if west.interactable != int0 && north.interactable != int1 && east.interactable != int2 && south.interactable != int3 {
		t.Error("NPC should have rotated all four interactables counter-clockwise twice")
	}

	north.interactable = nil
	east.interactable = nil
	south.interactable = int4 // not pushable

	rotate(npc, true)
	rotate(npc, true)
	rotate(npc, true)
	if south.interactable != int4 && north.interactable != nil && east.interactable != int0 && west.interactable != nil {
		t.Error("Non-pushable not obstructing properly")
	}

	north.interactable = int5 // not pushable
	rotate(npc, true)
	if north.interactable != int5 {
		t.Error("Non-pushable should not rotate")
	}
}
