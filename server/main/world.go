package main

import (
	"container/heap"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

type World struct {
	db                  *DB
	worldPlayers        map[string]*Player
	wPlayerMutex        sync.Mutex
	incomingPlayers     map[string]LoginRequest
	incomingPlayerMutex sync.Mutex
	worldStages         map[string]*Stage
	wStageMutex         sync.Mutex
	leaderBoard         *LeaderBoard
}

func createGameWorld(db *DB) *World {
	minimumKillstreak := Player{id: "HS-only", killstreak: 0} // Do somewhere else?
	lb := &LeaderBoard{mostDangerous: MaxStreakHeap{items: []*Player{&minimumKillstreak}, index: make(map[*Player]int)}}
	return &World{db: db, worldPlayers: make(map[string]*Player), incomingPlayers: make(map[string]LoginRequest), worldStages: make(map[string]*Stage), leaderBoard: lb}
}

type LoginRequest struct {
	Token     string
	Record    PlayerRecord
	timestamp time.Time
}

func createLoginRequest(record PlayerRecord) LoginRequest {
	return LoginRequest{
		Token:     createRandomToken(),
		Record:    record,
		timestamp: time.Now(),
	}
}

func isLessThan15SecondsAgo(t time.Time) bool {
	if time.Since(t) < 0 {
		// t is in the future
		return false
	}
	return time.Since(t) < 15*time.Second
}

func (world *World) addIncoming(loginRequest LoginRequest) {
	world.incomingPlayerMutex.Lock()
	defer world.incomingPlayerMutex.Unlock()
	world.incomingPlayers[loginRequest.Token] = loginRequest
}

func (world *World) retreiveIncoming(token string) *LoginRequest {
	world.incomingPlayerMutex.Lock()
	defer world.incomingPlayerMutex.Unlock()
	request, ok := world.incomingPlayers[token]
	if ok {
		delete(world.incomingPlayers, token)
		if isLessThan15SecondsAgo(request.timestamp) {
			return &request
		}
	}
	return nil
}

func createRandomToken() string {
	token := make([]byte, 16)
	_, err := rand.Read(token)
	if err != nil {
		panic(err)
	}
	return hex.EncodeToString(token)
}

func (world *World) join(incoming LoginRequest) *Player {
	//token := uuid.New().String()
	// need log levels
	//fmt.Println("New Player: " + record.Username)
	//fmt.Println("Token: " + token)

	if world.isLoggedInAlready(incoming.Record.Username) {
		fmt.Println("User attempting to log in but is logged in already: " + incoming.Record.Username)
		return nil
	}

	newPlayer := world.newPlayerFromRecord(incoming.Record, incoming.Token)

	//New Method
	world.wPlayerMutex.Lock()
	world.worldPlayers[incoming.Token] = newPlayer  // join that never upgrades is a leak, could be exploited into save point
	world.leaderBoard.mostDangerous.Push(newPlayer) // Has own mutex. pMutex needed?
	world.wPlayerMutex.Unlock()

	return newPlayer
}

func (world *World) newPlayerFromRecord(record PlayerRecord, id string) *Player {
	// probably take this out later...
	if record.Team == "" {
		record.Team = "sky-blue"
	}
	updatesForPlayer := make(chan []byte)
	newPlayer := &Player{
		id:        id,
		username:  record.Username,
		team:      record.Team,
		trim:      record.Trim,
		stage:     nil,
		updates:   updatesForPlayer,
		stageName: record.StageName,
		x:         record.X,
		y:         record.Y,
		actions:   createDefaultActions(),
		health:    record.Health,
		money:     record.Money,
		world:     world,
		menues:    map[string]Menu{"pause": pauseMenu, "map": mapMenu, "stats": statsMenu, "respawn": respawnMenu}, // terrifying
	}

	newPlayer.setIcon()
	return newPlayer
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
// References / Lookup

func (w *World) getRelativeTile(tile *Tile, yOff, xOff int) *Tile {
	destY := tile.y + yOff
	destX := tile.x + xOff
	if validCoordinate(destY, destX, tile.stage.tiles) {
		return tile.stage.tiles[destY][destX]
	} else {
		escapesVertically, escapesHorizontally := validityByAxis(destY, destX, tile.stage.tiles)
		if escapesVertically && escapesHorizontally {
			// in bloop world cardinal direction travel may be non-communative
			// therefore north-east etc neighbor is not uniquely defined
			// order can probably be uniquely determined when tile.y != tile.x
			return nil
		}
		if escapesVertically {
			var newStage *Stage
			if yOff > 0 {
				newStage = w.getStageByName(tile.stage.south)
				if newStage == nil {
					newStage = w.loadStageByName(tile.stage.south)
				}
			}
			if yOff < 0 {
				newStage = w.getStageByName(tile.stage.north)
				if newStage == nil {
					newStage = w.loadStageByName(tile.stage.north)
				}
			}

			if newStage != nil {
				if validCoordinate(mod(destY, len(newStage.tiles)), destX, newStage.tiles) {
					return newStage.tiles[mod(destY, len(newStage.tiles))][destX]
				}
			}
			return nil
		}
		if escapesHorizontally {
			var newStage *Stage
			if xOff > 0 {
				newStage = w.getStageByName(tile.stage.east)
				if newStage == nil {
					newStage = w.loadStageByName(tile.stage.east)
				}
			}
			if xOff < 0 {
				newStage = w.getStageByName(tile.stage.west)
				if newStage == nil {
					newStage = w.loadStageByName(tile.stage.west)
				}
			}

			if newStage != nil {
				if validCoordinate(destY, mod(destX, len(newStage.tiles)), newStage.tiles) {
					return newStage.tiles[destY][mod(destX, len(newStage.tiles))]
				}
			}
			return nil
		}

		return nil
	}
}

///////////////////////////////////////////////////////////////
// LeaderBoards

type LeaderBoard struct {
	//richest *Player
	mostDangerous MaxStreakHeap
	//oldest        *Player
	scoreboard Scoreboard
}

// Team Scoreboards
type TeamScore struct {
	sync.Mutex
	score int
}

type Scoreboard struct {
	data sync.Map
}

// Increment updates the score for a team atomically
func (s *Scoreboard) Increment(team string) {
	val, _ := s.data.LoadOrStore(team, &TeamScore{})
	teamScore := val.(*TeamScore)

	teamScore.Lock()
	teamScore.score += 1
	teamScore.Unlock()
}

func (s *Scoreboard) GetScore(team string) int {
	val, ok := s.data.Load(team)
	if !ok {
		return 0 // Team does not exist
	}

	teamScore := val.(*TeamScore)

	teamScore.Lock()
	defer teamScore.Unlock()
	return teamScore.score
}

//

type MaxStreakHeap struct {
	items []*Player
	index map[*Player]int // Keep track of item indices
	sync.Mutex
}

func (h *MaxStreakHeap) Len() int {
	h.Lock()
	defer h.Unlock()
	return len(h.items)
}

func (h *MaxStreakHeap) Less(i, j int) bool {
	h.Lock()
	defer h.Unlock()
	return h.items[i].getKillStreakSync() > h.items[j].getKillStreakSync()
}

func (h *MaxStreakHeap) Swap(i, j int) {
	h.Lock()
	h.items[i], h.items[j] = h.items[j], h.items[i]
	h.index[h.items[i]], h.index[h.items[j]] = i, j
	h.Unlock()
}

func (h *MaxStreakHeap) Push(x interface{}) {
	h.Lock()
	n := len(h.items)
	item := x.(*Player)
	h.items = append(h.items, item)
	h.index[h.items[n]] = n // would need fix if not at bottom. (e.g. richest)
	h.Unlock()
}

func (h *MaxStreakHeap) Pop() interface{} {
	h.Lock()
	defer h.Unlock()
	old := h.items
	n := len(old)
	item := old[n-1]
	h.items = old[0 : n-1]
	delete(h.index, item)
	return item
}

func (h *MaxStreakHeap) Peek() *Player {
	h.Lock()
	defer h.Unlock()
	if len(h.items) == 0 {
		return nil
		//panic("Heap Underflow")
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
		notifyChangeInMostDangerous(currentMostDangerous)
	}
}

func notifyChangeInMostDangerous(currentMostDangerous *Player) {
	if currentMostDangerous.id == "HS-only" {
		return
	}
	for _, p := range currentMostDangerous.world.worldPlayers {
		if p == currentMostDangerous {
			p.updateBottomText("You are the most dangerous bloop!")
		} else {
			p.updateBottomText(currentMostDangerous.username + " has become the most dangerous bloop!")
		}
	}
}
