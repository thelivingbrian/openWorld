package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// Replace this with r.ParseForm() / r.FormValue and/or r.Form
func requestToProperties(r *http.Request) (map[string]string, bool) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("Error reading body: %v", err)
		return nil, false
	}

	bodyS := string(body[:])

	// Not great, '&' included in body will be unescaped here producing weird results
	bodyS, err = url.QueryUnescape(bodyS)
	if err != nil {
		return nil, false
	}
	//fmt.Println(bodyS)

	return bodyStringToProperties(bodyS), true
}

// Replace this with r.ParseForm() / r.FormValue and/or r.Form
func bodyStringToProperties(body string) map[string]string {
	propMap := make(map[string]string)
	props := strings.Split(body, "&")
	for _, prop := range props {
		keyValue := strings.Split(prop, "=")
		if len(keyValue) > 1 {
			propMap[keyValue[0]] = keyValue[1] // 1: ?
		}
	}
	return propMap
}

func (c Context) collectionFromGet(r *http.Request) *Collection {
	queryValues := r.URL.Query()
	collectionName := queryValues.Get("currentCollection")
	return c.Collections[collectionName]
}

func (c Context) spaceFromGET(r *http.Request) *Space {
	queryValues := r.URL.Query()
	collectionName := queryValues.Get("currentCollection")
	spaceName := queryValues.Get("currentSpace")

	return c.spaceFromNames(collectionName, spaceName)
}

func (c Context) areaFromGET(r *http.Request) *AreaDescription {
	space := c.spaceFromGET(r)
	if space == nil {
		return nil
	}
	queryValues := r.URL.Query()
	name := queryValues.Get("area-name")
	return getAreaByName(space.Areas, name)
}
