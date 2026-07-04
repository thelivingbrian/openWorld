package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type RuntimeState string

const (
	RuntimeStarting RuntimeState = "starting"
	RuntimeRunning  RuntimeState = "running"
	RuntimeStopping RuntimeState = "stopping"
	RuntimeFailed   RuntimeState = "failed"
)

type RuntimeInfo struct {
	WorldID      string       `json:"worldId"`
	ReleaseID    string       `json:"releaseId"`
	State        RuntimeState `json:"state"`
	StartedAt    time.Time    `json:"startedAt"`
	PlayerCount  int          `json:"playerCount"`
	OwnerPresent bool         `json:"ownerPresent"`
	Error        string       `json:"error,omitempty"`
}

type ManagedRuntime struct {
	RuntimeInfo
	cmd      *exec.Cmd
	proxy    *httputil.ReverseProxy
	cancel   context.CancelFunc
	done     chan error
	basePath string
	ownerID  string
}

type RuntimeLauncher interface {
	Start(context.Context, *WorldDocument, *WorldRelease, string) (*ManagedRuntime, error)
	Stop(context.Context, *ManagedRuntime) error
}

type ProcessLauncher struct {
	Executable string
	Config     *Configuration
}

func (l *ProcessLauncher) Start(parent context.Context, world *WorldDocument, release *WorldRelease, contentDir string) (*ManagedRuntime, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}
	address := listener.Addr().String()
	_ = listener.Close()
	ctx, cancel := context.WithCancel(parent)
	cmd := exec.CommandContext(ctx, l.Executable)
	cmd.Dir, _ = os.Getwd()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	port := strings.TrimPrefix(address, "127.0.0.1")
	cmd.Env = append(os.Environ(), "OPENWORLD_MODE=runtime", "IS_HUB=FALSE", "WORLD_PLATFORM_ENABLED=FALSE", "USE_TLS=FALSE", "PPROF_ENABLED=FALSE", "WORLD_ID="+world.ID, "WORLD_OWNER_ID="+world.OwnerID, "WORLD_LIFECYCLE="+string(world.Lifecycle), "WORLD_BASE_PATH=/w/"+world.ID, "WORLD_CONTENT_DIR="+contentDir, "SERVER_NAME="+world.Name, "DOMAIN_NAME=/w/"+world.ID, "BLOOP_PORT="+port)
	if limit := strings.TrimSpace(os.Getenv("WORLD_RUNTIME_MEMORY_LIMIT")); limit != "" {
		cmd.Env = append(cmd.Env, "GOMEMLIMIT="+limit)
	}
	if procs := strings.TrimSpace(os.Getenv("WORLD_RUNTIME_GOMAXPROCS")); procs != "" {
		cmd.Env = append(cmd.Env, "GOMAXPROCS="+procs)
	}
	if err := cmd.Start(); err != nil {
		cancel()
		return nil, err
	}
	target, _ := url.Parse("http://" + address)
	basePath := "/w/" + world.ID
	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.ModifyResponse = prefixRuntimeResponse(basePath)
	runtime := &ManagedRuntime{RuntimeInfo: RuntimeInfo{WorldID: world.ID, ReleaseID: release.ID, State: RuntimeStarting, StartedAt: time.Now().UTC()}, cmd: cmd, proxy: proxy, cancel: cancel, done: make(chan error, 1), basePath: basePath, ownerID: world.OwnerID}
	go func() { runtime.done <- cmd.Wait(); close(runtime.done) }()
	deadline := time.Now().Add(10 * time.Second)
	client := http.Client{Timeout: 500 * time.Millisecond}
	for time.Now().Before(deadline) {
		select {
		case err := <-runtime.done:
			cancel()
			return nil, fmt.Errorf("runtime exited during startup: %w", err)
		default:
		}
		response, requestErr := client.Get(target.String() + "/healthz")
		if requestErr == nil {
			_ = response.Body.Close()
			if response.StatusCode == http.StatusOK {
				runtime.State = RuntimeRunning
				return runtime, nil
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
	cancel()
	return nil, errors.New("runtime readiness timeout")
}

func prefixRuntimeResponse(basePath string) func(*http.Response) error {
	return func(response *http.Response) error {
		contentType := response.Header.Get("Content-Type")
		if !strings.Contains(contentType, "text/html") && !strings.Contains(contentType, "javascript") && !strings.Contains(contentType, "text/css") {
			return nil
		}
		body, err := io.ReadAll(response.Body)
		if err != nil {
			return err
		}
		_ = response.Body.Close()
		attributes := []string{`href="`, `src="`, `hx-get="`, `hx-post="`, `ws-connect="`, `action="`}
		runtimePaths := []string{"/play", "/screen", "/admin", "/images", "/assets", "/stats", "/insert", "/status"}
		for _, attribute := range attributes {
			for _, path := range runtimePaths {
				body = bytes.ReplaceAll(body, []byte(attribute+path), []byte(attribute+basePath+path))
			}
		}
		response.Body = io.NopCloser(bytes.NewReader(body))
		response.ContentLength = int64(len(body))
		response.Header.Set("Content-Length", strconv.Itoa(len(body)))
		return nil
	}
}

func (l *ProcessLauncher) Stop(ctx context.Context, runtime *ManagedRuntime) error {
	runtime.State = RuntimeStopping
	if runtime.cmd != nil && runtime.cmd.Process != nil {
		_ = runtime.cmd.Process.Signal(os.Interrupt)
	}
	select {
	case <-runtime.done:
		runtime.cancel()
		return nil
	case <-ctx.Done():
		runtime.cancel()
		return ctx.Err()
	case <-time.After(5 * time.Second):
		runtime.cancel()
		return errors.New("runtime forced closed after shutdown timeout")
	}
}

type RuntimeManager struct {
	mu       sync.RWMutex
	store    *WorldStore
	launcher RuntimeLauncher
	cacheDir string
	runtimes map[string]*ManagedRuntime
	starts   map[string]chan struct{}
	context  context.Context
}

func newRuntimeManager(ctx context.Context, store *WorldStore, launcher RuntimeLauncher, cacheDir string) *RuntimeManager {
	return &RuntimeManager{store: store, launcher: launcher, cacheDir: cacheDir, runtimes: map[string]*ManagedRuntime{}, starts: map[string]chan struct{}{}, context: ctx}
}

func (m *RuntimeManager) Start(ctx context.Context, world *WorldDocument) (*RuntimeInfo, error) {
	for {
		m.mu.Lock()
		if runtime := m.runtimes[world.ID]; runtime != nil {
			if runtime.ReleaseID != world.PublishedReleaseID {
				m.mu.Unlock()
				if err := m.Stop(ctx, world.ID); err != nil {
					return nil, err
				}
				continue
			}
			info := runtime.RuntimeInfo
			m.mu.Unlock()
			return &info, nil
		}
		if waiting := m.starts[world.ID]; waiting != nil {
			m.mu.Unlock()
			select {
			case <-waiting:
				continue
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}
		for _, running := range m.runtimes {
			if running.ownerID == world.OwnerID && running.WorldID != world.ID && world.Lifecycle != LifecyclePersistent {
				m.mu.Unlock()
				return nil, errors.New("owner already has a running world")
			}
		}
		waiting := make(chan struct{})
		m.starts[world.ID] = waiting
		m.mu.Unlock()
		break
	}
	defer func() { m.mu.Lock(); close(m.starts[world.ID]); delete(m.starts, world.ID); m.mu.Unlock() }()
	if world.ModerationState != "active" {
		return nil, errors.New("world is suspended")
	}
	if world.PublishedReleaseID == "" {
		return nil, errors.New("world has no published release")
	}
	release, err := m.store.release(ctx, world.ID, world.PublishedReleaseID)
	if err != nil {
		return nil, err
	}
	directory, err := m.materialize(ctx, release)
	if err != nil {
		return nil, err
	}
	_, _ = m.store.db.runtimeInstances.UpdateByID(ctx, world.ID, bson.M{"$set": bson.M{"draining": false}}, options.Update().SetUpsert(true))
	runtime, err := m.launcher.Start(m.context, world, release, directory)
	if err != nil {
		return nil, err
	}
	m.mu.Lock()
	m.runtimes[world.ID] = runtime
	m.mu.Unlock()
	go m.observe(world, runtime)
	go m.monitorLifecycle(world, runtime)
	info := runtime.RuntimeInfo
	return &info, nil
}

func (m *RuntimeManager) monitorLifecycle(world *WorldDocument, runtime *ManagedRuntime) {
	if world.Lifecycle == LifecyclePersistent {
		return
	}
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	lastOwner := time.Now()
	ownerSeen := false
	for range ticker.C {
		m.mu.RLock()
		active := m.runtimes[world.ID] == runtime
		m.mu.RUnlock()
		if !active {
			return
		}
		var heartbeat struct {
			PlayerCount  int  `bson:"playerCount"`
			OwnerPresent bool `bson:"ownerPresent"`
		}
		err := m.store.db.runtimeInstances.FindOne(context.Background(), bson.M{"_id": world.ID}).Decode(&heartbeat)
		if err != nil {
			continue
		}
		m.mu.Lock()
		runtime.PlayerCount = heartbeat.PlayerCount
		runtime.OwnerPresent = heartbeat.OwnerPresent
		m.mu.Unlock()
		if heartbeat.OwnerPresent {
			lastOwner = time.Now()
			ownerSeen = true
			if world.Lifecycle == LifecycleUntilEmpty {
				_, _ = m.store.db.runtimeInstances.UpdateByID(context.Background(), world.ID, bson.M{"$set": bson.M{"draining": false}})
			}
			continue
		}
		if world.Lifecycle == LifecycleUntilEmpty && ownerSeen {
			_, _ = m.store.db.runtimeInstances.UpdateByID(context.Background(), world.ID, bson.M{"$set": bson.M{"draining": true}})
		}
		shouldStop := world.Lifecycle == LifecycleOwnerPresent && time.Since(lastOwner) >= 60*time.Second
		if world.Lifecycle == LifecycleUntilEmpty {
			shouldStop = ownerSeen && heartbeat.PlayerCount == 0
		}
		if shouldStop {
			ctx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
			_ = m.Stop(ctx, world.ID)
			cancel()
			return
		}
	}
}

func (m *RuntimeManager) materialize(ctx context.Context, release *WorldRelease) (string, error) {
	directory := filepath.Join(m.cacheDir, release.ArtifactHash)
	if info, err := os.Stat(filepath.Join(directory, "areas.json")); err == nil && !info.IsDir() {
		return directory, nil
	}
	archive, err := m.store.downloadArtifact(ctx, release)
	if err != nil {
		return "", err
	}
	if hashBytes(archive) != release.ArtifactHash {
		return "", errors.New("release artifact hash mismatch")
	}
	if err := os.MkdirAll(m.cacheDir, 0755); err != nil {
		return "", err
	}
	temporary, err := os.MkdirTemp(m.cacheDir, "extract-")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(temporary)
	if err := extractReleaseArchive(archive, temporary); err != nil {
		return "", err
	}
	if err := os.Rename(temporary, directory); err != nil {
		if _, statErr := os.Stat(directory); statErr != nil {
			return "", err
		}
	}
	return directory, nil
}

func (m *RuntimeManager) observe(world *WorldDocument, runtime *ManagedRuntime) {
	err := <-runtime.done
	m.mu.Lock()
	stopping := runtime.State == RuntimeStopping
	if m.runtimes[world.ID] == runtime {
		delete(m.runtimes, world.ID)
	}
	runtime.State = RuntimeFailed
	if err != nil {
		runtime.Error = err.Error()
	}
	m.mu.Unlock()
	if world.Lifecycle == LifecyclePersistent && !stopping {
		time.Sleep(2 * time.Second)
		if _, startErr := m.Start(context.Background(), world); startErr != nil {
			logger.Error().Err(startErr).Str("worldId", world.ID).Msg("Failed to restart persistent runtime")
		}
	}
}

func (m *RuntimeManager) Stop(ctx context.Context, worldID string) error {
	m.mu.Lock()
	runtime := m.runtimes[worldID]
	if runtime != nil {
		runtime.State = RuntimeStopping
		delete(m.runtimes, worldID)
	}
	m.mu.Unlock()
	if runtime == nil {
		return nil
	}
	return m.launcher.Stop(ctx, runtime)
}

func (m *RuntimeManager) Shutdown(ctx context.Context) {
	m.mu.RLock()
	ids := make([]string, 0, len(m.runtimes))
	for id := range m.runtimes {
		ids = append(ids, id)
	}
	m.mu.RUnlock()
	for _, id := range ids {
		_ = m.Stop(ctx, id)
	}
}

func (m *RuntimeManager) Info(worldID string) (RuntimeInfo, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	runtime := m.runtimes[worldID]
	if runtime == nil {
		return RuntimeInfo{}, false
	}
	return runtime.RuntimeInfo, true
}

func (m *RuntimeManager) Proxy(worldID string, w http.ResponseWriter, r *http.Request) bool {
	m.mu.RLock()
	runtime := m.runtimes[worldID]
	m.mu.RUnlock()
	if runtime == nil {
		return false
	}
	prefix := "/w/" + worldID
	r.URL.Path = strings.TrimPrefix(r.URL.Path, prefix)
	if r.URL.Path == "" {
		r.URL.Path = "/"
	}
	r.Header.Set("X-World-Base-Path", prefix)
	runtime.proxy.ServeHTTP(w, r)
	return true
}

func currentExecutable() (string, error) {
	path, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.Abs(path)
}
func parseExpectedRevision(r *http.Request) (int64, error) {
	value := strings.Trim(strings.TrimSpace(r.Header.Get("If-Match")), "\"")
	if value == "" {
		return 0, nil
	}
	return strconv.ParseInt(value, 10, 64)
}
