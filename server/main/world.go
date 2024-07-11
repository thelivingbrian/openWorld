package main

import "sync"

type World struct {
	db           *DB
	worldPlayers map[string]*Player
	wPlayerMutex sync.Mutex
	worldStages  map[string]*Stage
	wStageMutex  sync.Mutex
}

type LeaderBoard struct {
	richest       *Player
	wealth        int
	mostDangerous *Player
	streak        int
	oldest        *Player
}

func createGameWorld(db *DB) *World {
	return &World{db: db, worldPlayers: make(map[string]*Player), worldStages: make(map[string]*Stage)}
}
