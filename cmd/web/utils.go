// utils.go

package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"

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

func (app *application) getAuthenticatedClient(r *http.Request) (*http.Client, error) {
	tokenInterface := app.sessionManager.Get(r.Context(), "spotify_token")
	if tokenInterface == nil {
		return nil, errors.New("no token found for the session")
	}

	token := tokenInterface.(*oauth2.Token)

	client := app.oAuth.Client(r.Context(), token)

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
