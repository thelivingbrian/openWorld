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
	Name       string    `bson:"name,omitempty"`
	AcquiredAt time.Time `bson:"acquiredAt,omitempty"`
}

// Changing name invalidates previous accomplishment - Add ID?
// difficulty levels ?
const (
	becomeMostDangerous = "Become most dangerous"
	scoreAGoal          = "Score a goal"
	winningAGame        = "Score game winning goal"
	defeatPlayer        = "Defeat another player"
	tenKillStreak       = "10 Kill streak"
	hundredKillStreak   = "100 Kill streak"
	oneThousandMoney    = "1,000 money"
	fiftyThousandMoney  = "50,000 money"
	doubleKill          = "Double kill"
	tripleKill          = "Triple kill"
	puzzle0             = "Puzzle 0"
)

var everyAccomplishment = []string{
	becomeMostDangerous,
	scoreAGoal,
	winningAGame,
	defeatPlayer,
	tenKillStreak,
	hundredKillStreak,
	oneThousandMoney,
	fiftyThousandMoney,
	doubleKill,
	tripleKill,
	puzzle0,
}

func (accomplishments *SyncAccomplishmentList) addByName(name string) *Accomplishment {
	accomplishments.Lock()
	defer accomplishments.Unlock()
	_, ok := accomplishments.Accomplishments[name]
	if ok {
		return nil
	}
	newAccomplishment := Accomplishment{Name: name, AcquiredAt: time.Now().UTC()}
	accomplishments.Accomplishments[name] = newAccomplishment
	return &newAccomplishment
}

///////////////////////////////////////////////////////
// Check / Award

func checkFatalityAccomplishments(player *Player, fatalities int) {
	if fatalities >= 2 {
		player.addAccomplishmentByName(doubleKill)
	}
	if fatalities >= 3 {
		player.addAccomplishmentByName(tripleKill)
	}
}

func checkStreakAccomplishments(player *Player, streak int) {
	if streak >= 10 {
		player.addAccomplishmentByName(tenKillStreak)
	}
	if streak >= 100 {
		player.addAccomplishmentByName(hundredKillStreak)
	}
}

func checkMoneyAccomplishments(player *Player, money int) {
	if money >= 1000 {
		player.addAccomplishmentByName(oneThousandMoney)
	}
	if money >= 50000 {
		player.addAccomplishmentByName(fiftyThousandMoney)
	}
}
