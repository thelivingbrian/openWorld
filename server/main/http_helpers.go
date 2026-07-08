package main

import (
	"net/http"
)

//////////////////////////////////////////////////////////
// Forms

func requestToProperties(r *http.Request) (map[string]string, bool) {
	if err := r.ParseForm(); err != nil {
		logger.Error().Err(err).Msg("Error parsing form body")
		return nil, false
	}

	properties := make(map[string]string, len(r.PostForm))
	for key, values := range r.PostForm {
		if len(values) > 0 {
			// Preserve the old parser's behavior for duplicate fields.
			properties[key] = values[len(values)-1]
		}
	}
	return properties, true
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
