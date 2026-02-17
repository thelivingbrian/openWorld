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

type createCollectionRequest struct {
	Name string `json:"name"`
}

type createSpaceRequest struct {
	CollectionName string `json:"collectionName"`
	Name           string `json:"name"`
	Topology       string `json:"topology"`
	Latitude       int    `json:"latitude"`
	Longitude      int    `json:"longitude"`
	AreaWidth      int    `json:"areaWidth"`
	AreaHeight     int    `json:"areaHeight"`
	TileColor      string `json:"tileColor"`
	TileColor1     string `json:"tileColor1"`
	Weather        string `json:"weather"`
	BroadcastGroup string `json:"broadcastGroup"`
}

type createAreaRequest struct {
	CollectionName    string `json:"collectionName"`
	SpaceName         string `json:"spaceName"`
	Name              string `json:"name"`
	Safe              bool   `json:"safe"`
	Height            int    `json:"height"`
	Width             int    `json:"width"`
	DefaultTileColor  string `json:"defaultTileColor"`
	DefaultTileColor1 string `json:"defaultTileColor1"`
}

type flattenSpaceRequest struct {
	CollectionName string `json:"collectionName"`
	SpaceName      string `json:"spaceName"`
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

func (c *Context) apiCollectionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	req, err := decodeJSONBody[createCollectionRequest](r)
	if err != nil || strings.TrimSpace(req.Name) == "" {
		writeJSONError(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	name := strings.ToLower(strings.TrimSpace(req.Name))
	if c.Collections[name] != nil {
		writeJSONError(w, http.StatusConflict, "collection already exists")
		return
	}

	c.Collections[name] = &Collection{
		Name:             name,
		Spaces:           make(map[string]*Space),
		Fragments:        make(map[string][]Fragment),
		PrototypeSets:    make(map[string][]Prototype),
		InteractableSets: make(map[string][]InteractableDescription),
		StructureSets:    make(map[string][]Structure),
	}
	createCollectionDirectories(name)

	encodeJSON(w, http.StatusCreated, map[string]string{"status": "created", "collectionName": name})
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

func (c *Context) apiCreateSpaceHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	req, err := decodeJSONBody[createSpaceRequest](r)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	col := c.Collections[req.CollectionName]
	if col == nil {
		writeJSONError(w, http.StatusNotFound, "collection not found")
		return
	}
	if strings.TrimSpace(req.Name) == "" {
		writeJSONError(w, http.StatusBadRequest, "space name required")
		return
	}
	if req.Latitude <= 0 || req.Longitude <= 0 || req.AreaWidth <= 0 || req.AreaHeight <= 0 {
		writeJSONError(w, http.StatusBadRequest, "invalid dimensions")
		return
	}
	if col.Spaces[req.Name] != nil {
		writeJSONError(w, http.StatusConflict, "space already exists")
		return
	}

	space := createSpace(
		req.CollectionName,
		req.Name,
		req.Latitude,
		req.Longitude,
		req.Topology,
		req.AreaHeight,
		req.AreaWidth,
		req.TileColor,
		req.TileColor1,
		req.Weather,
		req.BroadcastGroup,
	)
	col.Spaces[req.Name] = &space
	col.saveSpace(req.Name)

	encodeJSON(w, http.StatusCreated, map[string]string{"status": "created", "spaceName": req.Name})
}

func (c *Context) apiCreateAreaHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	req, err := decodeJSONBody[createAreaRequest](r)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	col := c.Collections[req.CollectionName]
	if col == nil {
		writeJSONError(w, http.StatusNotFound, "collection not found")
		return
	}
	space := col.Spaces[req.SpaceName]
	if space == nil {
		writeJSONError(w, http.StatusNotFound, "space not found")
		return
	}
	if strings.TrimSpace(req.Name) == "" {
		writeJSONError(w, http.StatusBadRequest, "area name required")
		return
	}
	if req.Height <= 0 || req.Width <= 0 {
		writeJSONError(w, http.StatusBadRequest, "invalid dimensions")
		return
	}
	if getAreaByName(space.Areas, req.Name) != nil {
		writeJSONError(w, http.StatusConflict, "area already exists")
		return
	}

	area := createBaseArea(req.Height, req.Width, req.DefaultTileColor, req.DefaultTileColor1, "", "")
	area.Name = req.Name
	area.Safe = req.Safe
	space.Areas = append(space.Areas, area)
	col.saveSpace(req.SpaceName)

	encodeJSON(w, http.StatusCreated, map[string]string{"status": "created", "areaName": req.Name})
}

func (c *Context) apiFlattenSpaceHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	req, err := decodeJSONBody[flattenSpaceRequest](r)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	collectionName := strings.TrimSpace(req.CollectionName)
	spaceName := strings.TrimSpace(req.SpaceName)
	if collectionName == "" || spaceName == "" {
		writeJSONError(w, http.StatusBadRequest, "collectionName and spaceName are required")
		return
	}

	col := c.Collections[collectionName]
	if col == nil {
		writeJSONError(w, http.StatusNotFound, "collection not found")
		return
	}

	space := col.Spaces[spaceName]
	if space == nil {
		writeJSONError(w, http.StatusNotFound, "space not found")
		return
	}

	if !space.isSimplyTiled() {
		writeJSONError(w, http.StatusBadRequest, "only simply tiled spaces may be flattened")
		return
	}

	flattened, err := Flatten(*space)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	flattened.CollectionName = collectionName
	col.Spaces[flattened.Name] = &flattened
	col.saveSpace(flattened.Name)

	encodeJSON(w, http.StatusCreated, map[string]string{"status": "created", "spaceName": flattened.Name})
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
		w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
		http.ServeFile(w, r, filepath.Join(distRoot, "index.html"))
		return
	}

	rel = strings.TrimPrefix(rel, "/")
	target := filepath.Join(distRoot, filepath.Clean(rel))
	if info, err := os.Stat(target); err == nil && !info.IsDir() {
		http.ServeFile(w, r, target)
		return
	}

	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
	http.ServeFile(w, r, filepath.Join(distRoot, "index.html"))
}
