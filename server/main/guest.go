package main

import (
	"encoding/hex"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

type GuestLimiter struct {
	seen sync.Map // key:string -> time.Time
}

const GUEST_WINDOW = 2 * time.Minute

func (app *App) AllowGuest(key string) bool {
	if app.guestLimiter == nil {
		return true
	}
	now := time.Now()
	if v, ok := app.guestLimiter.seen.Load(key); ok && now.Sub(v.(time.Time)) < GUEST_WINDOW {
		return false
	}
	app.guestLimiter.seen.Store(key, now)
	return true
}

func (app *App) peekPermission(key string) bool {
	if app.guestLimiter == nil {
		return true
	}
	now := time.Now()
	if v, ok := app.guestLimiter.seen.Load(key); ok && now.Sub(v.(time.Time)) < GUEST_WINDOW {
		return false
	}
	return true
}

func (app *App) guestsHandler(w http.ResponseWriter, r *http.Request) {
	if !app.config.guestsEnabled.Load() {
		tmpl.ExecuteTemplate(w, "homepage", app.config.guestsEnabled.Load())
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	switch r.Method {
	case "GET":
		allowAttempt := app.peekPermission(clientIP(r))
		tmpl.ExecuteTemplate(w, "guests", allowAttempt)
	case "POST":
		if !app.AllowGuest(clientIP(r)) {
			w.WriteHeader(http.StatusTooManyRequests)
			tmpl.ExecuteTemplate(w, "guests-limited", nil)
			return
		}

		app.storeNewGuestSession(w, r)

		w.Header().Set("HX-Redirect", "/")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Request was successful - Redirecting"))
	}
}

func clientIP(r *http.Request) string {
	// Honour X‑Forwarded‑For / X‑Real‑IP in case of reverse proxy.
	if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
		return strings.Split(fwd, ",")[0]
	}
	if real := r.Header.Get("X-Real-Ip"); real != "" {
		return real
	}
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	return ip
}

func (app *App) storeNewGuestSession(w http.ResponseWriter, r *http.Request) {
	hexid, err := randomHex16()
	if err != nil {
		return
	}
	identifier := "guest:" + hexid
	team := "sky-blue"
	if rand.Intn(2) == 1 {
		team = "fuchsia"
	}

	// Add Capcha check

	record := createNewGuestPlayerRecord(identifier, team)
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
	b := make([]byte, 6) // 2 hex chars/byte -> 12 hex chars
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
