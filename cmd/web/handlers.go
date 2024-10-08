package main

import (
	"echo4eva/loona/internal/spotify"
	"echo4eva/loona/internal/youtube"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

type playlistReplaceForm struct {
	Playlist   string `form:"playlist"`
	FieldError string `form:"-"`
}

func (app *application) helloWorld(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.SpotifyForm = playlistReplaceForm{}
	data.YoutubeForm = playlistReplaceForm{}
	app.render(w, r, http.StatusOK, "home.html", data)
}

func (app *application) spotifyLogin(w http.ResponseWriter, r *http.Request) {

	state := generateRandomState()

	app.sessionManager.Put(r.Context(), "oauth_state", state)

	url := app.oAuth.AuthCodeURL(state)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (app *application) youtubeLogin(w http.ResponseWriter, r *http.Request) {

	state := generateRandomState()

	app.sessionManager.Put(r.Context(), "oauth_state", state)

	url := app.youtubeOAuth.AuthCodeURL(state)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (app *application) youtubeCallback(w http.ResponseWriter, r *http.Request) {
	receivedState := r.URL.Query().Get("state")

	storedState := app.sessionManager.GetString(r.Context(), "oauth_state")
	if storedState == "" {
		http.Error(w, "State not found", http.StatusBadRequest)
		return
	}

	if receivedState != storedState {
		http.Error(w, "Invalid state parameter", http.StatusBadRequest)
		return
	}

	app.sessionManager.Remove(r.Context(), "oauth_state")

	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Code not found", http.StatusBadRequest)
		return
	}

	token, err := app.youtubeOAuth.Exchange(r.Context(), code)
	if err != nil {
		http.Error(w, "Failed to exchange token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	err = app.sessionManager.RenewToken(r.Context())
	if err != nil {
		http.Error(w, "Failed to renew session: "+err.Error(), http.StatusInternalServerError)
		return
	}

	app.sessionManager.Put(r.Context(), "youtube_token", token)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *application) spotifyCallback(w http.ResponseWriter, r *http.Request) {

	receivedState := r.URL.Query().Get("state")

	storedState := app.sessionManager.GetString(r.Context(), "oauth_state")
	if storedState == "" {
		http.Error(w, "State not found", http.StatusBadRequest)
		return
	}

	if receivedState != storedState {
		http.Error(w, "Invalid state parameter", http.StatusBadRequest)
		return
	}

	app.sessionManager.Remove(r.Context(), "oauth_state")

	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Code not found", http.StatusBadRequest)
		return
	}

	token, err := app.oAuth.Exchange(r.Context(), code)
	if err != nil {
		http.Error(w, "Failed to exchange token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	err = app.sessionManager.RenewToken(r.Context())
	if err != nil {
		http.Error(w, "Failed to renew session: "+err.Error(), http.StatusInternalServerError)
		return
	}

	app.sessionManager.Put(r.Context(), "spotify_token", token)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *application) me(w http.ResponseWriter, r *http.Request) {
	client, err := app.getAuthenticatedSpotifyClient(r)
	if err != nil {
		http.Error(w, "Failed to get authenticated client: "+err.Error(), http.StatusInternalServerError)
		return
	}

	resp, err := client.Get("https://api.spotify.com/v1/me")
	if err != nil {
		http.Error(w, "Failed to fetch user info: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Failed to read response: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var userInfo map[string]interface{}
	err = json.Unmarshal(body, &userInfo)
	if err != nil {
		http.Error(w, "Failed to parse user info: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// You can now use the userInfo map to access the user's information
	// For example, to get the user's display name:
	displayName, ok := userInfo["display_name"].(string)
	if !ok {
		http.Error(w, "Display name not found", http.StatusInternalServerError)
		return
	}

	// Send the response back to the client
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"display_name": displayName})
}

func (app *application) getConversionMap(w http.ResponseWriter, r *http.Request) {
	songItems, err := loadJSONtoMap("data/sp_song_items.json")
	if err != nil {
		http.Error(w, "Failed to load episodeItems from JSON: "+err.Error(), http.StatusInternalServerError)
		return
	}

	episodeItems, err := loadJSONtoMap("data/sp_episode_items.json")
	if err != nil {
		http.Error(w, "Failed to load episodeItems from JSON: "+err.Error(), http.StatusInternalServerError)
		return
	}

	songsToEpisodeIDMap := make(map[string]string)
	for songName, songIDs := range songItems {
		for _, songID := range songIDs {
			if episodeID, exists := episodeItems[songName]; exists {
				songsToEpisodeIDMap[songID] = episodeID[0]
			} else {
				app.logger.Info(fmt.Sprintf("No matching episode found for song: %s", songName))
			}
		}
	}

	songsToEpisodeIDMapJSON, err := json.MarshalIndent(songsToEpisodeIDMap, "", "  ")
	if err != nil {
		http.Error(w, "Failed to marshal songsToEpisodeIDMap to JSON: "+err.Error(), http.StatusInternalServerError)
		return
	}
	err = os.WriteFile("data/sp_conversion_map.json", songsToEpisodeIDMapJSON, 0644)
	if err != nil {
		http.Error(w, "Failed to write songsToEpisodeIDMap to file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(songsToEpisodeIDMap); err != nil {
		http.Error(w, "Failed to encode songs to JSON: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func (app *application) getTestConversionMap(w http.ResponseWriter, r *http.Request) {

	client, err := app.getAuthenticatedSpotifyClient(r)
	if err != nil {
		http.Error(w, "Failed to get authenticated client: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// get items from song playlist
	songPlaylistID := "1VWVekcOvgOQOGrfJueTap"
	songItems, err := spotify.GetPlaylistItems(client, songPlaylistID)
	if err != nil {
		http.Error(w, "Failed to get playlist items: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// save songItems to a json file
	songItemsJSON, err := json.MarshalIndent(songItems, "", "  ")
	if err != nil {
		http.Error(w, "Failed to marshal songItems to JSON: "+err.Error(), http.StatusInternalServerError)
		return
	}
	err = os.WriteFile("temp_song_items.json", songItemsJSON, 0644)
	if err != nil {
		http.Error(w, "Failed to write songItems to file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// get items from episode playlist
	episodePlaylistID := "4A6yVtZHO6V3NfijUacSol"
	episodeItems, err := spotify.GetPlaylistItems(client, episodePlaylistID)
	if err != nil {
		http.Error(w, "Failed to get playlist items: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// save episodeItems to a json file
	episodeItemsJSON, err := json.MarshalIndent(episodeItems, "", "  ")
	if err != nil {
		http.Error(w, "Failed to marshal episodeItems to JSON: "+err.Error(), http.StatusInternalServerError)
		return
	}
	err = os.WriteFile("temp_episode_items.json", episodeItemsJSON, 0644)
	if err != nil {
		http.Error(w, "Failed to write episodeItems to file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	episodeItems, err = loadJSONtoMap("data/final_episode_items.json")
	if err != nil {
		http.Error(w, "Failed to load episodeItems from JSON: "+err.Error(), http.StatusInternalServerError)
		return
	}

	songsToEpisodeIDMap := make(map[string]string)
	for songName, songIDs := range songItems {
		for _, songID := range songIDs {
			if episodeID, exists := episodeItems[songName]; exists {
				songsToEpisodeIDMap[songID] = episodeID[0]
			} else {
				app.logger.Info(fmt.Sprintf("No matching episode found for song: %s", songName))
			}
		}
	}

	songsToEpisodeIDMapJSON, err := json.MarshalIndent(songsToEpisodeIDMap, "", "  ")
	if err != nil {
		http.Error(w, "Failed to marshal songsToEpisodeIDMap to JSON: "+err.Error(), http.StatusInternalServerError)
		return
	}
	err = os.WriteFile("temp_songs_to_episode_id_map.json", songsToEpisodeIDMapJSON, 0644)
	if err != nil {
		http.Error(w, "Failed to write songsToEpisodeIDMap to file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(songsToEpisodeIDMap); err != nil {
		http.Error(w, "Failed to encode songs to JSON: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func (app *application) test(w http.ResponseWriter, r *http.Request) {
	client, err := app.getAuthenticatedSpotifyClient(r)
	if err != nil {
		http.Error(w, "Failed to get authenticated client: "+err.Error(), http.StatusInternalServerError)
		return
	}

	playlistID := "7LQuibg5YN1O6mSi2KuZT8"
	trackIDs, err := spotify.GetPlaylistIDs(client, playlistID)
	if err != nil {
		http.Error(w, "Failed to get playlist items: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(trackIDs); err != nil {
		http.Error(w, "Failed to encode songs to JSON: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func (app *application) youtubeTest(w http.ResponseWriter, r *http.Request) {
	client, err := app.getAuthenticatedYoutubeClient(r)
	if err != nil {
		http.Error(w, "Failed to get authenticated client: "+err.Error(), http.StatusInternalServerError)
		return
	}

	thing, err := youtube.GetPlaylistIDs(client)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprint(w, thing)
}

func (app *application) youtubeTest2(w http.ResponseWriter, r *http.Request) {
	client, err := app.getAuthenticatedYoutubeClient(r)
	if err != nil {
		http.Error(w, "Failed to get authenticated client: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// weird4 PLXDnivaxCZoQ9p4tkW0LV0pP47-F2JwAB
	// boycott PLXDnivaxCZoTC_tcbuj0gRU5O_0NRhiHg
	name := "yt_playlist_boycott_items"
	thing, err := youtube.GetPlaylistItems(client, "PLXDnivaxCZoTC_tcbuj0gRU5O_0NRhiHg", name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprint(w, thing)
}

func (app *application) youtubeReplacePost(w http.ResponseWriter, r *http.Request) {

	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Failed to parse form: "+err.Error(), http.StatusInternalServerError)
		return
	}

	form := playlistReplaceForm{
		Playlist:   r.PostForm.Get("youtubePlaylist"),
		FieldError: "",
	}

	// check if valid playlist
	if !strings.HasPrefix(form.Playlist, "https://www.youtube.com/") ||
		!strings.HasPrefix(form.Playlist, "https://music.youtube.com/") {
		form.FieldError = "Please enter a valid Youtube playlist link"
	}

	if form.FieldError != "" {
		data := app.newTemplateData(r)
		data.YoutubeForm = form
		app.render(w, r, http.StatusUnprocessableEntity, "home.html", data)
		return
	}

	playlistID := getPlaylistID(form.Playlist)

	client, err := app.getAuthenticatedYoutubeClient(r)
	if err != nil {
		http.Error(w, "Failed to get authenticated client: "+err.Error(), http.StatusInternalServerError)
		return
	}

	err = youtube.UpdatePlaylistItems(client, playlistID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	app.sessionManager.Put(r.Context(), "youtubeFlash", "Playlist updated successfully!")

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *application) spotifyReplacePost(w http.ResponseWriter, r *http.Request) {

	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Failed to parse form: "+err.Error(), http.StatusInternalServerError)
		return
	}

	form := playlistReplaceForm{
		Playlist:   r.PostForm.Get("spotifyPlaylist"),
		FieldError: "",
	}

	// check if valid playlist
	if !strings.HasPrefix(form.Playlist, "https://open.spotify.com/playlist/") {
		form.FieldError = "Please enter a valid Spotify playlist link"
	}

	if form.FieldError != "" {
		data := app.newTemplateData(r)
		data.SpotifyForm = form
		app.render(w, r, http.StatusUnprocessableEntity, "home.html", data)
		return
	}

	playlistID := getPlaylistID(form.Playlist)

	client, err := app.getAuthenticatedSpotifyClient(r)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	err = spotify.UpdatePlaylistItems(client, playlistID)
	if err != nil {
		form.FieldError = "Error: User not in Beta, notify owner for access"
		data := app.newTemplateData(r)
		data.SpotifyForm = form
		app.render(w, r, http.StatusInternalServerError, "home.html", data)
		app.serverError(w, r, err)
		return
	}

	app.sessionManager.Put(r.Context(), "spotifyFlash", "Playlist updated successfully!")

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *application) spotifyLogout(w http.ResponseWriter, r *http.Request) {
	app.sessionManager.Remove(r.Context(), "spotify_token")

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *application) youtubeLogout(w http.ResponseWriter, r *http.Request) {
	app.sessionManager.Remove(r.Context(), "youtube_token")

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
