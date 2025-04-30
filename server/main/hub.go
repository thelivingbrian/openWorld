package main

import (
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"
)

const HIGHSCORE_CHECK_INTERVAL_IN_SECONDS = 15

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
	BorderColor string
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
var queryCategories = [3][2]string{
	{"Richest", "money"},
	{"Deadliest", "killCount"},
	{"MVP", "goalsScored"},
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

func homeHandler(w http.ResponseWriter, r *http.Request) {
	logger.Info().Msg("Home page accessed.") // Replace with metric
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	_, identifierFound := getUserIdFromSession(r)
	tmpl.ExecuteTemplate(w, "homepage", identifierFound)
}

func aboutHandler(w http.ResponseWriter, r *http.Request) {
	logger.Info().Msg("About page accessed.") // Metric > log
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	tmpl.ExecuteTemplate(w, "about", nil)
}

func (hub *Hub) highscoreHandler(w http.ResponseWriter, r *http.Request) {
	logger.Info().Msg("Highscore page accessed.")

	queryValues := r.URL.Query()
	category := queryValues.Get("category")

	var scores HighScoreList
	switch category {
	case "money":
		scores = generateRichestList(hub)
	case "killCount":
		scores = generateDeadliestList(hub)
	case "goalsScored":
		scores = generateMVPList(hub)
	default:
		break
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	tmpl.ExecuteTemplate(w, "highscore", scores)
}

//////////////////////////////////////////////////////////////////
// Highscores

func generateRichestList(hub *Hub) HighScoreList {
	hub.richest.Lock()
	defer hub.richest.Unlock()
	if isOverNSecondsAgo(hub.richest.lastChecked, HIGHSCORE_CHECK_INTERVAL_IN_SECONDS) {
		records, _ := hub.db.getTopNPlayersByField("money", 10)
		entries := make([]HighScoreEntry, 0)
		for _, record := range records {
			entries = append(entries, HighScoreEntry{
				Username:   record.Username,
				StatNames:  []string{"Money"},
				StatValues: []string{strconv.Itoa(record.Money)},
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
		records, _ := hub.db.getTopNPlayersByField("killCount", 10)
		entries := make([]HighScoreEntry, 0)
		for _, record := range records {
			entries = append(entries, HighScoreEntry{
				Username:   record.Username,
				StatNames:  []string{"Kill-Count", "K/D"},
				StatValues: []string{strconv.Itoa(int(record.Stats.KillCount)), DivideIntsFloatToString(int(record.Stats.KillCount), int(record.Stats.DeathCount))},
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
		records, _ := hub.db.getTopNPlayersByField("goalsScored", 10)
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
		if hs.Category == queryCategories[i][0] {
			return queryCategories[mod(i+1, len(queryCategories))][1]
		}
	}
	return "next-invalid"
}

func (hs HighScoreList) PrevCategory() string {
	for i := range queryCategories {
		if hs.Category == queryCategories[i][0] {
			return queryCategories[mod(i-1, len(queryCategories))][1]
		}
	}
	return "prev-invalid"
}
