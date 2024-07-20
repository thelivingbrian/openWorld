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
	world := createGameWorld(nil) // yolo?
	//world.db.getUserByEmail = func(email string) (*User, error) {}
	p := world.join(&PlayerRecord{Username: "test1", Y: 2, X: 2, StageName: "test-large"})
	fmt.Println(p.id)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		world.NewSocketConnection(w, r)
	}))
	defer server.Close()

	url := "ws" + server.URL[len("http"):]
	ws, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("could not dial: %v", err)
	}
	defer ws.Close()

	var msg = struct {
		Token string
	}{
		Token: p.id, //"TestToke",
	}
	initialTokenMessage, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("could not marshal: %v", err)
	}
	err = ws.WriteMessage(websocket.TextMessage, initialTokenMessage)
	if err != nil {
		t.Fatalf("could not send message: %v", err)
	}

	_, _, err = ws.ReadMessage()
	if err != nil {
		t.Fatalf("could not read message: %v", err)
	}
	fmt.Println("Player init tile X coordinate: ", p.tile.x)

	var msg2 = PlayerSocketEvent{
		Token: p.id,
		Name:  "d",
	}
	sendKeyMessage, err := json.Marshal(msg2)
	if err != nil {
		t.Fatalf("could not marshal: %v", err)
	}
	err = ws.WriteMessage(websocket.TextMessage, sendKeyMessage)
	if err != nil {
		t.Fatalf("could not send message: %v", err)
	}

	fmt.Println("Player X coordinate: ", p.x)
	fmt.Println("Player tile X coordinate: ", p.tile.x)

	_, resp, err := ws.ReadMessage()
	if err != nil {
		t.Fatalf("could not read message: %v", err)
	}
	fmt.Println("response ", string(resp))
	fmt.Println("Player tile X coordinate: ", p.tile.x)
	wp := world.worldPlayers[p.id]
	fmt.Println("World Player Tile X coordinate: ", wp.tile.x)
	fmt.Println("World Player X coordinate: ", wp.x)

	if len(world.worldPlayers) != 1 {
		t.Error("Incorrect number of players")
	}
	if wp == p {
		fmt.Println("duh.")
	}
	if world.leaderBoard.mostDangerous.Peek() != p {
		t.Error("New Player should be most dangerous")
	}
}
