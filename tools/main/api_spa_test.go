package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

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
