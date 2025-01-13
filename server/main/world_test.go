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
)

func TestSocketJoinAndMove(t *testing.T) {
	world := createGameWorld(testdb())

	req := createLoginRequest(PlayerRecord{Username: "test1", Y: 2, X: 2, StageName: "test-large"})
	world.addIncoming(req)

	//initialCoordiate := 2

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		world.NewSocketConnection(w, r)
	}))
	defer server.Close()

	testingSocket := createTestingSocket(server.URL)
	defer testingSocket.ws.Close()

	testingSocket.writeOrFatal(createInitialTokenMessage(req.Token))
	_ = testingSocket.readOrFatal()

	testingSocket.writeOrFatal(createSocketEventMessage(req.Token, "d"))
	_ = testingSocket.readOrFatal()

	testingSocket.writeOrFatal(createSocketEventMessage(req.Token, "d"))
	_ = testingSocket.readOrFatal()

	// Assert
	if len(world.worldPlayers) != 1 {
		t.Error("Incorrect number of players")
	}

	if world.leaderBoard.mostDangerous.Peek().username == req.Record.Username {
		t.Error("New Player should not be most dangerous") // Minimum killstreak player prevents newly joining from being most dangerous
	}

	//fmt.Println(p.x)
	/*
		Cannot be tested via the socket, but can check through player list?
			if initialCoordiate == p.x {
				t.Error("Player has not moved") // This can fail due to race
			}
	*/
}

func createWorldForTesting() (*World, context.CancelFunc) {
	loadFromJson()
	world := createGameWorld(testdb())

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

func TestLogoutAndDeath(t *testing.T) {
	world, shutDown := createWorldForTesting()
	defer shutDown()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		world.NewSocketConnection(w, r)
	}))
	defer server.Close()

	wg := &sync.WaitGroup{}
	cancelers := make([]context.CancelFunc, 0)
	firstToken := ""
	PLAYER_COUNT := 25
	for i := 0; i < PLAYER_COUNT; i++ {
		req := createLoginRequest(PlayerRecord{Username: fmt.Sprintf("TEST%d", i), Team: "test-blue", Y: 2, X: 2, Health: 50, StageName: "test-large"})
		if i == 0 {
			req.Record.Team = "test-red"
			firstToken = req.Token
		}
		world.addIncoming(req)

		testingSocket, cancel := createTestingSocketWithCancel(server.URL, wg)
		if cancel != nil {
			wg.Add(1)
		}
		cancelers = append(cancelers, cancel)
		testingSocket.writeOrFatal(createInitialTokenMessage(req.Token))
	}

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

	if player.getKillStreakSync() != PLAYER_COUNT-1 {
		t.Error("Player should have killed all others")

	}
	if len(world.worldPlayers) != PLAYER_COUNT {
		t.Error("All 100 should be logged in")

	}

	for index := range cancelers {
		// Log everyone out
		cancelers[index]()
	}

	wg.Wait()
	time.Sleep(1000 * time.Millisecond) // Logout completion is not tied to waitGroup

	if len(world.worldPlayers) != 0 {
		t.Error("All players should be logged out")
	}
}

func TestLogoutAndDeath_Concurrent(t *testing.T) {
	world, shutDown := createWorldForTesting()
	defer shutDown()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		world.NewSocketConnection(w, r)
	}))
	defer server.Close()

	wg := &sync.WaitGroup{}
	cancelers := make([]context.CancelFunc, 0)
	firstToken := ""
	PLAYER_COUNT := 25
	for i := 0; i < PLAYER_COUNT; i++ {
		req := createLoginRequest(PlayerRecord{Username: fmt.Sprintf("TEST%d", i), Team: "test-blue", Y: 2, X: 2, Health: 50, StageName: "test-large"})
		if i == 0 {
			req.Record.Team = "test-red"
			firstToken = req.Token
		}
		world.addIncoming(req)

		testingSocket, cancel := createTestingSocketWithCancel(server.URL, wg)
		if cancel != nil && i != 0 {
			wg.Add(1)
		}
		cancelers = append(cancelers, cancel)
		testingSocket.writeOrFatal(createInitialTokenMessage(req.Token))
	}
	// close last player
	defer wg.Add(1)
	defer cancelers[0]()

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
		if index == 0 {
			// Skip primary player or the will not be able to attack
			continue
		}
		go cancelers[index]()
	}

	wg.Wait()
	time.Sleep(1000 * time.Millisecond)
	if player.getKillStreakSync() == 0 {
		t.Error("Player should have killed at least one")

	}
	fmt.Println("players after logout:", len(world.worldPlayers))
	if len(world.worldPlayers) != 1 {
		t.Error("only main player should be logged in")

	}
}

func TestMostDangerous(t *testing.T) {
	loadFromJson()
	world := createGameWorld(testdb())
	stage := getStageFromStageName(world, "test-large")

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
	if world.leaderBoard.mostDangerous.Peek() != p2 {
		t.Error("Invalid leader should be p2")
	}

	p3.incrementKillStreak()
	p3.incrementKillStreak()
	p3.incrementKillStreak()

	p1.incrementKillStreak()
	p1.incrementKillStreak()
	if world.leaderBoard.mostDangerous.Peek() != p3 {
		t.Error("Invalid leader should be p3")
	}

	time.Sleep(25 * time.Millisecond)

	// bypass initiatelogout(p3) which executes non-deterministically
	fullyRemovePlayer(p3)
	completeLogout(p3)

	if world.leaderBoard.mostDangerous.Peek() != p1 {
		t.Error("Invalid leader should be p1")
	}

}

//////////////////////////////////////////////////////////////////////////////////////
// Helpers

type TestingSocket struct {
	ws *websocket.Conn
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
	ctx, cancel := context.WithCancel(context.Background())
	go func(ctx context.Context) {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				ts.ws.Close()
				return
				//default:
				// just usee ts.ws.ReadMessage() ?
				//ts.readOrFatal()
			}
		}
	}(ctx)
	return ts, cancel
}

func createInitialTokenMessage(token string) []byte {
	var msg = struct {
		Token string
	}{
		Token: token, //"TestToke",
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
