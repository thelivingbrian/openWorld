package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	mrand "math/rand"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

const HIGHSCORE_CHECK_INTERVAL_IN_SECONDS = 15

// Name arguably makes sense a level up - e.g. the hub is the central id authoritity that sends users to satalite bloop worlds
//
//	but this is not that thing. This is just a subset of that specific to highscores.
//	new name? RankingManager, Rankings?
type Hub struct {
	richest, deadliest, mvp *HighScoreListSync
	db                      RankingProvider
}

type HighScoreListSync struct {
	sync.Mutex
	lastChecked time.Time
	HighScoreList
}

type HighScoreList struct {
	BorderColor string // unused
	Category    string
	Entries     []HighScoreEntry
}

type HighScoreEntry struct {
	Username   string
	StatNames  []string
	StatValues []string
}

type RankingProvider interface {
	getTopNPlayersByField(field string, n int) ([]PlayerRecord, error)
}

// Used w/ Template methods to cycle through on site
var queryCategories = []string{
	"Richest",
	"Deadliest",
	"MVP",
}

func createDefaultHub(db RankingProvider) *Hub {
	return &Hub{
		db: db,
		richest: &HighScoreListSync{
			HighScoreList: HighScoreList{
				Category:    "Richest",
				BorderColor: "green",
			},
		},
		deadliest: &HighScoreListSync{
			HighScoreList: HighScoreList{
				Category:    "Deadliest",
				BorderColor: "red",
			},
		},
		mvp: &HighScoreListSync{
			HighScoreList: HighScoreList{
				Category:    "MVP",
				BorderColor: "gold",
			},
		},
	}
}

//////////////////////////////////////////////////////////
// Site Handlers

func (app *App) homeHandler(w http.ResponseWriter, r *http.Request) {
	logger.Info().Msg("Home page accessed.") // Replace with metric
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	_, identifierFound := getUserIdFromSession(r)
	if identifierFound {
		tmpl.ExecuteTemplate(w, "homepage-signed-in", nil)
		return
	}
	tmpl.ExecuteTemplate(w, "homepage", app.config.guestsEnabled.Load())
}

func aboutHandler(w http.ResponseWriter, r *http.Request) {
	logger.Info().Msg("About page accessed.") // Metric > log
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	tmpl.ExecuteTemplate(w, "about", nil)
}

func (hub *Hub) highscoreHandler(w http.ResponseWriter, r *http.Request) {
	logger.Info().Msg("Highscore page accessed.")

	queryValues := r.URL.Query()
	category := strings.ToLower(queryValues.Get("category"))

	var scores HighScoreList
	switch category {
	case "richest":
		scores = generateRichestList(hub)
	case "deadliest":
		scores = generateDeadliestList(hub)
	case "mvp":
		scores = generateMVPList(hub)
	default:
		break
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	tmpl.ExecuteTemplate(w, "highscore", scores)
}

//////////////////////////////////////////////////////////////////
// Guests

func (app *App) guestsHandler(w http.ResponseWriter, r *http.Request) {
	if !app.config.guestsEnabled.Load() {
		tmpl.ExecuteTemplate(w, "homepage", app.config.guestsEnabled.Load())
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	switch r.Method {
	case "GET":
		tmpl.ExecuteTemplate(w, "guests", nil)
	case "POST":
		app.storeNewGuestSession(w, r)
		w.Header().Set("HX-Redirect", "/")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Request was successful - Redirecting"))
	}
}

func (app *App) storeNewGuestSession(w http.ResponseWriter, r *http.Request) {
	hexid, err := randomHex16()
	if err != nil {
		return
	}
	identifier := "guest:" + hexid
	username := "#" + hexid
	team := "sky-blue"
	if mrand.Intn(2) == 1 {
		team = "fuchsia"
	}

	// Capcha check
	// Rate limiter

	record := createNewGuestPlayerRecord(username, team)
	err = app.db.InsertPlayerRecord(record)
	if err != nil {
		io.WriteString(w, divBottomInvalid("Error saving new player"))
		return
	}

	session, err := store.Get(r, "user-session")
	if err != nil {
		logger.Warn().Msg("Error getting new session?")
	}
	session.Values["identifier"] = identifier
	err = session.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func randomHex16() (string, error) {
	b := make([]byte, 8) // 8 bytes × 2 hex chars/byte = 16 hex chars
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to read random bytes: %w", err)
	}
	return hex.EncodeToString(b), nil
}

func createNewGuestPlayerRecord(username, team string) PlayerRecord {
	record := createNewPlayerRecord(username, team)
	timestamp := time.Now()
	record.GuestCreateTime = &timestamp
	return record
}

//////////////////////////////////////////////////////////////////
// Highscores

func generateRichestList(hub *Hub) HighScoreList {
	hub.richest.Lock()
	defer hub.richest.Unlock()
	if isOverNSecondsAgo(hub.richest.lastChecked, HIGHSCORE_CHECK_INTERVAL_IN_SECONDS) {
		records, _ := hub.db.getTopNPlayersByField("stats.peakWealth", 10)
		entries := make([]HighScoreEntry, 0)
		for _, record := range records {
			entries = append(entries, HighScoreEntry{
				Username: record.Username,
				StatNames: []string{
					"money",
				},
				StatValues: []string{
					strconv.Itoa(int(record.Stats.PeakWealth)),
				},
			})
		}
		hub.richest.Entries = entries
		hub.richest.lastChecked = time.Now()
	}
	return hub.richest.HighScoreList
}

func generateDeadliestList(hub *Hub) HighScoreList {
	hub.deadliest.Lock()
	defer hub.deadliest.Unlock()
	if isOverNSecondsAgo(hub.deadliest.lastChecked, HIGHSCORE_CHECK_INTERVAL_IN_SECONDS) {
		records, _ := hub.db.getTopNPlayersByField("stats.peakKillStreak", 10)
		entries := make([]HighScoreEntry, 0)
		for _, record := range records {
			entries = append(entries, HighScoreEntry{
				Username: record.Username,
				StatNames: []string{
					"Streak",
					"K/D",
				},
				StatValues: []string{
					strconv.Itoa(int(record.Stats.PeakKillStreak)),
					DivideIntsFloatToString(int(record.Stats.KillCount+record.Stats.KillCountNpc), int(record.Stats.DeathCount))},
			})
		}
		hub.deadliest.Entries = entries
		hub.deadliest.lastChecked = time.Now()
	}
	return hub.deadliest.HighScoreList
}

func DivideIntsFloatToString(a, b int) string {
	if b == 0 {
		return "NaN" // or handle the error as preferred
	}
	result := float64(a) / float64(b)
	return fmt.Sprintf("%.2f", result)
}

func generateMVPList(hub *Hub) HighScoreList {
	hub.mvp.Lock()
	defer hub.mvp.Unlock()
	if isOverNSecondsAgo(hub.mvp.lastChecked, HIGHSCORE_CHECK_INTERVAL_IN_SECONDS) {
		records, _ := hub.db.getTopNPlayersByField("stats.goalsScored", 10)
		entries := make([]HighScoreEntry, 0)
		for _, record := range records {
			entries = append(entries, HighScoreEntry{
				Username:   record.Username,
				StatNames:  []string{"Total Goals"},
				StatValues: []string{strconv.Itoa(int(record.Stats.GoalsScored))},
			})
		}
		hub.mvp.Entries = entries
		hub.mvp.lastChecked = time.Now()
	}
	return hub.mvp.HighScoreList
}

///////////////////////////////////////////////////////////
// Highscore Template Methods

func (hs HighScoreList) NextCategory() string {
	for i := range queryCategories {
		if hs.Category == queryCategories[i] {
			return queryCategories[mod(i+1, len(queryCategories))]
		}
	}
	return "next-invalid"
}

func (hs HighScoreList) PrevCategory() string {
	for i := range queryCategories {
		if hs.Category == queryCategories[i] {
			return queryCategories[mod(i-1, len(queryCategories))]
		}
	}
	return "prev-invalid"
}
