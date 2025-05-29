package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
)

var testingConfig = Configuration{
	envName:    "testbed",
	domainName: "fakeDomainTEST",
}

func TestSocketJoinAndMove(t *testing.T) {
	loadFromJson()
	world, _ := createWorldForTesting()
	// unsafe concurrently (send on closed)
	// defer shutDown()

	req := createLoginRequest(PlayerRecord{Username: "test1", Y: 2, X: 2, StageName: "test-large"})
	world.addIncoming(req)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		world.NewSocketConnection(w, r)
	}))
	defer server.Close()

	testingSocket := createTestingSocket(server.URL)
	defer testingSocket.ws.Close()

	testingSocket.writeOrFatal(createInitialTokenMessage(req.Token))

	testingSocket.writeOrFatal(createSocketEventMessage("d"))

	testingSocket.writeOrFatal(createSocketEventMessage("d"))

	_ = testingSocket.readOrFatal() // reading more than once is a lock risk.

	// Assert
	if len(world.worldPlayers) != 1 {
		t.Error("Incorrect number of players")
	}

	if world.leaderBoard.mostDangerous.Peek().username == req.Record.Username {
		t.Error("New Player should not be most dangerous") // Minimum killstreak player prevents newly joining from being most dangerous
	}
}

func TestLogoutAndDeath(t *testing.T) {
	PLAYER_COUNT := 10
	zerolog.SetGlobalLevel(zerolog.WarnLevel) // prevent log spam
	defer zerolog.SetGlobalLevel(zerolog.InfoLevel)

	world, shutDown := createWorldForTesting()
	defer shutDown() // Only safe in a single threaded context?

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		world.NewSocketConnection(w, r)
	}))
	defer server.Close()

	_, cancel, tokens, _ := socketsCancelsTokensWaiter(world, server.URL, 1, "test-red", true)
	firstToken := tokens[0]

	_, cancelers, _, wg := socketsCancelsTokensWaiter(world, server.URL, PLAYER_COUNT, "test-blue", true)

	time.Sleep(500 * time.Millisecond) // give time for all to join

	// Assert
	player, ok := world.worldPlayers[firstToken]
	if !ok || player == nil {
		t.Error("Player is nil")
	}
	powerUp1 := &PowerUp{areaOfInfluence: grid9x9}
	powerUp2 := &PowerUp{areaOfInfluence: grid5x5}
	player.actions.spaceStack.push(powerUp1)
	player.actions.spaceStack.push(powerUp2)

	if player.getKillStreakSync() != 0 {
		t.Error("Player kill streak should be 0")
	}

	if player.actions.spaceStack.peek() == nil {
		t.Error("Player Should have power ")
	}

	playerCount := len(world.worldPlayers)
	if playerCount != PLAYER_COUNT+1 {
		fmt.Println("Expected players: ", PLAYER_COUNT+1, " - Actual: ", playerCount)
		t.Error("All should be logged in - pre-activation")
	}

	moveEast(player)
	player.activatePower()
	player.activatePower()
	if player.actions.spaceStack.peek() != nil {
		t.Error("Player powers should be gone")

	}

	ks := player.getKillStreakSync()
	if ks != PLAYER_COUNT {
		fmt.Println("Expected KS: ", PLAYER_COUNT, " - Actual: ", ks)
		t.Error("Player should have killed all others")
	}

	playerCount = len(world.worldPlayers)
	if playerCount != PLAYER_COUNT+1 {
		fmt.Println("Expected players: ", PLAYER_COUNT+1, " - Actual: ", playerCount)
		t.Error("All should be logged in - post-activation")
	}

	for index := range cancelers {
		// Log everyone out
		cancelers[index]()
	}
	cancel[0]()

	wg.Wait()
	time.Sleep(1000 * time.Millisecond) // Logout completion is not tied to waitGroup

	if len(world.worldPlayers) != 0 {
		t.Error("All players should be logged out")
	}
}

func TestLogoutAndDeath_Concurrent(t *testing.T) {
	PLAYER_COUNT1 := 15
	PLAYER_COUNT2 := 15
	world, _ := createWorldForTesting()
	// unsafe concurrently (send on closed)
	// defer shutDown()

	zerolog.SetGlobalLevel(zerolog.WarnLevel) // prevent log spam
	defer zerolog.SetGlobalLevel(zerolog.InfoLevel)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		world.NewSocketConnection(w, r)
	}))
	defer server.Close()

	_, cancel, tokens, _ := socketsCancelsTokensWaiter(world, server.URL, PLAYER_COUNT1, "test-red", true)
	for i := range cancel {
		defer cancel[i]()
	}
	firstToken := tokens[0]

	_, cancelers, _, wg := socketsCancelsTokensWaiter(world, server.URL, PLAYER_COUNT2, "test-blue", true)

	time.Sleep(500 * time.Millisecond) // give time for all to join

	// Assert
	player, ok := world.worldPlayers[firstToken]
	if !ok || player == nil {
		t.Error("Player is nil")
	}
	powerUp1 := &PowerUp{areaOfInfluence: grid9x9}
	powerUp2 := &PowerUp{areaOfInfluence: grid5x5}
	player.actions.spaceStack.push(powerUp1)
	player.actions.spaceStack.push(powerUp2)

	if player.getKillStreakSync() != 0 {
		t.Error("Player kill streak should be 0")
	}

	if player.actions.spaceStack.peek() == nil {
		t.Error("Player Should have power ")
	}

	if len(world.worldPlayers) == 0 {
		t.Error("Should have players")
	}

	moveEast(player)
	go player.activatePower()
	go player.activatePower()

	time.Sleep(500 * time.Millisecond)
	for index := range cancelers {
		// Test has deadlocked with: WARN : FAILED TO REMOVE PLAYER
		go cancelers[index]()
	}

	wg.Wait()
	time.Sleep(1000 * time.Millisecond)
	if player.getKillStreakSync() == 0 {
		t.Error("Player should have killed at least one")
	}
	fmt.Println("players after logout:", len(world.worldPlayers))
	if len(world.worldPlayers) != PLAYER_COUNT1 {
		t.Error("Players from first group should be logged in")
	}
	if len(player.getTileSync().stage.playerMap) != PLAYER_COUNT1 {
		fmt.Println("Have this many players on stage: ", len(player.getTileSync().stage.playerMap))
		t.Error("Players from first group should be logged in")
	}
}

func TestMostDangerous(t *testing.T) {
	loadFromJson()
	world, shutDown := createWorldForTesting()
	defer shutDown()
	stage := loadStageByName(world, "test-large")

	req := createLoginRequest(PlayerRecord{Username: "test1", Y: 2, X: 2, StageName: stage.name})
	world.addIncoming(req)
	p1 := world.join(req, &MockConn{})

	req2 := createLoginRequest(PlayerRecord{Username: "test2", Y: 3, X: 3, StageName: stage.name})
	world.addIncoming(req2)
	p2 := world.join(req2, &MockConn{})

	req3 := createLoginRequest(PlayerRecord{Username: "test3", Y: 3, X: 3, StageName: stage.name})
	world.addIncoming(req3)
	p3 := world.join(req3, &MockConn{})

	// Assert
	p2.incrementKillStreak()

	p3.incrementKillStreak()
	p3.incrementKillStreak()
	p3.incrementKillStreak()

	p1.incrementKillStreak()
	p1.incrementKillStreak()
	if world.leaderBoard.mostDangerous.Peek().id != p3.id {
		t.Error("Invalid leader should be p3")
	}

	time.Sleep(25 * time.Millisecond)

	// bypass initiatelogout(p3) which executes non-deterministically
	removeFromTileAndStage(p3)
	completeLogout(p3)

	// Force synchronization
	world.leaderBoard.mostDangerous.incoming <- PlayerStreakRecord{}
	if world.leaderBoard.mostDangerous.Peek().id != p1.id {
		t.Error("Invalid leader should be p1")
	}

	// Cleanup / Prevent send on closed
	p1.tangible, p2.tangible, p3.tangible = false, false, false
}

func TestNewPlayerFromRecord(t *testing.T) {
	world, shutdown := createWorldForTesting()
	defer shutdown()

	tests := []struct {
		name          string
		rec           PlayerRecord
		wantTeam      string
		wantAccompLen int
	}{
		{
			name:          "explicit team retained",
			rec:           createPlayerRecordForTesting("Alice", "red"),
			wantTeam:      "red",
			wantAccompLen: 1,
		},
		{
			name:          "blank team gets default",
			rec:           createPlayerRecordForTesting("Bob", ""),
			wantTeam:      "sky-blue",
			wantAccompLen: 1,
		},
	}

	for _, tc := range tests {
		tc := tc // capture
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			player := world.newPlayerFromRecord(tc.rec, "id‑123")

			if player == nil {
				t.Fatalf("expected player, got nil")
			}
			if got := player.id; got != "id‑123" {
				t.Errorf("id: got %q want %q", got, "id‑123")
			}
			if got := player.username; got != tc.rec.Username {
				t.Errorf("username: got %q want %q", got, tc.rec.Username)
			}
			if got := player.team; got != tc.wantTeam {
				t.Errorf("team: got %q want %q", got, tc.wantTeam)
			}
			if got := player.health.Load(); got != tc.rec.Health {
				t.Errorf("health: got %d want %d", got, tc.rec.Health)
			}
			if got := player.money.Load(); got != tc.rec.Money {
				t.Errorf("money: got %d want %d", got, tc.rec.Money)
			}
			if got := len(player.accomplishments.Accomplishments); got != tc.wantAccompLen {
				t.Errorf("accomplishments length: got %d want %d", got, tc.wantAccompLen)
			}

			// Ensure Stats are assigned
			statCases := []struct {
				name string
				got  int64
				want int64
			}{
				{"Check for killCount", player.PlayerStats.killCount.Load(), tc.rec.Stats.KillCount},
				{"Check for killCountNpc", player.PlayerStats.killCountNpc.Load(), tc.rec.Stats.KillCountNpc},
				{"Check for deathCount", player.PlayerStats.deathCount.Load(), tc.rec.Stats.DeathCount},
				{"Check for goalsScored", player.PlayerStats.goalsScored.Load(), tc.rec.Stats.GoalsScored},
				{"Check for peakKillStreak", player.PlayerStats.peakKillStreak.Load(), tc.rec.Stats.PeakKillStreak},
				{"Check for peakWealth", player.PlayerStats.peakWealth.Load(), tc.rec.Stats.PeakWealth},
			}
			for _, sc := range statCases {
				if sc.got != sc.want {
					t.Errorf("%s: got %d want %d", sc.name, sc.got, sc.want)
				}
			}

			// channel created?
			select {
			case <-player.updates:
				t.Error("updates channel should be empty immediately after creation")
			default:
			}
		})
	}
}

//////////////////////////////////////////////////////////////////////////////////////
// Helpers

//////////// World /////////////

func createWorldForTesting() (*World, context.CancelFunc) {
	loadFromJson()
	world := createGameWorld(testdb(), &testingConfig)

	ctx, cancel := context.WithCancel(context.Background())
	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				close(world.playersToLogout)
				close(world.leaderBoard.mostDangerous.incoming)
				return
			}
		}
	}(ctx)
	return world, cancel
}

func loadStageByName(world *World, name string) *Stage {
	area, success := areaFromName(name)
	if !success {
		return nil
	}
	stage := createStageFromArea(area)
	if area.LoadStrategy == "Individual" {
		return stage
	}
	if stage != nil {
		world.wStageMutex.Lock()
		world.worldStages[name] = stage
		world.wStageMutex.Unlock()
	}
	return stage
}

//////////// Record /////////////

func createPlayerRecordForTesting(username, team string) PlayerRecord {
	now := time.Unix(1_700_000_000, 0) // deterministic
	return PlayerRecord{
		Username: username,
		Team:     team,
		Health:   88,
		Money:    42_000,
		//HatList:  HatList{},
		Accomplishments: map[string]Accomplishment{
			"fresh‑spawn": {Name: "fresh‑spawn", AcquiredAt: now},
		},
		Stats: PlayerStatsRecord{
			KillCount:      50,
			KillCountNpc:   51,
			PeakKillStreak: 10,
			DeathCount:     5,
			GoalsScored:    2,
			PeakWealth:     50_000,
		},
	}
}

/////////// Socket /////////////////

type TestingSocket struct {
	ws *websocket.Conn
}

func socketsCancelsTokensWaiter(world *World, serverURL string, PLAYER_COUNT int, team string, read bool) ([]*TestingSocket, []context.CancelFunc, []string, *sync.WaitGroup) {
	sockets := make([]*TestingSocket, 0)
	cancelers := make([]context.CancelFunc, 0)
	tokens := make([]string, 0)
	wg := &sync.WaitGroup{}
	random := createRandomToken()
	s := random[0 : len(random)/4]
	for i := 0; i < PLAYER_COUNT; i++ {
		req := createLoginRequest(PlayerRecord{Username: fmt.Sprintf("TEST-%s-%d", s, i), Team: team, Y: 2, X: 2, Health: 50, StageName: "test-large"})
		tokens = append(tokens, req.Token)
		world.addIncoming(req)

		testingSocket, cancel := createTestingSocketWithCancel(serverURL, wg)
		if testingSocket == nil || cancel == nil {
			panic("Failed to create cancel or socket")
		}
		wg.Add(1)
		cancelers = append(cancelers, cancel)
		testingSocket.writeOrFatal(createInitialTokenMessage(req.Token))
		if read {
			go testingSocket.readUntilClose()
		}

		sockets = append(sockets, testingSocket)
	}
	return sockets, cancelers, tokens, wg
}

func createTestingSocket(url string) *TestingSocket {
	if url[0:4] == "http" {
		url = "ws" + url[len("http"):]

	}
	ws, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		panic(fmt.Sprintf("could not dial: %v", err))
	}
	return &TestingSocket{ws: ws}
}

func createTestingSocketWithCancel(url string, wg *sync.WaitGroup) (*TestingSocket, context.CancelFunc) {
	ts := createTestingSocket(url)
	ctx, cancel := context.WithCancel(context.Background()) // Point of context vs just calling close? WG?
	go func(ctx context.Context) {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				ts.ws.Close()
				return
			}
		}
	}(ctx)
	return ts, cancel
}

func createInitialTokenMessage(token string) []byte {
	var msg = struct {
		Token string
	}{
		Token: token,
	}
	initialTokenMessage, err := json.Marshal(msg)
	if err != nil {
		panic(fmt.Sprintf("could not marshal: %v", err))
	}
	return initialTokenMessage
}

func createSocketEventMessage(name string) []byte {
	var msg = PlayerSocketEvent{
		Name: name,
	}
	socketMsg, err := json.Marshal(msg)
	if err != nil {
		panic(fmt.Sprintf("could not marshal: %v", err))
	}
	return socketMsg
}

func (ts *TestingSocket) writeOrFatal(msg []byte) {
	err := ts.ws.WriteMessage(websocket.TextMessage, msg)
	if err != nil {
		panic(fmt.Sprintf("could not send message: %s, Error: %v", string(msg), err))
	}
}

func (ts *TestingSocket) readOrFatal() []byte {
	_, msg, err := ts.ws.ReadMessage()
	if err != nil {
		panic(fmt.Sprintf("could not read message - Error: %v", err))
	}
	return msg
}

func (ts *TestingSocket) readUntilClose() {
	for {
		_, _, err := ts.ws.ReadMessage()
		if err != nil {
			return
		}
	}
}
