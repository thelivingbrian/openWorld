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
	AcquiredAt time.Time // need bson annotation
	Name       string
	// Event id? - which currently only is/could be mongo _id
}

const (
	becomeMostDangerous = "Become most dangerous"
	scoreAGoal          = "Score a goal"
	defeatPlayer        = "Defeat another player"
	tenKillStreak       = "10 Kill streak"
	hundredKillStreak   = "100 Kill streak"
	oneThousandMoney    = "1,000 money"
	fiftyThousandMoney  = "50,000 money"
	doubleNPCKill       = "Double npc kill - simultaneous"
	tripleNPCKill       = "Triple npc kill - simultaneous"
	puzzle0             = "Puzzle 0"
)

var everyAccomplishment = []string{
	becomeMostDangerous,
	scoreAGoal,
	defeatPlayer,
	tenKillStreak,
	hundredKillStreak,
	oneThousandMoney,
	fiftyThousandMoney,
	doubleNPCKill,
	tripleNPCKill,
	puzzle0,
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
