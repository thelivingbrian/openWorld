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

func TestCreateStageFromAreaAppliesSelectedStateConfiguration(t *testing.T) {
	area := Area{
		Name: "state-config-test",
		Tiles: [][]Material{{
			{Walkable: true},
		}},
		Interactables: [][]*InteractableDescription{{
			{
				Name:         "stateful-door",
				State:        "open",
				DefaultState: "closed",
				States: map[string]InteractableStateDescription{
					"closed": {
						CssClass:      "door-closed",
						Pushable:      false,
						Walkable:      false,
						Fragile:       false,
						Reactions:     "",
						ReactionRules: nil,
					},
					"open": {
						CssClass:      "door-open",
						Pushable:      true,
						Walkable:      true,
						Fragile:       true,
						Reactions:     "pass-all",
						ReactionRules: nil,
					},
				},
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

	if interactable.state != "open" {
		t.Fatalf("expected selected state 'open', got %q", interactable.state)
	}
	if interactable.cssClass != "door-open" {
		t.Fatalf("expected cssClass from selected state, got %q", interactable.cssClass)
	}
	if !interactable.pushable || !interactable.walkable || !interactable.fragile {
		t.Fatalf("expected selected state booleans to apply (pushable/walkable/fragile), got %v/%v/%v", interactable.pushable, interactable.walkable, interactable.fragile)
	}
	if len(interactable.reactions) == 0 {
		t.Fatal("expected selected state reactions to be resolved")
	}

	interactable.state = ""
	if !interactable.applyState(interactable.defaultState) {
		t.Fatal("expected applyState(defaultState) to succeed")
	}
	if interactable.cssClass != "door-closed" || interactable.pushable || interactable.walkable || interactable.fragile {
		t.Fatalf("expected default state to restore closed configuration, got css=%q pushable=%v walkable=%v fragile=%v", interactable.cssClass, interactable.pushable, interactable.walkable, interactable.fragile)
	}
}

// Test personal / individual load types
