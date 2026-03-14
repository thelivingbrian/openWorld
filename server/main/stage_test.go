package main

import "testing"

func TestEnsureClinicLoads(t *testing.T) {
	loadFromJson()
	stage := createStageByName("clinic")
	if stage == nil {
		t.Error("Clinic stage not loaded - Clinic is needed as backup for non-existant stages")
	}
}

func TestEnsureClinicIsDefault(t *testing.T) {
	loadFromJson()
	world, shutDown := createWorldForTesting()
	defer shutDown()
	player := createTestingPlayer(world, "")
	defer close(player.updates)

	stage := getStageByNameOrGetDefault(player, "non-existant-stage")
	if stage == nil {
		t.Error("Default stage is nil")
		return
	}
	if stage.name != "clinic" {
		t.Error("Default stage is not clinic")
	}
}

func TestCreateStageFromAreaLoadsMutableInteractableState(t *testing.T) {
	area := Area{
		Name: "state-test",
		Tiles: [][]Material{{
			{Walkable: true},
		}},
		Interactables: [][]*InteractableDescription{{
			{
				Name:     "stateful-block",
				State:    "armed",
				CssClass: "red",
				Pushable: true,
				Walkable: true,
			},
		}},
	}

	stage := createStageFromArea(area)
	if stage == nil {
		t.Fatal("expected stage")
	}

	interactable := stage.tiles[0][0].interactable
	if interactable == nil {
		t.Fatal("expected interactable at 0,0")
	}
	if interactable.state != "armed" {
		t.Fatalf("expected interactable state to load as 'armed', got %q", interactable.state)
	}

	interactable.state = "disarmed"
	if stage.tiles[0][0].interactable.state != "disarmed" {
		t.Fatalf("expected interactable state to be mutable and now 'disarmed', got %q", stage.tiles[0][0].interactable.state)
	}
}

// Test personal / individual load types
