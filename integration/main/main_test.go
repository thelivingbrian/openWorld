package main

import "testing"

func TestValidateLoadTestConfig(t *testing.T) {
	t.Setenv("BLOOP_HOST", "https://example.com")
	t.Setenv("AUTO_PLAYER_PASSWORD", "test-password")

	tests := []struct {
		name    string
		config  loadTestConfig
		wantErr bool
	}{
		{
			name: "valid minimum",
			config: loadTestConfig{
				StageName: "test-stage",
				Count:     1,
				TTL:       1,
			},
		},
		{
			name: "valid maximum",
			config: loadTestConfig{
				StageName: "test-stage",
				Count:     maxLoadTestPlayers,
				TTL:       maxLoadTestTTL,
			},
		},
		{
			name: "valid named style",
			config: loadTestConfig{
				Style: twoTeams512PlayersLoadTestStyle,
				TTL:   1,
			},
		},
		{
			name: "missing stage",
			config: loadTestConfig{
				Count: 1,
				TTL:   1,
			},
			wantErr: true,
		},
		{
			name: "too many players",
			config: loadTestConfig{
				StageName: "test-stage",
				Count:     maxLoadTestPlayers + 1,
				TTL:       1,
			},
			wantErr: true,
		},
		{
			name: "ttl too long",
			config: loadTestConfig{
				StageName: "test-stage",
				Count:     1,
				TTL:       maxLoadTestTTL + 1,
			},
			wantErr: true,
		},
		{
			name: "unsupported named style",
			config: loadTestConfig{
				Style: "unsupported",
				TTL:   1,
			},
			wantErr: true,
		},
		{
			name: "named style and stage",
			config: loadTestConfig{
				Style:     twoTeams512PlayersLoadTestStyle,
				StageName: "test-stage",
				TTL:       1,
			},
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validateLoadTestConfig(test.config)
			if (err != nil) != test.wantErr {
				t.Fatalf("validateLoadTestConfig() error = %v, wantErr %v", err, test.wantErr)
			}
		})
	}
}

func TestTwoTeams512PlayersBatches(t *testing.T) {
	batches := twoTeams512PlayersBatches()
	if len(batches) != 32 {
		t.Fatalf("twoTeams512PlayersBatches() returned %d batches, want 32", len(batches))
	}

	totalPlayers := 0
	stageNames := make(map[string]bool, len(batches))
	for index, batch := range batches {
		totalPlayers += batch.Count
		if batch.Count != 16 {
			t.Errorf("batch %q has %d players, want 16", batch.StageName, batch.Count)
		}
		if stageNames[batch.StageName] {
			t.Errorf("duplicate stage %q", batch.StageName)
		}
		stageNames[batch.StageName] = true

		wantTeam := "fuchsia"
		if index >= 16 {
			wantTeam = "sky-blue"
		}
		if batch.Team != wantTeam {
			t.Errorf("batch %q has team %q, want %q", batch.StageName, batch.Team, wantTeam)
		}
	}

	if totalPlayers != 512 {
		t.Errorf("twoTeams512PlayersBatches() has %d total players, want 512", totalPlayers)
	}

	for _, stageName := range []string{
		"team-blue:0-0",
		"team-fuchsia:1-3",
		"team-blue:2-4",
		"team-fuchsia:3-7",
	} {
		if !stageNames[stageName] {
			t.Errorf("expected stage %q was not generated", stageName)
		}
	}
}

func TestSocketActionForName(t *testing.T) {
	validActions := []string{"", "random", "circles", "lr", "movespace", "space"}
	for _, action := range validActions {
		if _, err := socketActionForName(action); err != nil {
			t.Errorf("socketActionForName(%q) returned error: %v", action, err)
		}
	}

	if _, err := socketActionForName("unsupported"); err == nil {
		t.Error("socketActionForName() accepted an unsupported action")
	}
}
