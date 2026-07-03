package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/markbates/goth/gothic"
	"go.mongodb.org/mongo-driver/bson"
)

const ALLOWED_HEADERS = "Content-Type, hx-current-url, hx-request, hx-target, hx-trigger"
const STATUS_CHECK_INTERVAL_IN_SECONDS = 5

// ///////////////////////////////////////////
// World Select and Status

type WorldSelectBanner struct {
	ServerName   string
	DomainName   string
	FuchsiaCount int
	SkyBlueCount int
	Vacancy      bool
}

func createWorldSelectHandler(config *Configuration) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		_, ok := getUserIdFromSession(r)
		if !ok {
			tmpl.ExecuteTemplate(w, "homepage", false) // this causes logo to duplicate
			return
		}
		tmpl.ExecuteTemplate(w, "world-select", config.domains)
	}
}

func (app *App) worldSelectHandler(w http.ResponseWriter, r *http.Request) {
	identifier, ok := getUserIdFromSession(r)
	if !ok {
		tmpl.ExecuteTemplate(w, "homepage", false)
		return
	}
	if app.platform == nil {
		createWorldSelectHandler(app.config)(w, r)
		return
	}
	worlds, err := app.platform.store.listWorlds(r.Context(), bson.M{"visibility": "public", "moderationState": "active", "publishedReleaseId": bson.M{"$ne": ""}})
	if err != nil {
		http.Error(w, "Unable to load worlds", http.StatusInternalServerError)
		return
	}
	type entry struct {
		ID, Name, Route  string
		Running, CanEdit bool
	}
	entries := make([]entry, 0, len(worlds))
	admin := app.config.isAdminIdentifier(identifier)
	for _, world := range worlds {
		route := world.ID
		if world.Slug != "" {
			route = world.Slug
		}
		_, running := app.platform.manager.Info(world.ID)
		entries = append(entries, entry{ID: world.ID, Name: world.Name, Route: route, Running: running, CanEdit: admin || world.OwnerID == identifier})
	}
	tmpl.ExecuteTemplate(w, "world-select-platform", entries)
}

func (world *World) statusHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", world.config.originForCORS())
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", ALLOWED_HEADERS)
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method == http.MethodGet {
		world.getStatus(w, r)
		return
	}
}

func (world *World) getStatus(w http.ResponseWriter, r *http.Request) {
	id, ok := getUserIdFromSession(r)
	if !ok {
		io.WriteString(w, "<div>Invalid Sign in</div>")
		return
	}
	if !world.config.guestsEnabled.Load() && idBelongsToGuest(id) {
		io.WriteString(w, "<div>Guests not allowed in server "+world.config.serverName+"</div>")
		return
	}
	world.teamPlayerStatus.Lock()
	defer world.teamPlayerStatus.Unlock()
	if isOverNSecondsAgo(world.teamPlayerStatus.lastStatusCheck, STATUS_CHECK_INTERVAL_IN_SECONDS) {
		world.wPlayerMutex.Lock()
		defer world.wPlayerMutex.Unlock()
		world.teamPlayerStatus.fuchsiaPlayerCount = world.teamQuantities["fuchsia"]
		world.teamPlayerStatus.skyBluePlayerCount = world.teamQuantities["sky-blue"]
		world.teamPlayerStatus.lastStatusCheck = time.Now()
	}
	statusDiv := WorldSelectBanner{
		ServerName:   world.config.serverName,
		DomainName:   world.config.domainName,
		FuchsiaCount: world.teamPlayerStatus.fuchsiaPlayerCount,
		SkyBlueCount: world.teamPlayerStatus.skyBluePlayerCount,
		Vacancy:      vacancyOfLockedWorldStatus(&world.teamPlayerStatus),
	}
	tmpl.ExecuteTemplate(w, "world-status", statusDiv)
}

func idBelongsToGuest(id string) bool {
	return strings.HasPrefix(id, "guest:")
}

func vacancyOfLockedWorldStatus(status *TeamPlayerStatus) bool {
	return status.fuchsiaPlayerCount < CAPACITY_PER_TEAM || status.skyBluePlayerCount < CAPACITY_PER_TEAM
}

var unavailableMessage = `Server unavailable :( <a href="#" hx-get="/worlds" hx-target="#page"> Try again</a>`

func unavailable(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, unavailableMessage)
}

var wrongMessage = `Something went wrong :( <a href="#" hx-get="/worlds" hx-target="#page">Choose other world</a>`

func somethingWentWrong(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, wrongMessage)
}

// ///////////////////////////////////////////
// Player Sign-in and Create

func (world *World) playHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", world.config.originForCORS())
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", ALLOWED_HEADERS)
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method == http.MethodPost {
		world.postPlay(w, r)
		return
	}
}

func (world *World) postPlay(w http.ResponseWriter, r *http.Request) {
	id, ok := getUserIdFromSession(r)
	if !ok {
		// Redirect to "/" to match homeserver guest config (As opposed to this world's config)
		tmpl.ExecuteTemplate(w, "homepage", world.config.guestsEnabled.Load())
		return
	}
	if world.config.worldLifecycle == LifecycleUntilEmpty && id != world.config.worldOwnerID {
		var runtime struct {
			Draining bool `bson:"draining"`
		}
		if err := world.db.runtimeInstances.FindOne(r.Context(), bson.M{"_id": world.config.worldID}).Decode(&runtime); err == nil && runtime.Draining {
			http.Error(w, "World is closing to new players", http.StatusServiceUnavailable)
			return
		}
	}
	var userRecord *UserRecord
	if idBelongsToGuest(id) {
		world.postPlayAsGuest(w, r)
		return
	}
	userRecord = world.db.getAuthorizedUserById(id)
	if userRecord == nil {
		// Could imply tampered/invalid cookie.
		// Has happened when db record is lost/destroyed
		// Is confusing if this happens because you get a blank page with no explanation
		return
	}

	if isUserCurrentlyBanned(userRecord) {
		if userRecord.BanExpiresAt != nil {
			io.WriteString(w, "Account temporarily banned. Try again later.")
		} else {
			io.WriteString(w, "Account is banned.")
		}
		return
	}
	if isUserBanExpired(userRecord) {
		if err := world.db.clearBanForIdentifier(userRecord.Identifier); err != nil {
			logger.Warn().Err(err).Msg("Failed to clear expired ban for: " + userRecord.Identifier)
		}
	}

	if userRecord.Username == "" {
		colorPage := struct {
			DomainName        string
			SuggestedUsername string
		}{
			DomainName:        world.config.domainName,
			SuggestedUsername: world.db.UniqueName(),
		}
		tmpl.ExecuteTemplate(w, "choose-your-color", colorPage)
	}

	record, err := world.db.getWorldPlayerRecord(world.config.worldID, userRecord.Username)
	if err != nil {
		if world.config.worldID == legacyWorldID {
			logger.Warn().Msg("User: " + id + " found but not corresponding Player with username: " + userRecord.Username)
			io.WriteString(w, "Unable to sign in")
			return
		}
		record = world.initialPlayerRecord(id, userRecord.Username)
		if err := world.db.InsertPlayerRecord(record); err != nil {
			io.WriteString(w, "Unable to create world profile")
			return
		}
	}

	receipt := world.initiateLogin(record, id)
	tmpl.ExecuteTemplate(w, "player-page", receipt)
}

type LoginReceipt struct {
	LoginRequest *LoginRequest
	DomainName   string
}

func (w *World) initiateLogin(record PlayerRecord, userID ...string) *LoginReceipt {
	loginReq := createLoginRequest(record)
	if len(userID) > 0 {
		loginReq.UserID = userID[0]
	}
	w.addIncoming(loginReq)
	logger.Info().Msg("loginRequest for: " + loginReq.Record.Username)
	return &LoginReceipt{
		LoginRequest: loginReq,
		DomainName:   w.config.domainName,
	}
}

func (world *World) postPlayAsGuest(w http.ResponseWriter, r *http.Request) {
	if !world.config.guestsEnabled.Load() {
		io.WriteString(w, "Guests prohibited.")
		return
	}
	id, ok := getUserIdFromSession(r)
	if !ok {
		io.WriteString(w, "Unexpected Error.")
		return
	}
	if !idBelongsToGuest(id) {
		io.WriteString(w, "Unexpected Error.")
		return
	}
	record, err := world.db.getWorldPlayerRecord(world.config.worldID, id)
	if err != nil {
		if world.config.worldID == legacyWorldID {
			logger.Warn().Msg("Record for guest with id: " + id + " not found.")
			io.WriteString(w, "Unable to sign in")
			return
		}
		record = world.initialPlayerRecord(id, id)
		if err := world.db.InsertPlayerRecord(record); err != nil {
			io.WriteString(w, "Unable to create world profile")
			return
		}
	}
	receipt := world.initiateLogin(record, id)
	tmpl.ExecuteTemplate(w, "player-page", receipt)
}

func (db *DB) postNew(w http.ResponseWriter, r *http.Request) {
	id, ok := getUserIdFromSession(r)
	if !ok {
		tmpl.ExecuteTemplate(w, "homepage", false)
		return
	}
	userRecord := db.getAuthorizedUserById(id)
	if userRecord == nil {
		// deeply confusing
		// Could imply hacked cookie?
		return
	}

	logger.Info().Msg("New player request from: " + userRecord.Username)

	props, ok := requestToProperties(r)
	if !ok {
		logger.Debug().Msg("invalid props")
		tmpl.ExecuteTemplate(w, "homepage", false)
		return
	}

	team := props["player-team"]
	usernameEncoded := props["player-name"]
	username, err := url.QueryUnescape(usernameEncoded)
	if err != nil {
		logger.Error().Err(err).Msg("Error decoding username:" + username)
		return
	}

	if !validUsername(username) {
		io.WriteString(w, divBottomInvalid("Invalid Username"))
		return
	}

	desiredHostUrlEncoded := props["desired-host"]
	desiredHost, err := url.QueryUnescape(desiredHostUrlEncoded)
	if err != nil {
		logger.Error().Err(err).Msg("Error decoding host:")
		return
	}

	if !validTeam(team) {
		io.WriteString(w, divBottomInvalid("Invalid Player Color"))
		return
	}
	// not atomic
	if db.foundUsername(username) {
		io.WriteString(w, divBottomInvalid("Username unavailable. Try again."))
		return
	}

	record := createNewPlayerRecord(username, team)
	record.UserID = id
	record.WorldID = legacyWorldID
	err = db.InsertPlayerRecord(record)
	if err != nil {
		io.WriteString(w, divBottomInvalid("Error saving new player"))
		return
	}
	ok = db.updateUsernameForUserWithId(id, username)
	if !ok {
		io.WriteString(w, divBottomInvalid("Error, username not updated"))
		return
	}

	tmpl.ExecuteTemplate(w, "post-play-on-load", desiredHost)
}

func createNewPlayerRecord(username, team string) PlayerRecord {
	return PlayerRecord{
		Username:  username,
		WorldID:   legacyWorldID,
		Team:      team,
		Health:    100,
		StageName: "tutorial1:0-0",
		X:         3,
		Y:         3,
		Money:     80,
	}
}

func (world *World) initialPlayerRecord(userID, username string) PlayerRecord {
	manifest := world.config.manifest
	team := manifest.DefaultTeam
	if team == "" {
		team = "sky-blue"
	}
	entry := manifest.Entry
	if entry.Stage == "" {
		entry = WorldLocation{Stage: "tutorial1:0-0", Y: 3, X: 3}
	}
	return PlayerRecord{Username: username, UserID: userID, WorldID: world.config.worldID, Team: team, Health: 100, Money: 80, StageName: entry.Stage, Y: entry.Y, X: entry.X}
}

func validUsername(username string) bool {
	if len(username) == 0 || len(username) >= 32 {
		return false
	}

	if strings.HasPrefix(username, "guest:") {
		return false
	}

	invalidChars := []string{"{", "}", "\"", "'", "`", "/", "[", "]", "<", ">", "\\", "\n", "\t", ":"}
	for _, char := range invalidChars {
		if strings.Contains(username, char) {
			return false
		}
	}

	return true
}

func validTeam(team string) bool {
	validTeams := []string{"fuchsia", "sky-blue"}
	for i := range validTeams {
		if validTeams[i] == team {
			return true
		}
	}
	return false
}

/////////////////////////////////////////////
// Stats

func (world *World) getStats(w http.ResponseWriter, r *http.Request) {
	world.wPlayerMutex.Lock()
	defer world.wPlayerMutex.Unlock()
	out := fmt.Sprintf("World Player Count: %d\n", len(world.worldPlayers))
	for key, val := range world.teamQuantities {
		out += fmt.Sprintf("%s: %d\n", key, val)
	}
	io.WriteString(w, out)
}

/////////////////////////////////////////////
//  Oauth

func auth(w http.ResponseWriter, r *http.Request) {
	/*
		 // Force Google to show account selection
		q := r.URL.Query()
		q.Add("prompt", "select_account")
		r.URL.RawQuery = q.Encode()
	*/
	gothic.BeginAuthHandler(w, r)
}

func (db *DB) callback(w http.ResponseWriter, r *http.Request) {
	user, err := gothic.CompleteUserAuth(w, r)
	if err != nil {
		logger.Error().Err(err).Msg("Callback error: ")
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	logger.Info().Msg("New callback from: " + user.UserID)
	if user.UserID == "" || user.Provider == "" {
		logger.Warn().Msg(fmt.Sprintf("Invalid id: %s or provider %s ", user.UserID, user.Provider))
	}

	identifier := user.Provider + ":" + user.UserID
	userRecord := db.getAuthorizedUserById(identifier)
	if userRecord == nil {
		logger.Info().Msg("Creating new user with identifier: " + identifier)
		newUser := UserRecord{Identifier: identifier, Username: "", CreationEmail: user.Email, Created: time.Now(), LastLogin: time.Now()}
		err := db.insertAuthorizedUser(newUser)
		if err != nil {
			logger.Warn().Msg("New User creation in mongo failed")
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}
	} else {
		err := db.updateLastLoginForUserWithId(identifier)
		if err != nil {
			logger.Warn().Err(err).Msg("Unable to update user lastLogin for identifier: " + identifier)
		}
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

	http.Redirect(w, r, "/", http.StatusFound) // redirects.
}

/////////////////////////////////////////////
// Integration Endpoint

// Add Basic Auth / real auth ?
func (world *World) postHorribleBypass(w http.ResponseWriter, r *http.Request) {
	secret := os.Getenv("AUTO_PLAYER_PASSWORD")
	if secret == "" {
		logger.Warn().Msg("Bypass is disabled - but has been requested.")
		return
	}
	props, ok := requestToProperties(r)
	if !ok {
		logger.Debug().Msg("invalid props")
		return
	}
	if props["secret"] != secret {
		logger.Warn().Msg("Bypass is disabled - but has been requested.")
		return
	}
	countString := props["count"]
	count, err := strconv.Atoi(countString)
	if err != nil {
		logger.Warn().Msg("Invalid count for player bypass")
		return
	}
	username := props["username"]
	stage := props["stagename"]
	team := props["team"]
	tokens := make([]string, 0, count)
	for i := 0; i < count; i++ {
		iStr := strconv.Itoa(i) // Add some easy regex match condition
		record := PlayerRecord{Username: username + iStr, Health: 50, Y: 12, X: 5, StageName: stage, Team: team}
		// Make optional
		world.db.InsertPlayerRecord(record)
		loginRequest := createLoginRequest(record)
		world.addIncoming(loginRequest)
		logger.Debug().Msg("New bypass request for: " + loginRequest.Token)
		tokens = append(tokens, loginRequest.Token)
	}
	io.WriteString(w, "[\""+strings.Join(tokens, "\",\"")+"\"]")
}

////////////////////////////////////
// Utilities for templates

func (record PlayerRecord) HeartsFromRecord() string {
	return getHeartsFromHealth(record.Health)
}
