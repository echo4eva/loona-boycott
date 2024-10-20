// utils.go

package main

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"runtime/debug"
	"strings"

	"golang.org/x/oauth2"
)

func generateRandomState() string {
	// initialize slice to hold 16 bytes
	b := make([]byte, 16)
	// files slice with random bytes
	_, err := rand.Read(b)
	if err != nil {
		return ""
	}
	// encode slice to base64 string
	return base64.URLEncoding.EncodeToString(b)
}

func getPlaylistID(url string) string {

	if strings.Contains(url, "spotify.com") {
		// Find the index of "playlist/" in the URL
		start := strings.Index(url, "playlist/")
		if start == -1 {
			return "" // "playlist/" not found in the URL
		}
		start += len("playlist/") // Move to the start of the ID

		// Find the index of "?" after "playlist/"
		end := strings.Index(url[start:], "?")
		if end == -1 {
			// If "?" is not found, return the rest of the string
			return url[start:]
		}

		// Return the substring between "playlist/" and "?"
		return url[start : start+end]
	} else if strings.Contains(url, "youtube.com") {

		start := strings.Index(url, "list=")
		if start == -1 {
			return ""
		}

		start += len("list=") // Move to the start of the ID

		return url[start:]
	}

	return ""
}

func (app *application) getAuthenticatedSpotifyClient(r *http.Request) (*http.Client, error) {
	tokenInterface := app.sessionManager.Get(r.Context(), "spotify_token")
	if tokenInterface == nil {
		return nil, errors.New("no token found for the session")
	}

	token := tokenInterface.(*oauth2.Token)

	client := app.oAuth.Client(r.Context(), token)

	return client, nil
}

func (app *application) getAuthenticatedYoutubeClient(r *http.Request) (*http.Client, error) {
	tokenInterface := app.sessionManager.Get(r.Context(), "youtube_token")
	if tokenInterface == nil {
		return nil, errors.New("no token found for the session")
	}

	token := tokenInterface.(*oauth2.Token)

	client := app.youtubeOAuth.Client(r.Context(), token)

	return client, nil
}

func loadJSONtoMap(filename string) (map[string][]string, error) {
	// Read the file
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %v", err)
	}

	// Unmarshal JSON data into map
	var results map[string][]string
	err = json.Unmarshal(data, &results)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling JSON: %v", err)
	}

	return results, nil
}

func (app *application) serverError(w http.ResponseWriter, r *http.Request, err error) {
	var (
		method = r.Method
		uri    = r.URL.RequestURI()
		trace  = string(debug.Stack())
	)

	app.logger.Error(err.Error(), "method", method, "uri", uri, "trace", trace)
}

func (app *application) clientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

func (app *application) render(w http.ResponseWriter, r *http.Request, status int, page string, data templateData) {
	ts, ok := app.templateCache[page]
	if !ok {
		err := errors.New("the template does not exist")
		app.serverError(w, r, err)
		return
	}

	buf := new(bytes.Buffer)

	err := ts.ExecuteTemplate(buf, "base", data)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	w.WriteHeader(status)
	buf.WriteTo(w)
}

func (app *application) isSpotifyAuthenticated(r *http.Request) bool {
	isAuthenticated, exists := r.Context().Value(isSpotifyAuthenticatedContextKey).(bool)
	if !exists {
		return false
	}

	return isAuthenticated
}

func (app *application) isYoutubeAuthenticated(r *http.Request) bool {
	isAuthenticated, exists := r.Context().Value(isYoutubeAuthenticatedContextKey).(bool)
	if !exists {
		return false
	}

	return isAuthenticated
}

func (app *application) newTemplateData(r *http.Request) templateData {
	return templateData{
		SpotifyFlash:           app.sessionManager.PopString(r.Context(), "spotifyFlash"),
		IsSpotifyAuthenticated: app.isSpotifyAuthenticated(r),
		YoutubeFlash:           app.sessionManager.PopString(r.Context(), "youtubeFlash"),
		IsYoutubeAuthenticated: app.isYoutubeAuthenticated(r),
	}
}
