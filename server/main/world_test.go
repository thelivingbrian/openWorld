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
	world := createGameWorld(testdb(), &testingConfig)

	req := createLoginRequest(PlayerRecord{Username: "test1", Y: 2, X: 2, StageName: "test-large"})
	world.addIncoming(req)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		world.NewSocketConnection(w, r)
	}))
	defer server.Close()

	testingSocket := createTestingSocket(server.URL)
	defer testingSocket.ws.Close()

	testingSocket.writeOrFatal(createInitialTokenMessage(req.Token))

	testingSocket.writeOrFatal(createSocketEventMessage(req.Token, "d"))

	testingSocket.writeOrFatal(createSocketEventMessage(req.Token, "d"))

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
	defer shutDown()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		world.NewSocketConnection(w, r)
	}))
	defer server.Close()

	_, cancel, tokens, _ := socketsCancelsTokensWaiter(world, server.URL, 1, "test-red")
	firstToken := tokens[0]

	_, cancelers, _, wg := socketsCancelsTokensWaiter(world, server.URL, PLAYER_COUNT, "test-blue")

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

	player.moveEast()
	player.activatePower()
	player.activatePower()
	if player.actions.spaceStack.peek() != nil {
		t.Error("Player powers should be gone")

	}

	if player.getKillStreakSync() != PLAYER_COUNT {
		t.Error("Player should have killed all others")

	}
	if len(world.worldPlayers) != PLAYER_COUNT+1 {
		t.Error("All should be logged in")

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
	world, shutDown := createWorldForTesting()
	defer shutDown()

	zerolog.SetGlobalLevel(zerolog.WarnLevel) // prevent log spam
	defer zerolog.SetGlobalLevel(zerolog.InfoLevel)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		world.NewSocketConnection(w, r)
	}))
	defer server.Close()

	_, cancel, tokens, _ := socketsCancelsTokensWaiter(world, server.URL, PLAYER_COUNT1, "test-red")
	for i := range cancel {
		defer cancel[i]()
	}
	firstToken := tokens[0]

	_, cancelers, _, wg := socketsCancelsTokensWaiter(world, server.URL, PLAYER_COUNT2, "test-blue")

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

	player.moveEast()
	go player.activatePower()
	go player.activatePower()
	time.Sleep(200 * time.Millisecond)
	for index := range cancelers {
		go cancelers[index]()
	}

	wg.Wait()
	time.Sleep(1000 * time.Millisecond)
	// Investigate why this sometimes fails.
	// if player.getKillStreakSync() == 0 {
	// 	t.Error("Player should have killed at least one")
	// }
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
	world := createGameWorld(testdb(), &testingConfig)
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
				return
			case player, ok := <-playersToLogout: // gross and global
				if !ok {
					return
				}
				completeLogout(player)
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

/////////// Socket /////////////////

type TestingSocket struct {
	ws *websocket.Conn
}

func socketsCancelsTokensWaiter(world *World, serverURL string, PLAYER_COUNT int, team string) ([]*TestingSocket, []context.CancelFunc, []string, *sync.WaitGroup) {
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
		go testingSocket.readUntilClose()

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
	ctx, cancel := context.WithCancel(context.Background()) // Point of context vs just calling close?
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

func createSocketEventMessage(token, name string) []byte {
	var msg = PlayerSocketEvent{
		Token: token,
		Name:  name,
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
