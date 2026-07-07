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
