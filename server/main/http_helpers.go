package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

//////////////////////////////////////////////////////////
// Forms

// Replace this with r.ParseForm() / r.FormValue and/or r.Form
func requestToProperties(r *http.Request) (map[string]string, bool) {
	// Works on standard htmx form post
	body, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Error().Err(err).Msg(fmt.Sprintf("Error reading body: %v", err))
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
		logger.Error().Err(err).Msg("Error with session: ")
		return "", false
	}
	if session == nil {
		logger.Error().Msg("Session is nil")
		return "", false
	}

	id, ok := session.Values["identifier"].(string)
	if !ok {
		return "", false
	}
	return id, true
}

func signOutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		return
	}
	session, err := store.Get(r, "user-session")
	if err != nil {
		logger.Warn().Msg("signOut: could not get session: " + err.Error())
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	session.Options.MaxAge = -1
	if err := session.Save(r, w); err != nil {
		logger.Warn().Msg("signOut: could not save session: " + err.Error())
	}

	// Send the user back to the home page
	w.Header().Set("HX-Redirect", "/")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Sign out successful - Redirecting"))
}
