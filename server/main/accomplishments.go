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
	// Event id? - which currently only is/could be mongo _id
}

// Changing name invalidates previous accomplishment - Add ID?
const (
	becomeMostDangerous = "Become most dangerous"
	scoreAGoal          = "Score a goal"
	winningTeam         = "Be on winning team"
	defeatPlayer        = "Defeat another player"
	tenKillStreak       = "10 Kill streak"
	hundredKillStreak   = "100 Kill streak"
	oneThousandMoney    = "1,000 money"
	fiftyThousandMoney  = "50,000 money"
	doubleKill          = "Double kill"
	tripleKill          = "Triple kill"
	puzzle0             = "Puzzle 0"
)

// Difficulty level ? e.g. Win a game is harder than winning team

var everyAccomplishment = []string{
	becomeMostDangerous,
	scoreAGoal,
	winningTeam,
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
