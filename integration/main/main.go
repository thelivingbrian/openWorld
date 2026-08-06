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
	maxLoadTestPlayers              = 900
	maxLoadTestTTL                  = 1800
	twoTeams256PlayersLoadTestStyle = "two_teams_256_players"
	twoTeams512PlayersLoadTestStyle = "two_teams_512_players"
)

type loadTestConfig struct {
	Style     string
	StageName string
	Count     int
	TTL       int
	Action    string
	Team      string
	Read      bool
}

type loadTestBatch struct {
	StageName string
	Count     int
	Team      string
}

func main() {
	fmt.Println("Initializing...")
	_ = godotenv.Load()

	config := parseFlags()
	if config.Style != "" || config.StageName != "" {
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
	flag.StringVar(&config.Style, "style", "", "named load-test style")
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

	batches, err := loadTestBatches(config)
	if err != nil {
		return err
	}

	action, err := socketActionForName(config.Action)
	if err != nil {
		return err
	}

	expectedPlayers := 0
	connectedPlayers := 0
	doneChannels := make([]<-chan struct{}, 0, len(batches))
	for _, batch := range batches {
		batchTokens, err := requestTokens(batch.StageName, strconv.Itoa(batch.Count), batch.Team)
		if err != nil {
			return fmt.Errorf("prepare stage %q: %w", batch.StageName, err)
		}
		if len(batchTokens) != batch.Count {
			return fmt.Errorf("stage %q requested %d tokens but received %d", batch.StageName, batch.Count, len(batchTokens))
		}

		connected, done, joinErr := createSocketsAndSendActions(batchTokens, config.Read, config.TTL, action)
		if joinErr != nil {
			fmt.Printf("Stage %q joined %d/%d player(s); first failure: %v\n", batch.StageName, connected, batch.Count, joinErr)
		}
		connectedPlayers += connected
		doneChannels = append(doneChannels, done)
		expectedPlayers += batch.Count
	}

	fmt.Printf("Connected %d/%d simulated player socket(s) across %d stage batch(es); running for %d second(s)\n", connectedPlayers, expectedPlayers, len(batches), config.TTL)
	for _, done := range doneChannels {
		<-done
	}

	if connectedPlayers != expectedPlayers {
		return fmt.Errorf("only %d of %d WebSocket connections were established", connectedPlayers, expectedPlayers)
	}

	fmt.Printf("Load test completed with %d simulated player socket(s)\n", connectedPlayers)
	return nil
}

func validateLoadTestConfig(config loadTestConfig) error {
	if config.Style == "" && config.StageName == "" {
		return errors.New("style or stagename is required")
	}
	if config.Style != "" && config.StageName != "" {
		return errors.New("style and stagename cannot be used together")
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
	batches, err := loadTestBatches(config)
	if err != nil {
		return err
	}
	totalPlayers := 0
	for _, batch := range batches {
		if batch.StageName == "" {
			return errors.New("load-test batches must specify a stagename")
		}
		if batch.Count < 1 {
			return errors.New("load-test batch counts must be positive")
		}
		totalPlayers += batch.Count
	}
	if totalPlayers > maxLoadTestPlayers {
		return fmt.Errorf("load-test styles cannot exceed %d players", maxLoadTestPlayers)
	}
	return nil
}

func loadTestBatches(config loadTestConfig) ([]loadTestBatch, error) {
	if config.Style == "" {
		return []loadTestBatch{{
			StageName: config.StageName,
			Count:     config.Count,
			Team:      config.Team,
		}}, nil
	}

	switch config.Style {
	case twoTeams256PlayersLoadTestStyle:
		return twoTeamsPlayersBatches(8), nil
	case twoTeams512PlayersLoadTestStyle:
		return twoTeamsPlayersBatches(16), nil
	default:
		return nil, fmt.Errorf("unsupported load-test style %q", config.Style)
	}
}

func twoTeamsPlayersBatches(playersPerStage int) []loadTestBatch {
	batches := make([]loadTestBatch, 0, 32)
	stageTeams := []string{"team-blue", "team-fuchsia"}

	for a := 0; a <= 1; a++ {
		for b := 0; b <= 3; b++ {
			for _, stageTeam := range stageTeams {
				batches = append(batches, loadTestBatch{
					StageName: fmt.Sprintf("%s:%d-%d", stageTeam, a, b),
					Count:     playersPerStage,
					Team:      "fuchsia",
				})
			}
		}
	}

	for a := 2; a <= 3; a++ {
		for b := 4; b <= 7; b++ {
			for _, stageTeam := range stageTeams {
				batches = append(batches, loadTestBatch{
					StageName: fmt.Sprintf("%s:%d-%d", stageTeam, a, b),
					Count:     playersPerStage,
					Team:      "sky-blue",
				})
			}
		}
	}

	return batches
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

func createSocketsAndSendActions(tokens []string, read bool, ttl int, action func(*TestingSocket, string)) (int, <-chan struct{}, error) {
	var wg sync.WaitGroup
	connected := 0
	done := make(chan struct{})
	var firstJoinError error

	for _, token := range tokens {
		testingSocket := createTestingSocket(os.Getenv("BLOOP_HOST") + "/screen")
		if testingSocket == nil {
			fmt.Println("failed to create testing socket")
			if firstJoinError == nil {
				firstJoinError = errors.New("failed to create WebSocket")
			}
			continue
		}

		if err := testingSocket.tryWrite(createInitialTokenMessage(token)); err != nil {
			testingSocket.close()
			if firstJoinError == nil {
				firstJoinError = fmt.Errorf("failed to send player token: %w", err)
			}
			continue
		}
		if err := testingSocket.waitForInitialScreen(); err != nil {
			testingSocket.close()
			if firstJoinError == nil {
				firstJoinError = err
			}
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

	return connected, done, firstJoinError
}

func requestTokens(stagename, count, team string) ([]string, error) {
	secret := os.Getenv("AUTO_PLAYER_PASSWORD")
	username := createRandomString()
	payload := tokenRequestPayload(secret, username, stagename, team, count)

	tokenEndpoint := os.Getenv("BLOOP_HOST") + "/insert"
	resp, err := http.Post(tokenEndpoint, "application/x-www-form-urlencoded", bytes.NewBufferString(payload))
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

func tokenRequestPayload(secret, username, stagename, team, count string) string {
	payload := url.Values{
		"secret":    {secret},
		"username":  {username},
		"stagename": {stagename},
		"team":      {team},
		"count":     {count},
	}

	// Keep stage-name colons literal while load tests may target server versions
	// whose legacy form parser does not URL-decode values. ParseForm accepts both.
	return strings.ReplaceAll(payload.Encode(), "%3A", ":")
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

func (ts *TestingSocket) waitForInitialScreen() error {
	if err := ts.ws.SetReadDeadline(time.Now().Add(10 * time.Second)); err != nil {
		return fmt.Errorf("failed to set initial screen deadline: %w", err)
	}
	_, msg, err := ts.ws.ReadMessage()
	_ = ts.ws.SetReadDeadline(time.Time{})
	if err != nil {
		return fmt.Errorf("server closed WebSocket before player joined: %w", err)
	}
	if bytes.Contains(msg, []byte("Unable to join")) {
		if bytes.Contains(msg, []byte("Your team is at capacity")) {
			return errors.New("server rejected player because the team is at capacity")
		}
		return errors.New("server rejected player")
	}
	return nil
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
