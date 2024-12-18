package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/websocket"
)

func TestSocketJoinAndMove(t *testing.T) {
	world := createGameWorld(testdb())

	req := createLoginRequest(PlayerRecord{Username: "test1", Y: 2, X: 2, StageName: "test-large"})
	world.addIncoming(req)
	//p := world.join(req)
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

func TestMostDangerous(t *testing.T) {
	world := createGameWorld(testdb())
	stage := getStageFromStageName(world, "test-large")

	req := createLoginRequest(PlayerRecord{Username: "test1", Y: 2, X: 2, StageName: stage.name})
	world.addIncoming(req)
	p1 := world.join(req)
	go p1.sendUpdates()
	p1.placeOnStage(stage)

	req2 := createLoginRequest(PlayerRecord{Username: "test2", Y: 3, X: 3, StageName: stage.name})
	world.addIncoming(req2)
	p2 := world.join(req2)
	go p2.sendUpdates()
	p2.placeOnStage(stage)

	req3 := createLoginRequest(PlayerRecord{Username: "test3", Y: 3, X: 3, StageName: stage.name})
	world.addIncoming(req3)
	p3 := world.join(req3)
	go p3.sendUpdates()
	p3.placeOnStage(stage)

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

	logOut(p3)
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
