package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/sessions"
	"go.mongodb.org/mongo-driver/bson"
)

type retryTestLauncher struct {
	fakeRuntimeLauncher
	failures atomic.Int64
}

func TestWorldAuthoringStaleLifecycleCannotStopReplacement(t *testing.T) {
	manager := newRuntimeManager(context.Background(), nil, nil, "")
	old := &ManagedRuntime{RuntimeInfo: RuntimeInfo{WorldID: "world", ReleaseID: "old"}}
	replacement := &ManagedRuntime{RuntimeInfo: RuntimeInfo{WorldID: "world", ReleaseID: "new"}}
	manager.runtimes["world"] = replacement
	if err := manager.stopInstance(context.Background(), "world", old); err != nil {
		t.Fatal(err)
	}
	if info, ok := manager.Info("world"); !ok || info.ReleaseID != "new" {
		t.Fatal("stale lifecycle check stopped the replacement")
	}
}

func (launcher *retryTestLauncher) Start(ctx context.Context, world *WorldDocument, release *WorldRelease, directory string) (*ManagedRuntime, error) {
	if launcher.failures.Load() > 0 {
		launcher.failures.Add(-1)
		return nil, errors.New("temporary startup failure")
	}
	return launcher.fakeRuntimeLauncher.Start(ctx, world, release, directory)
}

func authoringSession(t *testing.T, actor string) *http.Cookie {
	t.Helper()
	request := httptest.NewRequest("GET", "/", nil)
	session, err := store.Get(request, "user-session")
	if err != nil {
		t.Fatal(err)
	}
	session.Values["identifier"], session.Values["csrf"] = actor, "test-csrf"
	response := httptest.NewRecorder()
	if err := session.Save(request, response); err != nil {
		t.Fatal(err)
	}
	return response.Result().Cookies()[0]
}

func TestWorldAuthoringRequiresAdmin(t *testing.T) {
	previous := store
	store = sessions.NewCookieStore([]byte("world-editor-test-key"))
	defer func() { store = previous }()
	app := &App{config: &Configuration{mode: "controller", adminIdentifiers: []string{"admin"}}}
	platform := &WorldPlatform{}
	mux := http.NewServeMux()
	platform.register(mux, app)
	for _, route := range []string{"/api/design/worlds", "/api/design/worlds/world/draft", "/api/design/worlds/world", "/api/worlds/world/launch", "/api/worlds/world/stop", "/design/", "/admin/worlds", "/admin"} {
		for _, actor := range []string{"", "owner"} {
			request := httptest.NewRequest("GET", route, nil)
			if actor != "" {
				request.AddCookie(authoringSession(t, actor))
			}
			response := httptest.NewRecorder()
			mux.ServeHTTP(response, request)
			want := 401
			if actor != "" {
				want = 403
			}
			if response.Code != want {
				t.Fatalf("%s %s: got %d want %d", actor, route, response.Code, want)
			}
		}
	}
	if canEdit(&WorldDocument{OwnerID: "owner"}, "owner", false) {
		t.Fatal("ownership bypassed admin restriction")
	}
}

func TestWorldAuthoringVersionsRestoreAndDeployment(t *testing.T) {
	testdb()
	database := testClient.Database("bloop-authoring-test-" + uuid.NewString())
	t.Cleanup(func() { _ = database.Drop(context.Background()) })
	db := &DB{database: database, worlds: database.Collection("worlds"), worldResources: database.Collection("worldResources"), worldReleases: database.Collection("worldReleases"), runtimeInstances: database.Collection("runtimeInstances"), users: database.Collection("users"), playerRecords: database.Collection("players"), events: database.Collection("events"), sessionData: database.Collection("sessions")}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	worldStore, err := newWorldStore(db)
	if err != nil {
		t.Fatal(err)
	}
	if err := worldStore.ensureIndexes(ctx); err != nil {
		t.Fatal(err)
	}
	world, err := worldStore.createWorld(ctx, "admin", "Test", true)
	if err != nil {
		t.Fatal(err)
	}
	for _, seed := range []string{"bloop", "escape"} {
		t.Run("seed-"+seed, func(t *testing.T) {
			seeded, err := worldStore.createWorld(ctx, "admin", seed, true)
			if err != nil {
				t.Fatal(err)
			}
			if err := worldStore.seedWorld(ctx, seeded, seedRootFromEnvironment(), seed); err != nil {
				t.Fatal(err)
			}
			resources, err := worldStore.resources(ctx, seeded.ID)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := compileWorldResources(resources, true); err != nil {
				t.Fatal(err)
			}
		})
	}
	for _, resource := range testReleaseResources(t) {
		resource.WorldID = world.ID
		if _, err := worldStore.putResource(ctx, resource, 0); err != nil {
			t.Fatal(err)
		}
	}
	world, _ = worldStore.getWorld(ctx, world.ID)
	resources, err := worldStore.resources(ctx, world.ID)
	if err != nil {
		t.Fatal(err)
	}
	compiled, err := compileWorldResources(resources, true)
	if err != nil {
		t.Fatal(err)
	}
	v1, err := worldStore.publish(ctx, world, compiled, "admin", "First", true)
	if err != nil {
		t.Fatal(err)
	}
	var manifest WorldManifest
	for _, resource := range resources {
		if resource.Kind == "manifest" {
			_ = json.Unmarshal(resource.Content, &manifest)
		}
	}
	manifest.Name = "WIP"
	if _, err := worldStore.putResource(ctx, WorldResource{WorldID: world.ID, Kind: "manifest", Key: "world", Content: mustJSON(t, manifest)}, 1); err != nil {
		t.Fatal(err)
	}
	world, _ = worldStore.getWorld(ctx, world.ID)
	resources, _ = worldStore.resources(ctx, world.ID)
	compiled, err = compileWorldResources(resources, true)
	if err != nil {
		t.Fatal(err)
	}
	v2, err := worldStore.publish(ctx, world, compiled, "admin", "WIP", false)
	if err != nil {
		t.Fatal(err)
	}
	world, _ = worldStore.getWorld(ctx, world.ID)
	if world.PublishedReleaseID != v1.ID {
		t.Fatal("saving WIP selected it for launch")
	}
	if err := worldStore.restoreDraft(ctx, world, v1.ID); err != nil {
		t.Fatal(err)
	}
	if err := worldStore.restoreDraft(ctx, world, v2.ID); !errors.Is(err, ErrRevisionConflict) {
		t.Fatalf("stale restore: %v", err)
	}
	if _, err := worldStore.putResource(ctx, WorldResource{WorldID: world.ID, Kind: "manifest", Key: "world", Content: mustJSON(t, manifest)}, 2); !errors.Is(err, ErrRevisionConflict) {
		t.Fatalf("stale resource: %v", err)
	}
	resources, err = worldStore.resources(ctx, world.ID)
	if err != nil {
		t.Fatal(err)
	}
	for _, resource := range resources {
		if resource.Kind == "manifest" {
			_ = json.Unmarshal(resource.Content, &manifest)
		}
	}
	if manifest.Name != "Test" {
		t.Fatal("restored draft did not match snapshot")
	}

	previous := store
	store = sessions.NewCookieStore([]byte("world-editor-test-key"))
	defer func() { store = previous }()
	launcher := &fakeRuntimeLauncher{}
	manager := newRuntimeManager(context.Background(), worldStore, launcher, t.TempDir())
	t.Cleanup(func() { manager.Shutdown(context.Background()) })
	platform := &WorldPlatform{store: worldStore, manager: manager, compileSlots: make(chan struct{}, 2), rate: map[string]requestWindow{}}
	app := &App{config: &Configuration{mode: "controller", adminIdentifiers: []string{"admin"}}}
	mux := http.NewServeMux()
	platform.register(mux, app)
	cookie := authoringSession(t, "admin")
	request := func(method, path, body string, csrf bool) *httptest.ResponseRecorder {
		r := httptest.NewRequest(method, path, strings.NewReader(body))
		r.AddCookie(cookie)
		if csrf {
			r.Header.Set("X-CSRF-Token", "test-csrf")
		}
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)
		return w
	}
	base := "/api/design/worlds/" + world.ID
	if response := request("PATCH", base, `{"lifecycle":"persistent"}`, false); response.Code != 403 {
		t.Fatal("write accepted without CSRF")
	}
	if response := request("PATCH", base, `{"lifecycle":"invalid"}`, true); response.Code != 400 {
		t.Fatal("invalid lifecycle accepted")
	}
	if response := request("PATCH", base, `{"lifecycle":"persistent"}`, true); response.Code != 200 {
		t.Fatal(response.Body.String())
	}
	if response := request("POST", "/api/worlds/"+world.ID+"/launch", `{}`, true); response.Code != 200 {
		t.Fatal(response.Body.String())
	}
	if response := request("POST", base+"/rollback/"+v2.ID, `{}`, true); response.Code != 204 {
		t.Fatal(response.Body.String())
	}
	if info, _ := manager.Info(world.ID); info.ReleaseID != v1.ID {
		t.Fatal("selecting a version changed the live world")
	}
	world, _ = worldStore.getWorld(ctx, world.ID)
	if !world.DeploymentEnabled || world.DeployedReleaseID != v1.ID || world.PublishedReleaseID != v2.ID {
		t.Fatal("deployed and selected versions were not kept separate")
	}
	retryingLauncher := &retryTestLauncher{}
	second := newRuntimeManager(context.Background(), worldStore, retryingLauncher, t.TempDir())
	t.Cleanup(func() { second.Shutdown(context.Background()) })
	(&WorldPlatform{store: worldStore, manager: second}).restorePersistentWorlds(ctx)
	if info, _ := second.Info(world.ID); info.ReleaseID != v1.ID {
		t.Fatal("restart did not restore the deployed version")
	}
	second.mu.RLock()
	crashed := second.runtimes[world.ID]
	second.mu.RUnlock()
	retryingLauncher.failures.Store(1)
	crashed.done <- errors.New("simulated runtime crash")
	deadline := time.Now().Add(9 * time.Second)
	recovered := false
	for time.Now().Before(deadline) {
		second.mu.RLock()
		current := second.runtimes[world.ID]
		recovered = current != nil && current != crashed && current.ReleaseID == v1.ID
		second.mu.RUnlock()
		if recovered {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	if !recovered {
		t.Fatal("persistent world did not recover after a failed restart")
	}
	if response := request("POST", "/api/worlds/"+world.ID+"/stop", `{}`, true); response.Code != 204 {
		t.Fatal(response.Body.String())
	}
	world, _ = worldStore.getWorld(ctx, world.ID)
	if world.DeploymentEnabled {
		t.Fatal("shutdown did not persist")
	}
	if _, err := manager.Start(ctx, &WorldDocument{ID: world.ID, DeploymentEnabled: true}); err == nil {
		t.Fatal("stale launch intent bypassed persistent shutdown")
	}
	count, err := db.worlds.CountDocuments(ctx, bson.M{"lifecycle": LifecyclePersistent, "deploymentEnabled": bson.M{"$ne": false}})
	if err != nil || count != 0 {
		t.Fatal("stopped world would restart with controller")
	}
	if response := request("GET", base, "", false); response.Code != 200 {
		t.Fatal("world metadata endpoint failed")
	}
}
