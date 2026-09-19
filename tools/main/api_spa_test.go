package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestManifestSaveReloadAndCompile(t *testing.T) {
	previous, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(t.TempDir()); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(previous) })
	if err := os.MkdirAll("data/collections/test", 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(COMPILE_basePath, 0755); err != nil {
		t.Fatal(err)
	}
	collection := &Collection{Name: "test", Spaces: map[string]*Space{}}
	context := Context{Collections: map[string]*Collection{"test": collection}}
	body := `{"collectionName":"test","manifest":{"name":"Test","npcs":[{"id":"guard","program":[{"action":"wander","ticks":2}]}],"spawns":[],"achievements":[{"id":"explore","metric":"visit-stage","stage":"room"}]}}`
	response := httptest.NewRecorder()
	context.apiManifestHandler(response, httptest.NewRequest(http.MethodPut, "/api/manifest", strings.NewReader(body)))
	if response.Code != http.StatusOK {
		t.Fatal(response.Body.String())
	}
	loaded := context.getAllCollections(COLLECTION_PATH)["test"]
	if !json.Valid(loaded.Manifest) {
		t.Fatal("saved manifest was not reloaded")
	}
	context.compileCollection(loaded)
	data, err := os.ReadFile(filepath.Join(COMPILE_basePath, "manifest.json"))
	if err != nil || !strings.Contains(string(data), `"guard"`) || !strings.Contains(string(data), `"explore"`) {
		t.Fatal("compiled manifest lost NPC or achievement definitions", err)
	}
}

func TestServeSPARedirectsToolsRootToDesign(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/?world=test-world", nil)
	response := httptest.NewRecorder()

	serveSPA(t.TempDir(), response, req)

	if response.Code != http.StatusTemporaryRedirect {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusTemporaryRedirect)
	}
	if location := response.Header().Get("Location"); location != "/design/?world=test-world" {
		t.Fatalf("Location = %q, want %q", location, "/design/?world=test-world")
	}
}

func TestServeSPARedirectsDesignToTrailingSlash(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/design", nil)
	response := httptest.NewRecorder()

	serveSPA(t.TempDir(), response, req)

	if response.Code != http.StatusTemporaryRedirect {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusTemporaryRedirect)
	}
	if location := response.Header().Get("Location"); location != "/design/" {
		t.Fatalf("Location = %q, want %q", location, "/design/")
	}
}

func TestServeSPAServesAssetsBelowDesign(t *testing.T) {
	distRoot := t.TempDir()
	writeSPAFile(t, distRoot, "main.js", "console.log('workspace')")
	req := httptest.NewRequest(http.MethodGet, "/design/main.js", nil)
	response := httptest.NewRecorder()

	serveSPA(distRoot, response, req)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	if body := response.Body.String(); body != "console.log('workspace')" {
		t.Fatalf("body = %q", body)
	}
}

func TestServeSPAFallsBackToIndexBelowDesign(t *testing.T) {
	distRoot := t.TempDir()
	writeSPAFile(t, distRoot, "index.html", "<app-root></app-root>")
	req := httptest.NewRequest(http.MethodGet, "/design/worlds/example", nil)
	response := httptest.NewRecorder()

	serveSPA(distRoot, response, req)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	if body := response.Body.String(); body != "<app-root></app-root>" {
		t.Fatalf("body = %q", body)
	}
}

func writeSPAFile(t *testing.T, root, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(root, name), []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}
