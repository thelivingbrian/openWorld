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
