package main

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type requestWindow struct {
	started time.Time
	count   int
}
type WorldPlatform struct {
	store        *WorldStore
	manager      *RuntimeManager
	compileSlots chan struct{}
	designDir    string
	rateMu       sync.Mutex
	rate         map[string]requestWindow
}

func newWorldPlatform(ctx context.Context, db *DB, config *Configuration) (*WorldPlatform, error) {
	store, err := newWorldStore(db)
	if err != nil {
		return nil, err
	}
	if err := store.ensureIndexes(ctx); err != nil {
		return nil, err
	}
	executable, err := currentExecutable()
	if err != nil {
		return nil, err
	}
	cacheDir := strings.TrimSpace(osEnv("WORLD_CACHE_DIR", filepath.Join(".", "data", "world-cache")))
	designDir := strings.TrimSpace(osEnv("WORLD_DESIGN_DIR", filepath.Join(".", "design")))
	manager := newRuntimeManager(ctx, store, &ProcessLauncher{Executable: executable, Config: config}, cacheDir)
	platform := &WorldPlatform{store: store, manager: manager, compileSlots: make(chan struct{}, 2), designDir: designDir, rate: map[string]requestWindow{}}
	go platform.restorePersistentWorlds(ctx)
	return platform, nil
}

func (p *WorldPlatform) restorePersistentWorlds(ctx context.Context) {
	worlds, err := p.store.listWorlds(ctx, bson.M{"lifecycle": LifecyclePersistent, "moderationState": "active", "publishedReleaseId": bson.M{"$ne": ""}})
	if err != nil {
		logger.Error().Err(err).Msg("Failed to list persistent worlds")
		return
	}
	for i := range worlds {
		if _, err := p.manager.Start(ctx, &worlds[i]); err != nil {
			logger.Error().Err(err).Str("worldId", worlds[i].ID).Msg("Failed to restore persistent world")
		}
	}
}

func (p *WorldPlatform) allow(actor, operation string, limit int) bool {
	key := actor + ":" + operation
	p.rateMu.Lock()
	defer p.rateMu.Unlock()
	window := p.rate[key]
	now := time.Now()
	if window.started.IsZero() || now.Sub(window.started) >= time.Minute {
		window = requestWindow{started: now}
	}
	if window.count >= limit {
		return false
	}
	window.count++
	p.rate[key] = window
	return true
}

func osEnv(key, fallback string) string {
	if value := strings.TrimSpace(getenv(key)); value != "" {
		return value
	}
	return fallback
}

var getenv = func(key string) string { return os.Getenv(key) }

func (p *WorldPlatform) register(mux *http.ServeMux, app *App) {
	mux.HandleFunc("/api/csrf", app.csrfHandler)
	mux.HandleFunc("/api/worlds", p.publicWorldsHandler)
	mux.HandleFunc("/api/worlds/", p.worldActionHandler(app))
	mux.HandleFunc("/api/design/worlds", p.designWorldsHandler(app))
	mux.HandleFunc("/api/design/worlds/", p.designWorldHandler(app))
	mux.HandleFunc("/w/", p.worldProxyHandler)
	mux.HandleFunc("/design", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/design/", http.StatusPermanentRedirect)
	})
	mux.Handle("/design/", http.StripPrefix("/design/", spaFileHandler(p.designDir)))
}

func spaFileHandler(directory string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clean := strings.TrimLeft(filepath.Clean(r.URL.Path), `/\\`)
		path := filepath.Join(directory, clean)
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			http.ServeFile(w, r, path)
			return
		}
		http.ServeFile(w, r, filepath.Join(directory, "index.html"))
	})
}

func (app *App) csrfHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", 405)
		return
	}
	if _, ok := getUserIdFromSession(r); !ok {
		http.Error(w, "unauthorized", 401)
		return
	}
	session, err := store.Get(r, "user-session")
	if err != nil {
		http.Error(w, "session error", 500)
		return
	}
	token, _ := session.Values["csrf"].(string)
	if token == "" {
		raw := make([]byte, 32)
		if _, err := rand.Read(raw); err != nil {
			http.Error(w, "token error", 500)
			return
		}
		token = base64.RawURLEncoding.EncodeToString(raw)
		session.Values["csrf"] = token
		if err := session.Save(r, w); err != nil {
			http.Error(w, "session error", 500)
			return
		}
	}
	writePlatformJSON(w, 200, map[string]string{"token": token})
}

func requireCSRF(r *http.Request) bool {
	session, err := store.Get(r, "user-session")
	if err != nil {
		return false
	}
	expected, _ := session.Values["csrf"].(string)
	provided := r.Header.Get("X-CSRF-Token")
	return expected != "" && provided != "" && subtle.ConstantTimeCompare([]byte(expected), []byte(provided)) == 1
}

func actorFor(app *App, r *http.Request) (string, bool, bool) {
	id, ok := getUserIdFromSession(r)
	if !ok {
		return "", false, false
	}
	return id, app.config.isAdminIdentifier(id), true
}
func canEdit(world *WorldDocument, actor string, admin bool) bool {
	return admin || world.OwnerID == actor
}

func (p *WorldPlatform) publicWorldsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", 405)
		return
	}
	worlds, err := p.store.listWorlds(r.Context(), bson.M{"visibility": "public", "moderationState": "active", "publishedReleaseId": bson.M{"$ne": ""}})
	if err != nil {
		platformError(w, err)
		return
	}
	type item struct {
		WorldDocument
		Runtime *RuntimeInfo `json:"runtime,omitempty"`
	}
	out := make([]item, 0, len(worlds))
	for _, world := range worlds {
		entry := item{WorldDocument: world}
		if info, ok := p.manager.Info(world.ID); ok {
			entry.Runtime = &info
		}
		out = append(out, entry)
	}
	writePlatformJSON(w, 200, out)
}

func (p *WorldPlatform) designWorldsHandler(app *App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		actor, admin, ok := actorFor(app, r)
		if !ok {
			http.Error(w, "unauthorized", 401)
			return
		}
		switch r.Method {
		case http.MethodGet:
			filter := bson.M{"ownerId": actor}
			if admin {
				filter = bson.M{}
			}
			worlds, err := p.store.listWorlds(r.Context(), filter)
			if err != nil {
				platformError(w, err)
				return
			}
			writePlatformJSON(w, 200, worlds)
		case http.MethodPost:
			if !p.allow(actor, "create", 10) {
				http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
				return
			}
			if !requireCSRF(r) {
				http.Error(w, "invalid csrf token", 403)
				return
			}
			var request struct {
				Name string `json:"name"`
				Seed string `json:"seed"`
			}
			if err := decodePlatformJSON(w, r, &request, maxResourceBytes); err != nil {
				return
			}
			world, err := p.store.createWorld(r.Context(), actor, request.Name, admin)
			if err != nil {
				platformError(w, err)
				return
			}
			if request.Seed == "" {
				request.Seed = "bloop"
			}
			if err := p.store.seedWorld(r.Context(), world, seedRootFromEnvironment(), request.Seed); err != nil {
				p.store.removeFailedWorld(r.Context(), world.ID)
				platformError(w, err)
				return
			}
			writePlatformJSON(w, 201, world)
		default:
			http.Error(w, "method not allowed", 405)
		}
	}
}

func (p *WorldPlatform) designWorldHandler(app *App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		actor, admin, ok := actorFor(app, r)
		if !ok {
			http.Error(w, "unauthorized", 401)
			return
		}
		parts := splitPlatformPath(strings.TrimPrefix(r.URL.Path, "/api/design/worlds/"))
		if len(parts) < 1 {
			http.NotFound(w, r)
			return
		}
		world, err := p.store.getWorld(r.Context(), parts[0])
		if err != nil {
			platformError(w, err)
			return
		}
		if parts[1] == "report" {
			if !p.allow(actor, "report", 10) {
				http.Error(w, "rate limit exceeded", 429)
				return
			}
			var request struct {
				Reason string `json:"reason"`
			}
			if err := decodePlatformJSON(w, r, &request, 16<<10); err != nil {
				return
			}
			request.Reason = strings.TrimSpace(request.Reason)
			if request.Reason == "" || len(request.Reason) > 1000 {
				http.Error(w, "invalid report reason", 400)
				return
			}
			err := p.store.db.saveAdminAction(AdminActionRecord{ActionType: "world-report", ActingAdmin: actor, TargetIdentifier: world.OwnerID, Payload: bson.M{"worldId": world.ID, "reason": request.Reason}, ResultStatus: "reported", Created: time.Now().UTC()})
			if err != nil {
				platformError(w, err)
				return
			}
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if !canEdit(world, actor, admin) {
			http.Error(w, "forbidden", 403)
			return
		}
		if r.Method != http.MethodGet && !requireCSRF(r) {
			http.Error(w, "invalid csrf token", 403)
			return
		}
		if r.Method != http.MethodGet && !p.allow(actor, "write", 120) {
			http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
			return
		}
		switch {
		case len(parts) == 2 && parts[1] == "draft" && r.Method == http.MethodGet:
			resources, err := p.store.resources(r.Context(), world.ID)
			if err != nil {
				platformError(w, err)
				return
			}
			writePlatformJSON(w, 200, map[string]any{"world": world, "resources": resources})
		case len(parts) == 4 && parts[1] == "resources" && r.Method == http.MethodPut:
			expected, err := parseExpectedRevision(r)
			if err != nil {
				http.Error(w, "invalid If-Match revision", 400)
				return
			}
			body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxResourceBytes+1))
			if err != nil || len(body) > maxResourceBytes {
				http.Error(w, "resource too large", 413)
				return
			}
			saved, err := p.store.putResource(r.Context(), WorldResource{WorldID: world.ID, Kind: parts[2], Key: parts[3], Content: body}, expected)
			if err != nil {
				platformError(w, err)
				return
			}
			w.Header().Set("ETag", fmt.Sprintf("\"%d\"", saved.Revision))
			writePlatformJSON(w, 200, saved)
		case len(parts) == 2 && parts[1] == "releases" && r.Method == http.MethodGet:
			cursor, err := p.store.db.worldReleases.Find(r.Context(), bson.M{"worldId": world.ID}, options.Find().SetSort(bson.D{{Key: "number", Value: -1}}).SetLimit(20))
			if err != nil {
				platformError(w, err)
				return
			}
			defer cursor.Close(r.Context())
			var releases []WorldRelease
			if err := cursor.All(r.Context(), &releases); err != nil {
				platformError(w, err)
				return
			}
			writePlatformJSON(w, 200, releases)
		case len(parts) == 2 && parts[1] == "releases" && r.Method == http.MethodPost:
			p.publishHandler(w, r, world, actor, admin)
		case len(parts) == 3 && parts[1] == "rollback" && r.Method == http.MethodPost:
			if err := p.store.rollback(r.Context(), world.ID, parts[2]); err != nil {
				platformError(w, err)
				return
			}
			_ = p.manager.Stop(r.Context(), world.ID)
			w.WriteHeader(204)
		case len(parts) == 1 && r.Method == http.MethodPatch:
			p.patchWorldHandler(w, r, world, admin)
		default:
			http.NotFound(w, r)
		}
	}
}

func (p *WorldPlatform) publishHandler(w http.ResponseWriter, r *http.Request, world *WorldDocument, actor string, admin bool) {
	if !p.allow(actor, "publish", 10) {
		http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
		return
	}
	select {
	case p.compileSlots <- struct{}{}:
		defer func() { <-p.compileSlots }()
	default:
		http.Error(w, "compiler is busy", 429)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()
	resources, err := p.store.resources(ctx, world.ID)
	if err != nil {
		platformError(w, err)
		return
	}
	total := 0
	for _, resource := range resources {
		total += resource.Size
	}
	if total > maxWorldSourceBytes {
		http.Error(w, "world source exceeds quota", 413)
		return
	}
	compiled, err := compileWorldResources(resources, admin)
	if err != nil {
		http.Error(w, err.Error(), 422)
		return
	}
	release, err := p.store.publish(ctx, world, compiled, actor)
	if err != nil {
		platformError(w, err)
		return
	}
	writePlatformJSON(w, 201, release)
}

func (p *WorldPlatform) patchWorldHandler(w http.ResponseWriter, r *http.Request, world *WorldDocument, admin bool) {
	var request struct {
		Name            *string         `json:"name"`
		Slug            *string         `json:"slug"`
		Lifecycle       *WorldLifecycle `json:"lifecycle"`
		Visibility      *string         `json:"visibility"`
		ModerationState *string         `json:"moderationState"`
	}
	if err := decodePlatformJSON(w, r, &request, 64<<10); err != nil {
		return
	}
	updates := bson.M{"updatedAt": time.Now().UTC()}
	if request.Name != nil {
		value := strings.TrimSpace(*request.Name)
		if value == "" || len(value) > 80 {
			http.Error(w, "invalid name", 400)
			return
		}
		updates["name"] = value
	}
	if request.Slug != nil {
		value := strings.ToLower(strings.TrimSpace(*request.Slug))
		if value != "" && !slugPattern.MatchString(value) {
			http.Error(w, "invalid slug", 400)
			return
		}
		updates["slug"] = value
	}
	if request.Lifecycle != nil {
		if *request.Lifecycle == LifecyclePersistent && !admin {
			http.Error(w, "persistent lifecycle requires admin", 403)
			return
		}
		updates["lifecycle"] = *request.Lifecycle
	}
	if request.Visibility != nil {
		if *request.Visibility != "public" && *request.Visibility != "unlisted" {
			http.Error(w, "invalid visibility", 400)
			return
		}
		updates["visibility"] = *request.Visibility
	}
	if request.ModerationState != nil {
		if !admin {
			http.Error(w, "moderation requires admin", 403)
			return
		}
		if *request.ModerationState != "active" && *request.ModerationState != "suspended" {
			http.Error(w, "invalid moderation state", 400)
			return
		}
		updates["moderationState"] = *request.ModerationState
		if *request.ModerationState == "suspended" {
			_ = p.manager.Stop(r.Context(), world.ID)
		}
	}
	_, err := p.store.db.worlds.UpdateByID(r.Context(), world.ID, bson.M{"$set": updates})
	if err != nil {
		platformError(w, err)
		return
	}
	updated, err := p.store.getWorld(r.Context(), world.ID)
	if err != nil {
		platformError(w, err)
		return
	}
	writePlatformJSON(w, 200, updated)
}

func (p *WorldPlatform) worldActionHandler(app *App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		actor, admin, ok := actorFor(app, r)
		if !ok {
			http.Error(w, "unauthorized", 401)
			return
		}
		if r.Method != http.MethodPost || !requireCSRF(r) {
			http.Error(w, "forbidden", 403)
			return
		}
		parts := splitPlatformPath(strings.TrimPrefix(r.URL.Path, "/api/worlds/"))
		if len(parts) != 2 {
			http.NotFound(w, r)
			return
		}
		world, err := p.store.getWorld(r.Context(), parts[0])
		if err != nil {
			platformError(w, err)
			return
		}
		if !canEdit(world, actor, admin) {
			http.Error(w, "forbidden", 403)
			return
		}
		switch parts[1] {
		case "launch":
			info, err := p.manager.Start(r.Context(), world)
			if err != nil {
				platformError(w, err)
				return
			}
			writePlatformJSON(w, 200, info)
		case "stop":
			if err := p.manager.Stop(r.Context(), world.ID); err != nil {
				platformError(w, err)
				return
			}
			w.WriteHeader(204)
		default:
			http.NotFound(w, r)
		}
	}
}

func (p *WorldPlatform) worldProxyHandler(w http.ResponseWriter, r *http.Request) {
	parts := splitPlatformPath(strings.TrimPrefix(r.URL.Path, "/w/"))
	if len(parts) < 1 {
		http.NotFound(w, r)
		return
	}
	world, err := p.store.getWorld(r.Context(), parts[0])
	if err != nil {
		platformError(w, err)
		return
	}
	if world.ModerationState != "active" {
		http.Error(w, "world unavailable", 503)
		return
	}
	if _, ok := p.manager.Info(world.ID); !ok {
		http.Error(w, "world is not running", 503)
		return
	}
	original := "/w/" + parts[0]
	r.URL.Path = "/w/" + world.ID + strings.TrimPrefix(r.URL.Path, original)
	if !p.manager.Proxy(world.ID, w, r) {
		http.Error(w, "world unavailable", 503)
	}
}

func splitPlatformPath(path string) []string {
	raw := strings.Split(strings.Trim(path, "/"), "/")
	out := raw[:0]
	for _, part := range raw {
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}
func decodePlatformJSON(w http.ResponseWriter, r *http.Request, target any, limit int64) error {
	r.Body = http.MaxBytesReader(w, r.Body, limit)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		http.Error(w, "invalid JSON: "+err.Error(), 400)
		return err
	}
	return nil
}
func writePlatformJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
func platformError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, mongo.ErrNoDocuments):
		http.Error(w, "not found", 404)
	case errors.Is(err, ErrRevisionConflict):
		http.Error(w, err.Error(), 409)
	case mongo.IsDuplicateKeyError(err):
		http.Error(w, "conflict", 409)
	default:
		http.Error(w, err.Error(), 500)
	}
}
