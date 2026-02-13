package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type bootstrapResponse struct {
	Collections map[string]*Collection `json:"collections"`
	Colors      []Color                `json:"colors"`
}

type saveSpaceRequest struct {
	CollectionName string `json:"collectionName"`
	SpaceName      string `json:"spaceName"`
	Space          Space  `json:"space"`
}

type saveFragmentSetRequest struct {
	CollectionName string     `json:"collectionName"`
	SetName        string     `json:"setName"`
	Fragments      []Fragment `json:"fragments"`
}

type savePrototypeSetRequest struct {
	CollectionName string      `json:"collectionName"`
	SetName        string      `json:"setName"`
	Prototypes     []Prototype `json:"prototypes"`
}

type saveInteractableSetRequest struct {
	CollectionName string                    `json:"collectionName"`
	SetName        string                    `json:"setName"`
	Interactables  []InteractableDescription `json:"interactables"`
}

type collectionCommandRequest struct {
	CollectionName string `json:"collectionName"`
}

func decodeJSONBody[T any](r *http.Request) (T, error) {
	var body T
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&body); err != nil {
		return body, err
	}
	return body, nil
}

func encodeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		fmt.Println("failed to encode json:", err)
	}
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	encodeJSON(w, status, map[string]string{"error": message})
}

func (c *Context) apiBootstrapHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	encodeJSON(w, http.StatusOK, bootstrapResponse{
		Collections: c.Collections,
		Colors:      c.colors,
	})
}

func (c *Context) apiSaveSpaceHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	req, err := decodeJSONBody[saveSpaceRequest](r)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	col := c.Collections[req.CollectionName]
	if col == nil {
		writeJSONError(w, http.StatusNotFound, "collection not found")
		return
	}
	req.Space.CollectionName = req.CollectionName
	col.Spaces[req.SpaceName] = &req.Space
	col.saveSpace(req.SpaceName)

	encodeJSON(w, http.StatusOK, map[string]string{"status": "saved"})
}

func (c *Context) apiSaveFragmentSetHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	req, err := decodeJSONBody[saveFragmentSetRequest](r)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	col := c.Collections[req.CollectionName]
	if col == nil {
		writeJSONError(w, http.StatusNotFound, "collection not found")
		return
	}

	col.Fragments[req.SetName] = req.Fragments
	col.saveFragmentSet(req.SetName)
	encodeJSON(w, http.StatusOK, map[string]string{"status": "saved"})
}

func (c *Context) apiSavePrototypeSetHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	req, err := decodeJSONBody[savePrototypeSetRequest](r)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	col := c.Collections[req.CollectionName]
	if col == nil {
		writeJSONError(w, http.StatusNotFound, "collection not found")
		return
	}

	col.PrototypeSets[req.SetName] = req.Prototypes
	col.savePrototypeSet(req.SetName)
	encodeJSON(w, http.StatusOK, map[string]string{"status": "saved"})
}

func (c *Context) apiSaveInteractableSetHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	req, err := decodeJSONBody[saveInteractableSetRequest](r)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	col := c.Collections[req.CollectionName]
	if col == nil {
		writeJSONError(w, http.StatusNotFound, "collection not found")
		return
	}

	col.InteractableSets[req.SetName] = req.Interactables
	col.saveInteractableSet(req.SetName)
	encodeJSON(w, http.StatusOK, map[string]string{"status": "saved"})
}

func (c *Context) apiColorsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	req, err := decodeJSONBody[[]Color](r)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	c.colors = req
	if err := c.writeColorsToLocalFile(); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to save colors")
		return
	}
	c.createLocalCSSFile()

	encodeJSON(w, http.StatusOK, map[string]string{"status": "saved"})
}

func (c *Context) apiCompileHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	req, err := decodeJSONBody[collectionCommandRequest](r)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	c.createCSSFile(CSS_PATH)
	c.compileCollectionByName(req.CollectionName)
	encodeJSON(w, http.StatusOK, map[string]string{"status": "compiled"})
}

func (c *Context) apiDeployHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	req, err := decodeJSONBody[collectionCommandRequest](r)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	c.deploy(req.CollectionName)
	encodeJSON(w, http.StatusOK, map[string]string{"status": "deployed"})
}

func (c *Context) spaHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.NotFound(w, r)
		return
	}

	const distRoot = "./spa/dist/spa/browser"
	rel := strings.TrimPrefix(r.URL.Path, "/app")
	if rel == "" || rel == "/" {
		http.ServeFile(w, r, filepath.Join(distRoot, "index.html"))
		return
	}

	rel = strings.TrimPrefix(rel, "/")
	target := filepath.Join(distRoot, filepath.Clean(rel))
	if info, err := os.Stat(target); err == nil && !info.IsDir() {
		http.ServeFile(w, r, target)
		return
	}

	http.ServeFile(w, r, filepath.Join(distRoot, "index.html"))
}
