package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

type SourceCollection struct {
	Name             string                               `json:"Name"`
	Spaces           map[string]*SourceSpace              `json:"Spaces"`
	Fragments        map[string]json.RawMessage           `json:"Fragments"`
	PrototypeSets    map[string][]SourcePrototype         `json:"PrototypeSets"`
	InteractableSets map[string][]InteractableDescription `json:"InteractableSets"`
}

type SourceSpace struct {
	Name       string
	Topology   string
	Latitude   int
	Longitude  int
	AreaHeight int
	AreaWidth  int
	Areas      []SourceArea
}

type SourceArea struct {
	Name           string           `json:"Name"`
	Safe           bool             `json:"Safe"`
	Blueprint      *SourceBlueprint `json:"Blueprint"`
	Transports     []Transport      `json:"Transports"`
	North          string           `json:"North"`
	South          string           `json:"South"`
	East           string           `json:"East"`
	West           string           `json:"West"`
	LoadStrategy   string           `json:"LoadStrategy"`
	SpawnStrategy  string           `json:"SpawnStrategy"`
	BroadcastGroup string           `json:"BroadcastGroup"`
	Weather        string           `json:"Weather"`
}

type SourceBlueprint struct {
	Tiles             [][]SourceTile `json:"Tiles"`
	Ground            [][]SourceCell `json:"Ground"`
	DefaultTileColor  string         `json:"DefaultTileColor"`
	DefaultTileColor1 string         `json:"DefaultTileColor1"`
}

type SourceTile struct {
	PrototypeID       string               `json:"prototypeId"`
	Transformation    SourceTransformation `json:"transformation"`
	InteractableID    string               `json:"interactableId"`
	InteractableState string               `json:"interactableState"`
}

type SourceTransformation struct {
	ClockwiseRotations int `json:"clockwiseRotations"`
}
type SourceCell struct {
	Status      int
	BottomRight bool
	BottomLeft  bool
	TopRight    bool
	TopLeft     bool
}

type SourcePrototype struct {
	ID          string `json:"id"`
	CommonName  string `json:"commonName"`
	CssColor    string `json:"cssColor"`
	Walkable    bool   `json:"walkable"`
	Floor1Css   string `json:"layer1css"`
	Floor2Css   string `json:"layer2css"`
	Ceiling1Css string `json:"ceiling1css"`
	Ceiling2Css string `json:"ceiling2css"`
	MapColor    string `json:"mapColor"`
	DisplayText string `json:"displayText"`
}

func compileWorldResources(resources []WorldResource, admin bool) (CompiledRelease, error) {
	source, sourceHash, err := canonicalResources(resources)
	if err != nil {
		return CompiledRelease{}, err
	}
	collection, err := decodeSourceCollection(resources)
	if err != nil {
		return CompiledRelease{}, err
	}
	var manifest WorldManifest
	var palette []WorldColor
	for _, resource := range resources {
		switch resource.Kind + "/" + resource.Key {
		case "manifest/world":
			err = json.Unmarshal(resource.Content, &manifest)
		case "palette/colors":
			palette, err = decodeWorldPalette(resource.Content)
		}
		if err != nil {
			return CompiledRelease{}, fmt.Errorf("decode %s/%s: %w", resource.Kind, resource.Key, err)
		}
	}
	if collection.Name == "" {
		return CompiledRelease{}, fmt.Errorf("collection resources are required")
	}
	if err := validateManifest(manifest, admin); err != nil {
		return CompiledRelease{}, err
	}
	if err := validatePalette(palette); err != nil {
		return CompiledRelease{}, err
	}
	files, err := compileCollectionFiles(collection, palette)
	if err != nil {
		return CompiledRelease{}, err
	}
	manifestJSON, _ := json.Marshal(manifest)
	files["manifest.json"] = manifestJSON
	artifact, err := writeDeterministicArchive(files)
	if err != nil {
		return CompiledRelease{}, err
	}
	return CompiledRelease{Source: source, Artifact: artifact, SourceHash: sourceHash, ArtifactHash: hashBytes(artifact)}, nil
}

func decodeSourceCollection(resources []WorldResource) (SourceCollection, error) {
	collection := SourceCollection{
		Spaces:           map[string]*SourceSpace{},
		Fragments:        map[string]json.RawMessage{},
		PrototypeSets:    map[string][]SourcePrototype{},
		InteractableSets: map[string][]InteractableDescription{},
	}
	foundLegacy := false
	foundGranular := false
	for _, resource := range resources {
		switch {
		case resource.Kind == "collection" && resource.Key == "source":
			if err := json.Unmarshal(resource.Content, &collection); err != nil {
				return SourceCollection{}, fmt.Errorf("decode %s/%s: %w", resource.Kind, resource.Key, err)
			}
			if collection.Spaces == nil {
				collection.Spaces = map[string]*SourceSpace{}
			}
			if collection.Fragments == nil {
				collection.Fragments = map[string]json.RawMessage{}
			}
			if collection.PrototypeSets == nil {
				collection.PrototypeSets = map[string][]SourcePrototype{}
			}
			if collection.InteractableSets == nil {
				collection.InteractableSets = map[string][]InteractableDescription{}
			}
			foundLegacy = true
		case resource.Kind == "collection" && resource.Key == "meta":
			var meta struct {
				Name string `json:"name"`
			}
			if err := json.Unmarshal(resource.Content, &meta); err != nil {
				return SourceCollection{}, fmt.Errorf("decode %s/%s: %w", resource.Kind, resource.Key, err)
			}
			collection.Name = meta.Name
			foundGranular = true
		case resource.Kind == "space":
			var space SourceSpace
			if err := json.Unmarshal(resource.Content, &space); err != nil {
				return SourceCollection{}, fmt.Errorf("decode %s/%s: %w", resource.Kind, resource.Key, err)
			}
			collection.Spaces[resource.Key] = &space
			foundGranular = true
		case resource.Kind == "prototype-set":
			var set []SourcePrototype
			if err := json.Unmarshal(resource.Content, &set); err != nil {
				return SourceCollection{}, fmt.Errorf("decode %s/%s: %w", resource.Kind, resource.Key, err)
			}
			collection.PrototypeSets[resource.Key] = set
			foundGranular = true
		case resource.Kind == "interactable-set":
			var set []InteractableDescription
			if err := json.Unmarshal(resource.Content, &set); err != nil {
				return SourceCollection{}, fmt.Errorf("decode %s/%s: %w", resource.Kind, resource.Key, err)
			}
			collection.InteractableSets[resource.Key] = set
			foundGranular = true
		case resource.Kind == "fragment-set":
			collection.Fragments[resource.Key] = append(json.RawMessage(nil), resource.Content...)
			foundGranular = true
		}
	}
	if foundLegacy || foundGranular {
		return collection, nil
	}
	return SourceCollection{}, nil
}

func decodeWorldPalette(data []byte) ([]WorldColor, error) {
	var raw []struct {
		Name         string   `json:"name"`
		CSSClassName string   `json:"cssClassName"`
		R            uint8    `json:"R"`
		LowerR       uint8    `json:"r"`
		G            uint8    `json:"G"`
		LowerG       uint8    `json:"g"`
		B            uint8    `json:"B"`
		LowerB       uint8    `json:"b"`
		A            any      `json:"A"`
		LowerA       *float64 `json:"a"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}
	out := make([]WorldColor, 0, len(raw))
	for _, entry := range raw {
		name := entry.Name
		if name == "" {
			name = entry.CSSClassName
		}
		r := entry.R
		if r == 0 {
			r = entry.LowerR
		}
		g := entry.G
		if g == 0 {
			g = entry.LowerG
		}
		b := entry.B
		if b == 0 {
			b = entry.LowerB
		}
		alpha := 1.0
		if entry.LowerA != nil {
			alpha = *entry.LowerA
		} else if text, ok := entry.A.(string); ok && text != "" {
			if parsed, err := strconv.ParseFloat(text, 64); err == nil {
				alpha = parsed
			}
		}
		out = append(out, WorldColor{Name: name, R: r, G: g, B: b, A: alpha})
	}
	return out, nil
}

func compileCollectionFiles(collection SourceCollection, palette []WorldColor) (map[string][]byte, error) {
	prototypes := map[string]SourcePrototype{}
	for _, set := range collection.PrototypeSets {
		for _, prototype := range set {
			if err := validatePrototypeSource(prototype); err != nil {
				return nil, err
			}
			if _, ok := prototypes[prototype.ID]; ok {
				return nil, fmt.Errorf("duplicate prototype %q", prototype.ID)
			}
			prototypes[prototype.ID] = prototype
		}
	}
	interactables := map[string]InteractableDescription{}
	for _, set := range collection.InteractableSets {
		for _, item := range set {
			if err := validateInteractableSource(item); err != nil {
				return nil, err
			}
			if _, ok := interactables[item.ID]; ok {
				return nil, fmt.Errorf("duplicate interactable %q", item.ID)
			}
			interactables[item.ID] = item
		}
	}
	colors := map[string]WorldColor{}
	for _, entry := range palette {
		colors[entry.Name] = entry
	}
	spaceNames := make([]string, 0, len(collection.Spaces))
	for name := range collection.Spaces {
		spaceNames = append(spaceNames, name)
	}
	sort.Strings(spaceNames)
	compiled := make([]Area, 0)
	totalTiles := 0
	files := map[string][]byte{}
	mapAreas := map[string]WorldMapArea{}
	for _, spaceName := range spaceNames {
		space := collection.Spaces[spaceName]
		if space == nil {
			continue
		}
		mapID := ""
		if space.Topology == "plane" || space.Topology == "torus" {
			mapID = "space-" + hashBytes([]byte(spaceName))[:16]
			mapPNG, err := renderSpaceMap(space, prototypes, colors)
			if err != nil {
				return nil, err
			}
			files["images/"+mapID+".png"] = mapPNG
			for row := 0; row < space.Latitude; row++ {
				for column := 0; column < space.Longitude; column++ {
					stage := fmt.Sprintf("%s:%d-%d", space.Name, row, column)
					mapAreas[stage] = WorldMapArea{MapID: mapID, Top: float64(row) / float64(space.Latitude) * 100, Left: float64(column) / float64(space.Longitude) * 100, Width: 100 / float64(space.Longitude), Height: 100 / float64(space.Latitude)}
				}
			}
		}
		areas := append([]SourceArea(nil), space.Areas...)
		sort.Slice(areas, func(i, j int) bool { return areas[i].Name < areas[j].Name })
		for _, sourceArea := range areas {
			area, tiles, err := compileArea(sourceArea, mapID, prototypes, interactables, colors)
			if err != nil {
				return nil, err
			}
			totalTiles += tiles
			if totalTiles > maxWorldTiles {
				return nil, fmt.Errorf("world exceeds %d tiles", maxWorldTiles)
			}
			compiled = append(compiled, area)
		}
	}
	areasJSON, err := json.Marshal(compiled)
	if err != nil {
		return nil, err
	}
	files["areas.json"] = areasJSON
	files["colors.css"] = []byte(paletteCSS(palette))
	mapsJSON, err := json.Marshal(mapAreas)
	if err != nil {
		return nil, err
	}
	files["maps.json"] = mapsJSON
	return files, nil
}

func validatePrototypeSource(prototype SourcePrototype) error {
	if prototype.ID != "" {
		if err := validateResourcePart(prototype.ID); err != nil {
			return fmt.Errorf("prototype: %w", err)
		}
	}
	for _, value := range []string{prototype.CssColor, prototype.Floor1Css, prototype.Floor2Css, prototype.Ceiling1Css, prototype.Ceiling2Css, prototype.MapColor} {
		if err := validateClassTokens(value); err != nil {
			return fmt.Errorf("prototype %q: %w", prototype.ID, err)
		}
	}
	if len(prototype.DisplayText) > 2000 || strings.ContainsAny(prototype.DisplayText, "<>") {
		return fmt.Errorf("prototype %q contains unsafe display text", prototype.ID)
	}
	return nil
}

func validateInteractableSource(item InteractableDescription) error {
	if item.ID != "" {
		if err := validateResourcePart(item.ID); err != nil {
			return fmt.Errorf("interactable: %w", err)
		}
	}
	if err := validateClassTokens(item.CssClass); err != nil {
		return fmt.Errorf("interactable %q: %w", item.ID, err)
	}
	if item.Reactions != "" {
		if _, ok := interactableReactions[item.Reactions]; !ok {
			return fmt.Errorf("interactable %q uses unknown reaction set %q", item.ID, item.Reactions)
		}
	}
	validateRules := func(rules []ReactionRule) error {
		for _, rule := range rules {
			if _, ok := reactsWithRegistry[rule.ReactsWith]; !ok {
				return fmt.Errorf("unknown reactsWith %q", rule.ReactsWith)
			}
			if _, ok := reactionRegistry[rule.Reaction]; !ok {
				if _, offsetOK := reactionWithOffsetRegistry[rule.Reaction]; !offsetOK {
					return fmt.Errorf("unknown reaction %q", rule.Reaction)
				}
			}
		}
		return nil
	}
	if err := validateRules(item.ReactionRules); err != nil {
		return fmt.Errorf("interactable %q: %w", item.ID, err)
	}
	for stateName, state := range item.States {
		if err := validateClassTokens(state.CssClass); err != nil {
			return fmt.Errorf("interactable %q state %q: %w", item.ID, stateName, err)
		}
		if state.Reactions != "" {
			if _, ok := interactableReactions[state.Reactions]; !ok {
				return fmt.Errorf("interactable %q state %q uses unknown reaction set", item.ID, stateName)
			}
		}
		if err := validateRules(state.ReactionRules); err != nil {
			return fmt.Errorf("interactable %q state %q: %w", item.ID, stateName, err)
		}
	}
	return nil
}

func validateClassTokens(value string) error {
	lower := strings.ToLower(value)
	if len(value) > 512 || strings.ContainsAny(value, ";<>") || strings.Contains(lower, "javascript:") || strings.Contains(lower, "url(") {
		return errors.New("unsafe class token string")
	}
	return nil
}

func compileArea(source SourceArea, mapID string, prototypes map[string]SourcePrototype, interactables map[string]InteractableDescription, colors map[string]WorldColor) (Area, int, error) {
	if source.Blueprint == nil || len(source.Blueprint.Tiles) == 0 {
		return Area{}, 0, fmt.Errorf("area %q has no tiles", source.Name)
	}
	height := len(source.Blueprint.Tiles)
	if height > maxAreaSide {
		return Area{}, 0, fmt.Errorf("area %q exceeds maximum height", source.Name)
	}
	tiles := make([][]Material, height)
	outInteractables := make([][]*InteractableDescription, height)
	count := 0
	for y, row := range source.Blueprint.Tiles {
		if len(row) == 0 || len(row) > maxAreaSide {
			return Area{}, 0, fmt.Errorf("area %q has invalid width", source.Name)
		}
		tiles[y] = make([]Material, len(row))
		outInteractables[y] = make([]*InteractableDescription, len(row))
		count += len(row)
		for x, cell := range row {
			prototype, ok := prototypes[cell.PrototypeID]
			if !ok {
				return Area{}, 0, fmt.Errorf("area %q references unknown prototype %q at %d,%d", source.Name, cell.PrototypeID, y, x)
			}
			material := Material{Walkable: prototype.Walkable, Ground2Css: prototype.CssColor, Floor1Css: prototype.Floor1Css, Floor2Css: prototype.Floor2Css, Ceiling1Css: prototype.Ceiling1Css, Ceiling2Css: prototype.Ceiling2Css, DisplayText: prototype.DisplayText}
			material.Floor1Css = rotateSourceCSS(material.Floor1Css, cell.Transformation.ClockwiseRotations)
			material.Floor2Css = rotateSourceCSS(material.Floor2Css, cell.Transformation.ClockwiseRotations)
			material.Ceiling1Css = rotateSourceCSS(material.Ceiling1Css, cell.Transformation.ClockwiseRotations)
			material.Ceiling2Css = rotateSourceCSS(material.Ceiling2Css, cell.Transformation.ClockwiseRotations)
			primary := source.Blueprint.DefaultTileColor
			secondary := source.Blueprint.DefaultTileColor1
			if _, ok := colors[primary]; primary != "" && !ok {
				return Area{}, 0, fmt.Errorf("area %q references unknown color %q", source.Name, primary)
			}
			if y < len(source.Blueprint.Ground) && x < len(source.Blueprint.Ground[y]) {
				ground := source.Blueprint.Ground[y][x]
				if ground.Status != 0 {
					primary, secondary = secondary, primary
				}
				material.Ground2Css = primary
				if ground.BottomRight || ground.BottomLeft || ground.TopRight || ground.TopLeft {
					material.Ground1Css = secondary
				}
				if ground.TopLeft {
					material.Ground2Css += " r0-tl"
				}
				if ground.TopRight {
					material.Ground2Css += " r0-tr"
				}
				if ground.BottomLeft {
					material.Ground2Css += " r0-bl"
				}
				if ground.BottomRight {
					material.Ground2Css += " r0-br"
				}
			} else {
				material.Ground1Css = primary
			}
			tiles[y][x] = material
			if base, ok := interactables[cell.InteractableID]; ok {
				resolved := resolvePublishedInteractable(base, cell.InteractableState)
				outInteractables[y][x] = &resolved
			} else if cell.InteractableID != "" {
				return Area{}, 0, fmt.Errorf("area %q references unknown interactable %q", source.Name, cell.InteractableID)
			}
		}
	}
	return Area{Name: source.Name, Safe: source.Safe, Tiles: tiles, Interactables: outInteractables, Transports: source.Transports, North: source.North, South: source.South, East: source.East, West: source.West, MapId: mapID, LoadStrategy: source.LoadStrategy, SpawnStrategy: source.SpawnStrategy, BroadcastGroup: source.BroadcastGroup, Weather: source.Weather}, count, nil
}

func resolvePublishedInteractable(out InteractableDescription, stateName string) InteractableDescription {
	if out.DefaultState == "" {
		out.DefaultState = "default"
	}
	if stateName == "" {
		stateName = out.DefaultState
	}
	state, ok := out.States[stateName]
	if !ok {
		return out
	}
	out.State = stateName
	out.CssClass = state.CssClass
	out.Pushable = state.Pushable
	out.Walkable = state.Walkable
	out.Fragile = state.Fragile
	out.StickyGroups = append([]string(nil), state.StickyGroups...)
	out.RejectTeleport = state.RejectTeleport
	out.Reactions = state.Reactions
	out.ReactionRules = append([]ReactionRule(nil), state.ReactionRules...)
	return out
}

func rotateSourceCSS(value string, rotations int) string {
	for _, pair := range []struct{ from, to string }{{"{rotate:tr}", []string{"tr", "br", "bl", "tl"}[(rotations%4+4)%4]}, {"{rotate:br}", []string{"br", "bl", "tl", "tr"}[(rotations%4+4)%4]}, {"{rotate:bl}", []string{"bl", "tl", "tr", "br"}[(rotations%4+4)%4]}, {"{rotate:tl}", []string{"tl", "tr", "br", "bl"}[(rotations%4+4)%4]}} {
		value = strings.ReplaceAll(value, pair.from, pair.to)
	}
	return value
}

func renderSpaceMap(space *SourceSpace, prototypes map[string]SourcePrototype, colors map[string]WorldColor) ([]byte, error) {
	width := space.Longitude * space.AreaWidth
	height := space.Latitude * space.AreaHeight
	if width < 1 || height < 1 || width > 4096 || height > 4096 {
		return nil, fmt.Errorf("space %q has invalid map dimensions", space.Name)
	}
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	areas := map[string]SourceArea{}
	for _, area := range space.Areas {
		areas[area.Name] = area
	}
	for row := 0; row < space.Latitude; row++ {
		for column := 0; column < space.Longitude; column++ {
			area, ok := areas[fmt.Sprintf("%s:%d-%d", space.Name, row, column)]
			if !ok || area.Blueprint == nil {
				continue
			}
			for y, tiles := range area.Blueprint.Tiles {
				for x, tile := range tiles {
					colorName := area.Blueprint.DefaultTileColor
					if prototype, ok := prototypes[tile.PrototypeID]; ok && prototype.MapColor != "" {
						colorName = prototype.MapColor
					}
					entry, ok := colors[colorName]
					if !ok {
						continue
					}
					img.Set(column*space.AreaWidth+x, row*space.AreaHeight+y, color.RGBA{entry.R, entry.G, entry.B, 255})
				}
			}
		}
	}
	var out bytes.Buffer
	if err := png.Encode(&out, img); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

func paletteCSS(palette []WorldColor) string {
	sort.Slice(palette, func(i, j int) bool { return palette[i].Name < palette[j].Name })
	var out strings.Builder
	for _, entry := range palette {
		fmt.Fprintf(&out, ".%s{background-color:rgba(%d,%d,%d,%.3f)}\n.%s-b{border-color:rgba(%d,%d,%d,%.3f)}\n.%s-t{color:rgba(%d,%d,%d,%.3f)}\n", entry.Name, entry.R, entry.G, entry.B, entry.A, entry.Name, entry.R, entry.G, entry.B, entry.A, entry.Name, entry.R, entry.G, entry.B, entry.A)
	}
	return out.String()
}

func writeDeterministicArchive(files map[string][]byte) ([]byte, error) {
	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	sort.Strings(names)
	var out bytes.Buffer
	zw := zip.NewWriter(&out)
	for _, name := range names {
		clean := filepath.ToSlash(filepath.Clean(name))
		if strings.HasPrefix(clean, "../") || strings.HasPrefix(clean, "/") {
			return nil, fmt.Errorf("unsafe archive path %q", name)
		}
		data := files[name]
		header := &zip.FileHeader{Name: clean, Method: zip.Deflate}
		header.SetMode(0444)
		header.SetModTime(time.Date(1980, time.January, 1, 0, 0, 0, 0, time.UTC))
		writer, err := zw.CreateHeader(header)
		if err != nil {
			return nil, err
		}
		if _, err := writer.Write(data); err != nil {
			return nil, err
		}
	}
	if err := zw.Close(); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

func extractReleaseArchive(archive []byte, directory string) error {
	zr, err := zip.NewReader(bytes.NewReader(archive), int64(len(archive)))
	if err != nil {
		return err
	}
	for _, entry := range zr.File {
		clean := filepath.Clean(entry.Name)
		target := filepath.Join(directory, clean)
		rel, err := filepath.Rel(directory, target)
		if err != nil || strings.HasPrefix(rel, "..") || filepath.IsAbs(rel) {
			return fmt.Errorf("unsafe archive path %q", entry.Name)
		}
		if entry.FileInfo().IsDir() {
			return fmt.Errorf("unsupported archive entry %q", entry.Name)
		}
		reader, err := entry.Open()
		if err != nil {
			return err
		}
		if err := writeReadonlyFile(target, reader, int64(entry.UncompressedSize64)); err != nil {
			reader.Close()
			return err
		}
		if err := reader.Close(); err != nil {
			return err
		}
	}
	return nil
}

func writeReadonlyFile(path string, source io.Reader, size int64) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0444)
	if err != nil {
		return err
	}
	_, copyErr := io.CopyN(file, source, size)
	closeErr := file.Close()
	if copyErr != nil {
		return copyErr
	}
	return closeErr
}
