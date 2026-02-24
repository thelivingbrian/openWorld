package main

import (
	"regexp"
	"strings"
)

type Prototype struct {
	ID          string `json:"id"`
	CommonName  string `json:"commonName"`
	CssColor    string `json:"cssColor"`
	Walkable    bool   `json:"walkable"`
	Floor1Css   string `json:"layer1css"`
	Floor2Css   string `json:"layer2css"`
	Ceiling1Css string `json:"ceiling1css"`
	Ceiling2Css string `json:"ceiling2css"`
	SetName     string `json:"setName"`
	MapColor    string `json:"mapColor"`
	EditorColor string `json:"editorColor"`
	DisplayText string `json:"displayText"`
}

func (proto *Prototype) applyTransform(transformation Transformation) Material {
	return Material{
		Walkable:    proto.Walkable,
		Ground2Css:  proto.CssColor,
		Floor1Css:   transformCss(proto.Floor1Css, transformation),
		Floor2Css:   transformCss(proto.Floor2Css, transformation),
		Ceiling1Css: transformCss(proto.Ceiling1Css, transformation),
		Ceiling2Css: transformCss(proto.Ceiling2Css, transformation),
		DisplayText: proto.DisplayText,
	}
}

func transformCss(input string, transformation Transformation) string {
	// We are looking for {key:value} : key, value are strings
	pattern := regexp.MustCompile(`{([^:]*):([^}]*)}`)

	result := pattern.ReplaceAllStringFunc(input, func(s string) string {
		matches := pattern.FindStringSubmatch(s)
		// matches[0] is the full match, matches[1] is the key, matches[2] is the value
		if len(matches) == 3 {
			if matches[1] == "rotate" {
				return rotateCss(matches[2], transformation.ClockwiseRotations)
			}
			return matches[2]
		}
		panic("Have match " + s + " But submatch behavior is undefined (submatches != 3)")
	})
	return result
}

func rotateCss(input string, clockwiseRotations int) string {
	options := []string{"tr", "br", "bl", "tl"}
	currentIndex := findIndex(input, options)
	if currentIndex == -1 {
		panic("invalid rotation attempted")
	}
	return options[mod(currentIndex+clockwiseRotations, 4)]
}

func findIndex(s string, list []string) int {
	for i := range list {
		if list[i] == s {
			return i
		}
	}
	return -1
}

func (c Context) getMapColorFromProto(proto Prototype) string {
	if proto.MapColor != "" {
		return proto.MapColor
	}
	return c.inferMapColorFromProto(proto)
}

func (c Context) inferMapColorFromProto(proto Prototype) string {
	color := proto.CssColor
	layersToCheck := []string{proto.Floor1Css, proto.Floor2Css, proto.Ceiling1Css, proto.Ceiling2Css}
	for _, layerString := range layersToCheck {
		extractedColor := c.getColorFromString(layerString)
		if extractedColor != "" {
			color = extractedColor
		}
	}
	return color
}

func (c Context) getColorFromString(s string) string {
	words := strings.Fields(s)
	for _, word := range words {
		for _, color := range c.colors {
			if word == color.CssClassName {
				return word
			}
		}
	}
	return ""
}
