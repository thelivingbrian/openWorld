package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

//////////////////////////////////////////////////////////
// Forms

func requestToProperties(r *http.Request) (map[string]string, bool) {
	// Works on standard htmx form post
	body, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("Error reading body: %v", err)
		return nil, false
	}

	bodyS := string(body[:])
	return bodyStringToProperties(bodyS), true
}

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

/////////////////////////////////////////////////////////
// OAuth

func getUserIdFromSession(r *http.Request) (string, bool) {
	session, err := store.Get(r, "user-session")
	if err != nil {
		fmt.Println("Error with session: ")
		fmt.Println(err)
		return "", false
	}
	if session == nil {
		fmt.Println("Session is nil")
	}

	id, ok := session.Values["identifier"].(string)
	if !ok {
		return "", false
	}
	return id, true
}
