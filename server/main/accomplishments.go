package main

import (
	"sync"
	"time"
)

type SyncAccomplishmentList struct {
	sync.Mutex
	Accomplishments map[string]Accomplishment
}

type Accomplishment struct {
	AcquiredAt time.Time
	Name       string
	// Event id? - which currently only is/could be mongo _id
}

var everyAccomplishment = []string{
	"Become most dangerous",
	"Score a goal",
	"Defeat another player",
	"10 Kill streak",
	"100 Kill streak",
	"1,000 money",
	"50,000 money",
	"Double npc kill - simultaneous",
	"Triple npc kill - simultaneous",
	"Puzzle 0",
}

func (accomplishments *SyncAccomplishmentList) addByName(name string) *Accomplishment {
	accomplishments.Lock()
	defer accomplishments.Unlock()
	_, ok := accomplishments.Accomplishments[name]
	if ok {
		return nil
	}
	newAccomplishment := Accomplishment{Name: name, AcquiredAt: time.Now()}
	accomplishments.Accomplishments[name] = newAccomplishment
	return &newAccomplishment
}
