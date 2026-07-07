package main

import (
	"bytes"
	crand "crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
)

var WAIT_DURATION = 100 * time.Millisecond

const (
	maxLoadTestPlayers = 900
	maxLoadTestTTL     = 900
)

type loadTestConfig struct {
	StageName string
	Count     int
	TTL       int
	Action    string
	Team      string
	Read      bool
}

func main() {
	fmt.Println("Initializing...")
	_ = godotenv.Load()

	config := parseFlags()
	if config.StageName != "" {
		if err := runLoadTest(config); err != nil {
			fmt.Fprintln(os.Stderr, "Load test failed:", err)
			os.Exit(1)
		}
		return
	}

	fmt.Println("Preparing for interactions...")
	http.HandleFunc("/mass", IntegrationClientBed)

	err := http.ListenAndServe(":4440", nil)
	if err != nil {
		fmt.Println("Failed to start server", err)
		return
	}
}

func parseFlags() loadTestConfig {
	config := loadTestConfig{}
	flag.StringVar(&config.StageName, "stagename", "", "stage to populate; when omitted, serve the local /mass endpoint")
	flag.IntVar(&config.Count, "count", 100, "number of simulated players")
	flag.IntVar(&config.TTL, "ttl", 300, "test duration in seconds")
	flag.StringVar(&config.Action, "action", "random", "player action: random, circles, lr, movespace, or space")
	flag.StringVar(&config.Team, "team", "", "team assigned to simulated players")
	flag.BoolVar(&config.Read, "read", true, "read messages received from the server")
	flag.Parse()
	return config
}

func runLoadTest(config loadTestConfig) error {
	if err := validateLoadTestConfig(config); err != nil {
		return err
	}

	action, err := socketActionForName(config.Action)
	if err != nil {
		return err
	}

	tokens, err := requestTokens(config.StageName, strconv.Itoa(config.Count), config.Team)
	if err != nil {
		return err
	}
	if len(tokens) != config.Count {
		return fmt.Errorf("requested %d tokens but received %d", config.Count, len(tokens))
	}

	connected, done := createSocketsAndSendActions(tokens, config.Read, config.TTL, action)
	fmt.Printf("Connected %d/%d simulated player(s); running for %d second(s)\n", connected, len(tokens), config.TTL)
	<-done

	if connected != len(tokens) {
		return fmt.Errorf("only %d of %d WebSocket connections were established", connected, len(tokens))
	}

	fmt.Printf("Load test completed with %d simulated player(s)\n", connected)
	return nil
}

func validateLoadTestConfig(config loadTestConfig) error {
	if config.StageName == "" {
		return errors.New("stagename is required")
	}
	if config.Count < 1 || config.Count > maxLoadTestPlayers {
		return fmt.Errorf("count must be between 1 and %d", maxLoadTestPlayers)
	}
	if config.TTL < 1 || config.TTL > maxLoadTestTTL {
		return fmt.Errorf("ttl must be between 1 and %d seconds", maxLoadTestTTL)
	}
	if strings.TrimSpace(os.Getenv("BLOOP_HOST")) == "" {
		return errors.New("BLOOP_HOST is required")
	}
	if os.Getenv("AUTO_PLAYER_PASSWORD") == "" {
		return errors.New("AUTO_PLAYER_PASSWORD is required")
	}
	return nil
}

func IntegrationClientBed(w http.ResponseWriter, r *http.Request) {
	// curl.exe -X POST "http://localhost:4440/mass?stagename=camera-test&read=true&count=200&ttl=55&action=random&team=fuchsia"
	// curl -X GET "http://localhost:4440/stats"
	stagename := r.URL.Query().Get("stagename")

	var read bool
	readParam := r.URL.Query().Get("read")
	if readParam != "" {
		var err error
		read, err = strconv.ParseBool(readParam)
		if err != nil {
			http.Error(w, "Invalid 'read' parameter", http.StatusBadRequest)
			return
		}
	}

	count := r.URL.Query().Get("count")
	if stagename == "" || count == "" {
		http.Error(w, "Missing required param stagename", http.StatusBadRequest)
		return
	}

	ttlString := r.URL.Query().Get("ttl")
	ttl, err := strconv.Atoi(ttlString)
	if err != nil {
		fmt.Println("Invalid TTL")
		ttl = 30
	}

	team := r.URL.Query().Get("team")

	socketAction, err := socketActionForName(r.URL.Query().Get("action"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	tokens, err := requestTokens(stagename, count, team)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	go func() {
		createSocketsAndSendActions(tokens, read, ttl, socketAction)
	}()

	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "Request successful!")
}

func socketActionForName(name string) (func(*TestingSocket, string), error) {
	switch name {
	case "", "circles":
		return moveInCircles, nil
	case "random":
		return moveRandomly, nil
	case "lr":
		return leftRight, nil
	case "movespace":
		return moveAndSpace, nil
	case "space":
		return spamSpace, nil
	default:
		return nil, fmt.Errorf("unsupported action %q", name)
	}
}

func createSocketsAndSendActions(tokens []string, read bool, ttl int, action func(*TestingSocket, string)) (int, <-chan struct{}) {
	var wg sync.WaitGroup
	connected := 0
	done := make(chan struct{})

	for _, token := range tokens {
		testingSocket := createTestingSocket(os.Getenv("BLOOP_HOST") + "/screen")
		if testingSocket == nil {
			fmt.Println("failed to create testing socket")
			continue
		}

		if err := testingSocket.tryWrite(createInitialTokenMessage(token)); err != nil {
			testingSocket.close()
			continue
		}
		connected++
		wg.Add(1)

		if read {
			go testingSocket.readUntilNil()
		}

		go action(testingSocket, token)
		go func(ts *TestingSocket) {
			defer wg.Done()
			time.Sleep(time.Duration(ttl) * time.Second)
			ts.close()
		}(testingSocket)
	}

	go func() {
		wg.Wait()
		close(done)
	}()

	return connected, done
}

func requestTokens(stagename, count, team string) ([]string, error) {
	secret := os.Getenv("AUTO_PLAYER_PASSWORD")
	username := createRandomString()
	payload := url.Values{
		"secret":    {secret},
		"username":  {username},
		"stagename": {stagename},
		"team":      {team},
		"count":     {count},
	}

	tokenEndpoint := os.Getenv("BLOOP_HOST") + "/insert"
	resp, err := http.Post(tokenEndpoint, "application/x-www-form-urlencoded", bytes.NewBufferString(payload.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch tokens: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read token response: %w", err)
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("token endpoint returned %s", resp.Status)
	}
	var tokens []string
	if err := json.Unmarshal(body, &tokens); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	fmt.Printf("Received %d Token(s)\n", len(tokens))
	return tokens, nil
}

func createRandomString() string {
	bytes := make([]byte, 8)
	_, err := crand.Read(bytes)
	if err != nil {
		panic(err)
	}
	return hex.EncodeToString(bytes)
}

type TestingSocket struct {
	ws      *websocket.Conn
	closing atomic.Bool
}

func createTestingSocket(url string) *TestingSocket {
	if strings.HasPrefix(url, "https://") {
		url = "wss://" + strings.TrimPrefix(url, "https://")
	} else if strings.HasPrefix(url, "http://") {
		url = "ws://" + strings.TrimPrefix(url, "http://")
	}
	ws, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		fmt.Printf("could not dial: %v\n", err)
		return nil
	}
	return &TestingSocket{ws: ws}
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

// import ?
type PlayerSocketEvent struct {
	Token    string `json:"token"`
	Name     string `json:"eventname"`
	MenuName string `json:"menuName"`
	Arg0     string `json:"arg0"`
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

func (ts *TestingSocket) tryWrite(msg []byte) error {
	err := ts.ws.WriteMessage(websocket.TextMessage, msg)
	if err != nil && !ts.closing.Load() {
		fmt.Printf("could not send WebSocket message: %v\n", err)
	}
	return err
}

func (ts *TestingSocket) close() {
	ts.closing.Store(true)
	ts.ws.Close()
}

func (ts *TestingSocket) tryRead() []byte {
	_, msg, err := ts.ws.ReadMessage()
	if err != nil {
		if !ts.closing.Load() {
			fmt.Printf("could not read WebSocket message: %v\n", err)
		}
		return nil
	}
	return msg
}

func (ts *TestingSocket) readUntilNil() {
	for ts.tryRead() != nil {

	}
}

func moveRandomly(ts *TestingSocket, token string) {
	for {
		randn := rand.Intn(5000)
		if randn%4 == 0 {
			if ts.tryWrite(createSocketEventMessage(token, "a")) != nil {
				break
			}
			time.Sleep(WAIT_DURATION)
		}
		if randn%4 == 1 {
			if ts.tryWrite(createSocketEventMessage(token, "w")) != nil {
				break
			}
			time.Sleep(WAIT_DURATION)
		}
		if randn%4 == 2 {
			if ts.tryWrite(createSocketEventMessage(token, "d")) != nil {
				break
			}
			time.Sleep(WAIT_DURATION)
		}
		if randn%4 == 3 {
			if ts.tryWrite(createSocketEventMessage(token, "s")) != nil {
				break
			}
			time.Sleep(WAIT_DURATION)
		}

		if randn%250 == 0 {
			if ts.tryWrite(createSocketEventMessage(token, "Space-On")) != nil {
				break
			}
		}
	}
}

func moveAndSpace(ts *TestingSocket, token string) {
	for {
		randn := rand.Intn(5000)
		if randn%4 == 0 {
			if ts.tryWrite(createSocketEventMessage(token, "a")) != nil {
				break
			}
			time.Sleep(WAIT_DURATION)
		}
		if randn%4 == 1 {
			if ts.tryWrite(createSocketEventMessage(token, "w")) != nil {
				break
			}
			time.Sleep(WAIT_DURATION)
		}
		if randn%4 == 2 {
			if ts.tryWrite(createSocketEventMessage(token, "d")) != nil {
				break
			}
			time.Sleep(WAIT_DURATION)
		}
		if randn%4 == 3 {
			if ts.tryWrite(createSocketEventMessage(token, "s")) != nil {
				break
			}
			time.Sleep(WAIT_DURATION)
		}

		if ts.tryWrite(createSocketEventMessage(token, "Space-On")) != nil {
			break
		}

	}
}

func spamSpace(ts *TestingSocket, token string) {
	for {
		if ts.tryWrite(createSocketEventMessage(token, "Space-On")) != nil {
			break
		}
		time.Sleep(WAIT_DURATION)
	}
}

func moveInCircles(ts *TestingSocket, token string) {
	for {
		if ts.tryWrite(createSocketEventMessage(token, "w")) != nil {
			break
		}
		time.Sleep(WAIT_DURATION)

		if ts.tryWrite(createSocketEventMessage(token, "a")) != nil {
			break
		}
		time.Sleep(WAIT_DURATION)

		if ts.tryWrite(createSocketEventMessage(token, "s")) != nil {
			break
		}
		time.Sleep(WAIT_DURATION)

		if ts.tryWrite(createSocketEventMessage(token, "d")) != nil {
			break
		}
		time.Sleep(WAIT_DURATION)
	}
}

func leftRight(ts *TestingSocket, token string) {
	for {
		if ts.tryWrite(createSocketEventMessage(token, "d")) != nil {
			break
		}
		time.Sleep(WAIT_DURATION)

		if ts.tryWrite(createSocketEventMessage(token, "d")) != nil {
			break
		}
		time.Sleep(WAIT_DURATION)

		if ts.tryWrite(createSocketEventMessage(token, "d")) != nil {
			break
		}
		time.Sleep(WAIT_DURATION)

		if ts.tryWrite(createSocketEventMessage(token, "a")) != nil {
			break
		}
		time.Sleep(WAIT_DURATION)

		if ts.tryWrite(createSocketEventMessage(token, "a")) != nil {
			break
		}
		time.Sleep(WAIT_DURATION)

		if ts.tryWrite(createSocketEventMessage(token, "a")) != nil {
			break
		}
		time.Sleep(WAIT_DURATION)
	}
}
