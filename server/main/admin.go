package main

import (
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

type AdminPageData struct {
	AdminIdentifier string
	Message         string
	Players         []AdminPlayerRow
	Stages          []AdminStageRow
	Session         AdminSessionRow
	SelectedPlayer  *AdminPlayerDetails
	SelectedStage   *AdminStageRow
}

type AdminPlayerRow struct {
	Username  string
	Team      string
	StageName string
	Y         int
	X         int
	Health    int64
	Money     int64
}

type AdminStageRow struct {
	StageName    string
	PlayerCount  int
	StageHeight  int
	StageWidth   int
	WeatherClass string
}

type AdminSessionRow struct {
	SessionStartTime       time.Time
	PeakSessionPlayerCount int64
	PeakSessionKillStreak  int64
	PeakSessionKiller      string
	TotalSessionLogins     int64
	TotalSessionLogouts    int64
	Scoreboard             map[string]int
	TeamCounts             map[string]int
}

type AdminPlayerDetails struct {
	Username         string
	Team             string
	StageName        string
	Y                int
	X                int
	Health           int64
	Money            int64
	Accomplishments  string
	LastLogin        time.Time
	LastLogout       time.Time
	UserIdentifier   string
	UserBanned       bool
	UserBanReason    string
	UserBanExpiresAt string
}

func (world *World) adminHandler(w http.ResponseWriter, r *http.Request) {
	adminIdentifier, ok := world.requireAdminIdentifier(w, r)
	if !ok {
		return
	}
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	message := r.URL.Query().Get("message")
	selectedUsername := r.URL.Query().Get("player")
	selectedStageName := r.URL.Query().Get("stage")

	pageData := world.adminSnapshot(adminIdentifier, message, selectedUsername, selectedStageName)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	tmpl.ExecuteTemplate(w, "admin", pageData)
}

func (world *World) adminUpdatePlayerHandler(w http.ResponseWriter, r *http.Request) {
	adminIdentifier, ok := world.requireAdminIdentifier(w, r)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	props, ok := requestToProperties(r)
	if !ok {
		world.redirectAdminWithMessage(w, r, "Invalid request payload", "")
		return
	}

	username := decodedFormProperty(props, "username")
	if username == "" {
		world.redirectAdminWithMessage(w, r, "Missing username", "")
		return
	}

	record, err := world.db.getWorldPlayerRecord(world.config.worldID, username)
	if err != nil {
		world.logAdminAction(AdminActionRecord{
			ActionType:        "player.update",
			ActingAdmin:       adminIdentifier,
			TargetPlayer:      username,
			ResultStatus:      "failed",
			ResultDescription: "player record not found",
			Payload:           bson.M{"username": username},
			Created:           time.Now().UTC(),
		})
		world.redirectAdminWithMessage(w, r, "Player record not found", username)
		return
	}

	money, err := strconv.ParseInt(decodedFormProperty(props, "money"), 10, 64)
	if err != nil {
		world.redirectAdminWithMessage(w, r, "Invalid money value", username)
		return
	}
	health, err := strconv.ParseInt(decodedFormProperty(props, "health"), 10, 64)
	if err != nil {
		world.redirectAdminWithMessage(w, r, "Invalid health value", username)
		return
	}
	y, err := strconv.Atoi(decodedFormProperty(props, "y"))
	if err != nil {
		world.redirectAdminWithMessage(w, r, "Invalid Y coordinate", username)
		return
	}
	x, err := strconv.Atoi(decodedFormProperty(props, "x"))
	if err != nil {
		world.redirectAdminWithMessage(w, r, "Invalid X coordinate", username)
		return
	}

	team := decodedFormProperty(props, "team")
	if !world.validTeam(team) {
		world.redirectAdminWithMessage(w, r, "Invalid team", username)
		return
	}

	stageName := decodedFormProperty(props, "stagename")
	if stageName == "" {
		world.redirectAdminWithMessage(w, r, "Stage name is required", username)
		return
	}

	accomplishments := parseAccomplishmentsCSV(decodedFormProperty(props, "accomplishments"), record.Accomplishments)

	updates := bson.M{
		"money":           money,
		"health":          health,
		"team":            team,
		"stagename":       stageName,
		"y":               y,
		"x":               x,
		"accomplishments": accomplishments,
	}
	if err := world.db.adminUpdatePlayerRecord(world.config.worldID, username, updates); err != nil {
		world.logAdminAction(AdminActionRecord{
			ActionType:        "player.update",
			ActingAdmin:       adminIdentifier,
			TargetPlayer:      username,
			ResultStatus:      "failed",
			ResultDescription: err.Error(),
			Payload:           updates,
			Created:           time.Now().UTC(),
		})
		world.redirectAdminWithMessage(w, r, "Failed to update player record", username)
		return
	}

	if livePlayer := world.getPlayerByUsername(username); livePlayer != nil {
		world.applyAdminLiveUpdate(livePlayer, money, health, team, stageName, y, x, accomplishments)
	}

	world.logAdminAction(AdminActionRecord{
		ActionType:   "player.update",
		ActingAdmin:  adminIdentifier,
		TargetPlayer: username,
		Payload:      updates,
		ResultStatus: "success",
		Created:      time.Now().UTC(),
	})

	world.redirectAdminWithMessage(w, r, "Player updated", username)
}

func (world *World) adminBanPlayerHandler(w http.ResponseWriter, r *http.Request) {
	adminIdentifier, ok := world.requireAdminIdentifier(w, r)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	props, ok := requestToProperties(r)
	if !ok {
		world.redirectAdminWithMessage(w, r, "Invalid request payload", "")
		return
	}

	username := decodedFormProperty(props, "username")
	if username == "" {
		world.redirectAdminWithMessage(w, r, "Missing username", "")
		return
	}

	reason := decodedFormProperty(props, "reason")
	if reason == "" {
		reason = "No reason provided"
	}

	targetUser := world.db.getAuthorizedUserByUsername(username)
	targetIdentifier := ""
	if targetUser != nil {
		targetIdentifier = targetUser.Identifier
		if err := world.db.upsertBanForIdentifier(targetIdentifier, adminIdentifier, reason, nil); err != nil {
			world.logAdminAction(AdminActionRecord{
				ActionType:        "player.ban",
				ActingAdmin:       adminIdentifier,
				TargetPlayer:      username,
				TargetIdentifier:  targetIdentifier,
				Payload:           bson.M{"reason": reason, "type": "permanent"},
				ResultStatus:      "failed",
				ResultDescription: err.Error(),
				Created:           time.Now().UTC(),
			})
			world.redirectAdminWithMessage(w, r, "Failed to save ban", username)
			return
		}
	}

	wasOnline := world.kickOnlinePlayer(username)

	resultDescription := "ban applied"
	if targetUser == nil {
		resultDescription = "player kicked; no authorized user record found to persist ban"
	}
	if wasOnline {
		resultDescription += "; player was online"
	} else {
		resultDescription += "; player not currently online"
	}
	world.logAdminAction(AdminActionRecord{
		ActionType:        "player.ban",
		ActingAdmin:       adminIdentifier,
		TargetPlayer:      username,
		TargetIdentifier:  targetIdentifier,
		Payload:           bson.M{"reason": reason, "type": "permanent"},
		ResultStatus:      "success",
		ResultDescription: resultDescription,
		Created:           time.Now().UTC(),
	})

	world.redirectAdminWithMessage(w, r, "Ban applied and player kicked", username)
}

func (world *World) adminKickPlayerHandler(w http.ResponseWriter, r *http.Request) {
	adminIdentifier, ok := world.requireAdminIdentifier(w, r)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	props, ok := requestToProperties(r)
	if !ok {
		world.redirectAdminWithMessage(w, r, "Invalid request payload", "")
		return
	}

	username := decodedFormProperty(props, "username")
	if username == "" {
		world.redirectAdminWithMessage(w, r, "Missing username", "")
		return
	}

	reason := decodedFormProperty(props, "reason")
	if reason == "" {
		reason = "No reason provided"
	}
	minutesText := decodedFormProperty(props, "durationMinutes")
	durationMinutes := 0
	if minutesText != "" {
		parsedDuration, parseErr := strconv.Atoi(minutesText)
		durationMinutes = parsedDuration
		if parseErr != nil || durationMinutes < 0 {
			world.redirectAdminWithMessage(w, r, "Invalid kick duration", username)
			return
		}
	}

	targetUser := world.db.getAuthorizedUserByUsername(username)
	targetIdentifier := ""
	if targetUser != nil {
		targetIdentifier = targetUser.Identifier
	}

	if durationMinutes > 0 && targetIdentifier != "" {
		expiresAt := time.Now().UTC().Add(time.Duration(durationMinutes) * time.Minute)
		if err := world.db.upsertBanForIdentifier(targetIdentifier, adminIdentifier, reason, &expiresAt); err != nil {
			world.logAdminAction(AdminActionRecord{
				ActionType:        "player.kick",
				ActingAdmin:       adminIdentifier,
				TargetPlayer:      username,
				TargetIdentifier:  targetIdentifier,
				Payload:           bson.M{"reason": reason, "durationMinutes": durationMinutes, "type": "temporary"},
				ResultStatus:      "failed",
				ResultDescription: err.Error(),
				Created:           time.Now().UTC(),
			})
			world.redirectAdminWithMessage(w, r, "Failed to apply kick duration", username)
			return
		}
	}

	wasOnline := world.kickOnlinePlayer(username)
	resultDescription := "kick requested"
	if wasOnline {
		resultDescription += "; player was online"
	} else {
		resultDescription += "; player not currently online"
	}
	if durationMinutes > 0 && targetIdentifier == "" {
		resultDescription += "; temporary duration not persisted (no authorized user record)"
	}
	world.logAdminAction(AdminActionRecord{
		ActionType:        "player.kick",
		ActingAdmin:       adminIdentifier,
		TargetPlayer:      username,
		TargetIdentifier:  targetIdentifier,
		Payload:           bson.M{"reason": reason, "durationMinutes": durationMinutes, "type": "temporary-or-none"},
		ResultStatus:      "success",
		ResultDescription: resultDescription,
		Created:           time.Now().UTC(),
	})
	if durationMinutes > 0 {
		world.redirectAdminWithMessage(w, r, "Player kicked with temporary lockout", username)
		return
	}
	world.redirectAdminWithMessage(w, r, "Player kicked", username)
}

func (world *World) adminSnapshot(adminIdentifier, message, selectedUsername, selectedStageName string) AdminPageData {
	players := world.snapshotPlayers()
	stages := world.snapshotStages()
	selected := world.getAdminPlayerDetails(selectedUsername)
	selectedStage := selectAdminStage(stages, selectedStageName)

	return AdminPageData{
		AdminIdentifier: adminIdentifier,
		Message:         message,
		Players:         players,
		Stages:          stages,
		Session: AdminSessionRow{
			SessionStartTime:       world.sessionStats.sessionStartTime,
			PeakSessionPlayerCount: world.sessionStats.peakSessionPlayerCount.Load(),
			PeakSessionKillStreak:  world.sessionStats.peakSessionKillStreak.Load(),
			PeakSessionKiller:      world.sessionStats.peakSessionKiller,
			TotalSessionLogins:     world.sessionStats.TotalSessionLogins.Load(),
			TotalSessionLogouts:    world.sessionStats.TotalSessionLogouts.Load(),
			Scoreboard:             world.leaderBoard.scoreboard.Export(),
			TeamCounts:             CopyTeamQuantities(world),
		},
		SelectedPlayer: selected,
		SelectedStage:  selectedStage,
	}
}

func selectAdminStage(stages []AdminStageRow, stageName string) *AdminStageRow {
	stageName = strings.TrimSpace(stageName)
	for i := range stages {
		if stages[i].StageName == stageName {
			return &stages[i]
		}
	}
	return nil
}

func (world *World) snapshotPlayers() []AdminPlayerRow {
	world.wPlayerMutex.Lock()
	defer world.wPlayerMutex.Unlock()

	rows := make([]AdminPlayerRow, 0, len(world.worldPlayers))
	for _, player := range world.worldPlayers {
		tile := player.getTileSync()
		stageName := ""
		y, x := 0, 0
		if tile != nil && tile.stage != nil {
			stageName = tile.stage.name
			y = tile.y
			x = tile.x
		}
		rows = append(rows, AdminPlayerRow{
			Username:  player.username,
			Team:      player.getTeamNameSync(),
			StageName: stageName,
			Y:         y,
			X:         x,
			Health:    player.health.Load(),
			Money:     player.money.Load(),
		})
	}

	sort.Slice(rows, func(i, j int) bool {
		return rows[i].Username < rows[j].Username
	})

	return rows
}

func (world *World) snapshotStages() []AdminStageRow {
	world.wStageMutex.Lock()
	defer world.wStageMutex.Unlock()

	rows := make([]AdminStageRow, 0, len(world.worldStages))
	for name, stage := range world.worldStages {
		if stage == nil {
			continue
		}
		stage.playerMutex.RLock()
		playerCount := len(stage.playerMap)
		stage.playerMutex.RUnlock()

		height, width := 0, 0
		if len(stage.tiles) > 0 {
			height = len(stage.tiles)
			width = len(stage.tiles[0])
		}
		rows = append(rows, AdminStageRow{
			StageName:    name,
			PlayerCount:  playerCount,
			StageHeight:  height,
			StageWidth:   width,
			WeatherClass: stage.weather,
		})
	}

	sort.Slice(rows, func(i, j int) bool {
		if rows[i].PlayerCount == rows[j].PlayerCount {
			return rows[i].StageName < rows[j].StageName
		}
		return rows[i].PlayerCount > rows[j].PlayerCount
	})
	return rows
}

func (world *World) getAdminPlayerDetails(username string) *AdminPlayerDetails {
	username = strings.TrimSpace(username)
	if username == "" {
		return nil
	}

	record, err := world.db.getWorldPlayerRecord(world.config.worldID, username)
	if err != nil {
		return nil
	}

	accomplishments := make([]string, 0, len(record.Accomplishments))
	for name := range record.Accomplishments {
		accomplishments = append(accomplishments, name)
	}
	sort.Strings(accomplishments)

	details := &AdminPlayerDetails{
		Username:        record.Username,
		Team:            record.Team,
		StageName:       record.StageName,
		Y:               record.Y,
		X:               record.X,
		Health:          record.Health,
		Money:           record.Money,
		Accomplishments: strings.Join(accomplishments, ", "),
		LastLogin:       record.LastLogin,
		LastLogout:      record.LastLogout,
	}

	userRecord := world.db.getAuthorizedUserByUsername(username)
	if userRecord != nil {
		details.UserIdentifier = userRecord.Identifier
		details.UserBanned = isUserCurrentlyBanned(userRecord)
		details.UserBanReason = userRecord.BanReason
		if userRecord.BanExpiresAt != nil {
			details.UserBanExpiresAt = userRecord.BanExpiresAt.UTC().Format(time.RFC3339)
		}
	}

	return details
}

func (world *World) getPlayerByUsername(username string) *Player {
	world.wPlayerMutex.Lock()
	defer world.wPlayerMutex.Unlock()
	for _, player := range world.worldPlayers {
		if player.username == username {
			return player
		}
	}
	return nil
}

func (world *World) applyAdminLiveUpdate(player *Player, money, health int64, team, stageName string, y, x int, accomplishments map[string]Accomplishment) {
	previousTeam := player.getTeamNameSync()
	if previousTeam != team {
		player.viewLock.Lock()
		player.team = team
		player.viewLock.Unlock()

		world.wPlayerMutex.Lock()
		world.teamQuantities[previousTeam] = max(0, world.teamQuantities[previousTeam]-1)
		world.teamQuantities[team] = world.teamQuantities[team] + 1
		world.wPlayerMutex.Unlock()
		updateIconForAllIfTangible(player)
	}

	player.health.Store(health)
	player.money.Store(money)

	player.accomplishments.Lock()
	player.accomplishments.Accomplishments = accomplishments
	player.accomplishments.Unlock()

	targetStage := player.fetchStageSync(stageName)
	if targetStage != nil && validCoordinate(y, x, targetStage) {
		currentTile := player.getTileSync()
		if currentTile != nil && (currentTile.stage != targetStage || currentTile.y != y || currentTile.x != x) {
			player.transferBetween(currentTile, targetStage.tiles[y][x])
		}
	}

	player.updatePlayerHud()
	player.updateRecord()
}

func parseAccomplishmentsCSV(value string, existing map[string]Accomplishment) map[string]Accomplishment {
	if value == "" {
		return map[string]Accomplishment{}
	}

	out := make(map[string]Accomplishment)
	tokens := strings.Split(value, ",")
	for _, token := range tokens {
		name := strings.TrimSpace(token)
		if name == "" {
			continue
		}
		if existing != nil {
			if accomplishment, ok := existing[name]; ok {
				out[name] = accomplishment
				continue
			}
		}
		out[name] = Accomplishment{Name: name, AcquiredAt: time.Now().UTC()}
	}
	return out
}

func decodedFormProperty(props map[string]string, key string) string {
	raw := strings.TrimSpace(props[key])
	decoded, err := url.QueryUnescape(raw)
	if err != nil {
		return raw
	}
	return strings.TrimSpace(decoded)
}

func (world *World) requireAdminIdentifier(w http.ResponseWriter, r *http.Request) (string, bool) {
	identifier, ok := getUserIdFromSession(r)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Missing session"))
		return "", false
	}
	if !world.config.isAdminIdentifier(identifier) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("Admin access required"))
		return "", false
	}
	return identifier, true
}

func (world *World) redirectAdminWithMessage(w http.ResponseWriter, r *http.Request, message, selectedPlayer string) {
	messageEscaped := url.QueryEscape(message)
	target := "/admin?message=" + messageEscaped
	if selectedPlayer != "" {
		target += "&player=" + url.QueryEscape(selectedPlayer)
	}
	http.Redirect(w, r, target, http.StatusFound)
}

func (world *World) logAdminAction(record AdminActionRecord) {
	if err := world.db.saveAdminAction(record); err != nil {
		logger.Warn().Err(err).Msg("failed to save admin action")
	}
}

func (world *World) kickOnlinePlayer(username string) bool {
	if livePlayer := world.getPlayerByUsername(username); livePlayer != nil {
		sendUpdate(livePlayer, divLogOutResume("You were removed by an admin", world.config.domainName))
		go initiateLogout(livePlayer)
		return true
	}
	return false
}

func isUserBanExpired(user *UserRecord) bool {
	if user == nil || user.BanExpiresAt == nil {
		return false
	}
	return time.Now().UTC().After(*user.BanExpiresAt)
}

func isUserCurrentlyBanned(user *UserRecord) bool {
	if user == nil {
		return false
	}
	if user.BanReason == "" && user.BanExpiresAt == nil {
		return false
	}
	if isUserBanExpired(user) {
		return false
	}
	return true
}

func (session AdminSessionRow) SessionDuration() string {
	if session.SessionStartTime.IsZero() {
		return "unknown"
	}
	return time.Since(session.SessionStartTime).Round(time.Second).String()
}

func (details AdminPlayerDetails) LastLoginText() string {
	if details.LastLogin.IsZero() {
		return "never"
	}
	return details.LastLogin.UTC().Format(time.RFC3339)
}

func (details AdminPlayerDetails) LastLogoutText() string {
	if details.LastLogout.IsZero() {
		return "never"
	}
	return details.LastLogout.UTC().Format(time.RFC3339)
}

func (details AdminPlayerDetails) BanStatusText() string {
	if !details.UserBanned {
		return "not banned"
	}
	if details.UserBanExpiresAt == "" {
		return "permanent"
	}
	return fmt.Sprintf("until %s", details.UserBanExpiresAt)
}
