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
