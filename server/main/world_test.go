package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/websocket"
)

func TestMostDangerous(t *testing.T) {
	world := createGameWorld(testdb) // yolo?

	p := world.join(&PlayerRecord{Username: "test1", Y: 2, X: 2, StageName: "test-large"})
	fmt.Println(p.id)
	fmt.Println("Player X coordinate: ", p.x)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		world.NewSocketConnection(w, r)
	}))
	defer server.Close()

	testingSocket := createTestingSocket(server.URL)
	defer testingSocket.ws.Close()

	testingSocket.writeOrFatal(createInitialTokenMessage(p.id))

	_ = testingSocket.readOrFatal()

	testingSocket.writeOrFatal(createSocketEventMessage(p.id, "d"))

	_ = testingSocket.readOrFatal()

	fmt.Println("Player tile X coordinate: ", p.x)

	if len(world.worldPlayers) != 1 {
		t.Error("Incorrect number of players")
	}

	if world.leaderBoard.mostDangerous.Peek() != p {
		t.Error("New Player should be most dangerous")
	}
}

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
