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
		Token: "TestToke",
	}
	bytes, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("could not marshal: %v", err)
	}
	err = ws.WriteMessage(websocket.TextMessage, bytes)
	if err != nil {
		t.Fatalf("could not send message: %v", err)
	}

	if len(world.worldPlayers) != 1 {
		t.Error("Incorrect number of players")
	}
}
