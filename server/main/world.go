package main

import (
	"container/heap"
	"fmt"
	"sync"

	"github.com/google/uuid"
)

type World struct {
	db           *DB
	worldPlayers map[string]*Player
	wPlayerMutex sync.Mutex
	worldStages  map[string]*Stage
	wStageMutex  sync.Mutex
	leaderBoard  *LeaderBoard
}

func createGameWorld(db *DB) *World {
	lb := &LeaderBoard{mostDangerous: MaxStreakHeap{items: make([]*Player, 0), index: make(map[*Player]int)}}
	return &World{db: db, worldPlayers: make(map[string]*Player), worldStages: make(map[string]*Stage), leaderBoard: lb}
}

func (world *World) join(record *PlayerRecord) *Player {
	token := uuid.New().String()
	fmt.Println("New Player: " + record.Username)
	fmt.Println("Token: " + token)

	if world.isLoggedInAlready(record.Username) {
		fmt.Println("User attempting to log in but is logged in already: " + record.Username)
		return nil
	}

	newPlayer := &Player{
		id:        token,
		username:  record.Username,
		stage:     nil,
		stageName: record.StageName,
		x:         record.X,
		y:         record.Y,
		actions:   createDefaultActions(),
		health:    record.Health,
		money:     record.Money,
		world:     world,
	}

	//New Method
	world.wPlayerMutex.Lock()
	world.worldPlayers[token] = newPlayer
	world.leaderBoard.mostDangerous.Push(newPlayer) // Give own mutex?
	world.wPlayerMutex.Unlock()

	return newPlayer // Player.world is nil at this point at is assigned later when socket is established
}

func (world *World) isLoggedInAlready(username string) bool {
	world.wPlayerMutex.Lock()
	defer world.wPlayerMutex.Unlock()
	for _, player := range world.worldPlayers {
		if player.username == username {
			return true
		}
	}
	return false
}

///////////////////////////////////////////////////////////////
// LeaderBoards

type LeaderBoard struct {
	richest *Player
	//wealth        int
	mostDangerous MaxStreakHeap
	oldest        *Player
}

type MaxStreakHeap struct {
	items []*Player
	index map[*Player]int // Keep track of item indices
}

func (h MaxStreakHeap) Len() int { return len(h.items) }
func (h MaxStreakHeap) Less(i, j int) bool {
	return h.items[i].getKillStreakSync() > h.items[j].getKillStreakSync()
}
func (h MaxStreakHeap) Swap(i, j int) {
	h.items[i], h.items[j] = h.items[j], h.items[i]
	h.index[h.items[i]] = i
	h.index[h.items[j]] = j
}

func (h *MaxStreakHeap) Push(x interface{}) {
	n := len(h.items)
	item := x.(*Player)
	h.items = append(h.items, item)
	h.index[h.items[n]] = n
}

func (h *MaxStreakHeap) Pop() interface{} {
	old := h.items
	n := len(old)
	item := old[n-1]
	h.items = old[0 : n-1]
	delete(h.index, item)
	return item
}

func (h *MaxStreakHeap) Peek() *Player {
	if h.Len() == 0 {
		panic("Heap Underflow")
	}
	return h.items[0]
}

// Update fixes the heap after player has a change in killstreak, notiying any change in most dangerous
func (h *MaxStreakHeap) Update(player *Player) {
	previousMostDangerous := h.Peek()

	index := h.index[player]
	heap.Fix(h, index)

	currentMostDangerous := h.Peek()
	if currentMostDangerous != previousMostDangerous {
		// New method update channge in most dangerous
		for _, p := range player.world.worldPlayers {
			if p == currentMostDangerous {
				p.updateBottomText("You are the most dangerous bloop!")
			} else {
				p.updateBottomText(player.username + " has become the most dangerous bloop...")
			}
		}
	}
}

/*
func (world *World) incrementKillStreak(player *Player) {
	newStreak := player.getKillStreakSync() + 1
	world.leaderBoard.mostDangerous.Update(player)//, newStreak)

	item := world.leaderBoard.mostDangerous.Peek().(*Player)
	if item == player {
		for _, p2 := range world.worldPlayers {
			if player == p2 {
				p2.updateBottomText("You are the most dangerous bloop!")
			} else {
				p2.updateBottomText(player.username + " has become the most dangerous bloop...")
			}
		}

		fmt.Println(player.username + " is the most dangerous!")
	} else {
		fmt.Println(" Actually... " + item.username + " is the most dangerous")
	}
}

func (world *World) zeroKillStreak(player *Player) {
	world.leaderBoard.mostDangerous.Update(player, 0)
}
*/
