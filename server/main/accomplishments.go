package main

import (
	"sync"
	"time"
)

type SyncAccomplishmentList struct {
	sync.Mutex
	AccomplishmentList
}

type AccomplishmentList struct {
	Accomplishments map[string]Accomplishment
}

type Accomplishment struct {
	acquiredAt time.Time
	name       string
	// Event id? - which currently only is/could be mongo _id
}
