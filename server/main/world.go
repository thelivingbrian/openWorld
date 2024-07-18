package main

import (
	"container/heap"
	"fmt"
	"sync"
)

type World struct {
	db           *DB
	worldPlayers map[string]*Player
	wPlayerMutex sync.Mutex
	worldStages  map[string]*Stage
	wStageMutex  sync.Mutex
	leaderBoard  *LeaderBoard
}

type LeaderBoard struct {
	richest       *Player
	wealth        int
	mostDangerous MaxStreakHeap // Full sorted list?
	//streak        int
	oldest *Player
}

func createGameWorld(db *DB) *World {
	lb := &LeaderBoard{mostDangerous: MaxStreakHeap{items: make([]*Player, 0), index: make(map[*Player]int)}}
	return &World{db: db, worldPlayers: make(map[string]*Player), worldStages: make(map[string]*Stage), leaderBoard: lb}
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

func (h *MaxStreakHeap) Peek() interface{} {
	if h.Len() == 0 {
		panic("Heap Underflow")
	}
	return h.items[0]
}

// Update changes the value of an item in the heap and fixes the heap.
func (h *MaxStreakHeap) Update(player *Player, streak int) {
	index := h.index[player]
	h.items[index].setKillStreak(streak)
	heap.Fix(h, index)
}

func (world *World) incrementKillStreak(player *Player) {
	newStreak := player.getKillStreakSync() + 1
	world.leaderBoard.mostDangerous.Update(player, newStreak)

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
