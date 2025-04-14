package main

import (
	"testing"
	"time"
)

type MockRankProvider struct {
	topNCalls map[string]int
	players   map[string][]PlayerRecord
}

func (m *MockRankProvider) getTopNPlayersByField(field string, n int) ([]PlayerRecord, error) {
	m.topNCalls[field]++
	return m.players[field], nil
}

func TestGenerateRichestList(t *testing.T) {
	mock := &MockRankProvider{
		topNCalls: make(map[string]int),
		players: map[string][]PlayerRecord{
			"money": {
				{Username: "Alice", Money: 1000},
				{Username: "Bob", Money: 800},
			},
		},
	}
	hub := createDefaultHub(mock)

	// First call should trigger DB access
	list := generateRichestList(hub)
	if len(list.Entries) != 2 || list.Entries[0].Username != "Alice" {
		t.Errorf("unexpected high score list: %+v", list.Entries)
	}
	if mock.topNCalls["money"] != 1 {
		t.Errorf("expected DB call, got: %d", mock.topNCalls["money"])
	}

	// Second call within cache interval should not trigger DB access
	list = generateRichestList(hub)
	if mock.topNCalls["money"] != 1 {
		t.Errorf("expected cache hit, but DB was called again")
	}

	// Force cache expiration
	hub.richest.lastChecked = time.Now().Add(-20 * time.Second)
	list = generateRichestList(hub)
	if mock.topNCalls["money"] != 2 {
		t.Errorf("expected second DB call after expiry")
	}
}

func TestGenerateDeadliestList(t *testing.T) {
	mock := &MockRankProvider{
		topNCalls: make(map[string]int),
		players: map[string][]PlayerRecord{
			"killCount": {
				{Username: "Charlie", KillCount: 10, DeathCount: 2},
				{Username: "Dana", KillCount: 5, DeathCount: 1},
			},
		},
	}
	hub := createDefaultHub(mock)

	list := generateDeadliestList(hub)
	if len(list.Entries) != 2 || list.Entries[0].Username != "Charlie" {
		t.Errorf("unexpected list: %+v", list.Entries)
	}
	expectedKD := "5.00"
	if list.Entries[0].StatValues[1] != expectedKD {
		t.Errorf("expected K/D %s, got %s", expectedKD, list.Entries[0].StatValues[1])
	}
}
