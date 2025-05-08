package main

import (
	"sync"
	"time"
)

var EVERY_HAT_NAME_TO_TRIM map[string]string = map[string]string{
	"score-1-goal":   "black-b med",
	"winning-team":   "black-b thick",
	"most-dangerous": "red-b thick",
	// "made-of-money":   "green-b med",
	// "made-of-money-2": "green-b med",
	"puzzle-solve": "lavender-b thick",
	"contributor":  "gold-b thick",
}

type HatList struct {
	Hats    []Hat `bson:"hats"`
	Current *int  `bson:"current"`
}

type Hat struct {
	Name           string    `bson:"name"`
	ToggleDisabled bool      `bson:"toggleDisabled"`
	UnlockedAt     time.Time `bson:"unlockedAt"`
}

///////////////////////////////////////////////////
//  Player HatList (SyncHatList)

type SyncHatList struct {
	sync.Mutex
	HatList
}

func (hatList *SyncHatList) addByName(hatName string) *Hat {
	hatList.Lock()
	defer hatList.Unlock()
	for i := range hatList.Hats {
		if hatList.Hats[i].Name == hatName {
			hatList.Current = &i
			return nil
		}
	}
	newHat := Hat{Name: hatName, ToggleDisabled: false, UnlockedAt: time.Now()}
	hatList.Hats = append(hatList.Hats, newHat)
	hatCount := len(hatList.Hats) - 1
	hatList.Current = &hatCount
	return &hatList.Hats[hatCount]
}

func (hatList *SyncHatList) peek() *Hat {
	hatList.Lock()
	defer hatList.Unlock()
	if hatList.Current == nil {
		return nil
	}
	i := *hatList.Current
	if i < 0 || i >= len(hatList.Hats) {
		hatList.Current = nil
		return nil
	}
	return &hatList.Hats[*hatList.Current]
}

func (hatList *SyncHatList) next() *Hat {
	hatList.Lock()
	defer hatList.Unlock()
	hatCount := len(hatList.Hats)
	if hatCount == 0 {
		return nil
	}
	if hatList.Current == nil {
		current := 0
		hatList.Current = &current
		return &hatList.Hats[0]
	}
	if *hatList.Current == hatCount-1 {
		hatList.Current = nil
		return nil
	}
	*hatList.Current++
	return &hatList.Hats[*hatList.Current]
}

func (hatList *SyncHatList) nextValid() *Hat {
	for {
		hat := hatList.next()
		if hat == nil {
			return nil
		}
		if !hat.ToggleDisabled {
			return hat
		}
	}
}

func (hatList *SyncHatList) indexSync() *int {
	hatList.Lock()
	defer hatList.Unlock()
	return hatList.Current
}

func (hatList *SyncHatList) currentTrim() string {
	hat := hatList.peek()
	if hat == nil {
		return ""
	}
	hatName := hat.Name
	trim, ok := EVERY_HAT_NAME_TO_TRIM[hatName]
	if !ok {
		logger.Warn().Msg("INVALID HATNAME: " + hatName)
		return ""
	}
	return trim
}
