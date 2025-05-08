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
const (
	becomeMostDangerous = "Become most dangerous"   // X
	scoreAGoal          = "Score a goal"            // X
	winningAGame        = "Score game winning goal" // X
	defeatPlayer        = "Defeat another player"   // X
	tenKillStreak       = "10 Kill streak"          // X
	hundredKillStreak   = "100 Kill streak"         // X
	oneThousandMoney    = "1,000 money"             // X
	fiftyThousandMoney  = "50,000 money"            // X
	doubleKill          = "Double kill"             // X
	tripleKill          = "Triple kill"             // X
	puzzle0             = "Puzzle 0"                // X
)

// Difficulty level ?

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

func (player *Player) addAccomplishmentByName(accomplishmentName string) {
	acc := player.accomplishments.addByName(accomplishmentName)
	if acc == nil {
		return
	}
	logger.Debug().Msg("Adding Accomplishment: " + acc.Name)
	player.world.db.addAccomplishmentToPlayer(player.username, acc.Name, *acc)
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
