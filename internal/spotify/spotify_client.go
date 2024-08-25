package spotify

import (
	"bytes"
	"echo4eva/loona/internal/utils"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

type Track struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type TrackItem struct {
	Track Track `json:"track"`
}

type PlaylistResponse struct {
	Items []TrackItem `json:"items"`
	Next  string      `json:"next"`
}

type TrackURI struct {
	URI string `json:"uri"`
}

type DeleteTracksRequest struct {
	Items []TrackURI `json:"tracks"`
}

type PostTracksRequest struct {
	URIs []string `json:"uris"`
}

func GetPlaylistItems(client *http.Client, playlistID string) (map[string][]string, error) {
	// "song name" : ["id1", "id2"]
	// there are duplicates of songs with different ids
	nameToIDsMap := make(map[string][]string)
	nextURL := fmt.Sprintf("https://api.spotify.com/v1/playlists/%s/tracks", playlistID)

	// if there's no more items to go through in response
	// nextURL will be empty string
	for nextURL != "" {
		resp, err := client.Get(nextURL)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch playlist info: %w", err)
		}
		defer resp.Body.Close()

		playlistResponse := &PlaylistResponse{}
		if err := json.NewDecoder(resp.Body).Decode(playlistResponse); err != nil {
			return nil, fmt.Errorf("failed to decode playlist info: %w", err)
		}

		for _, item := range playlistResponse.Items {
			track := item.Track
			if trackIDs, exists := nameToIDsMap[track.Name]; exists {
				nameToIDsMap[track.Name] = append(trackIDs, track.ID)
			} else {
				nameToIDsMap[track.Name] = []string{track.ID}
			}
		}

		nextURL = playlistResponse.Next
	}

	return nameToIDsMap, nil
}

func GetPlaylistIDs(client *http.Client, playlistID string) ([]string, error) {
	// "song name" : ["id1", "id2"]
	// there are duplicates of songs with different ids
	playlistIDs := []string{}
	nextURL := fmt.Sprintf("https://api.spotify.com/v1/playlists/%s/tracks", playlistID)

	// if there's no more items to go through in response
	// nextURL will be empty string
	for nextURL != "" {
		resp, err := client.Get(nextURL)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch playlist info: %w", err)
		}
		defer resp.Body.Close()

		playlistResponse := &PlaylistResponse{}
		if err := json.NewDecoder(resp.Body).Decode(playlistResponse); err != nil {
			return nil, fmt.Errorf("failed to decode playlist info: %w", err)
		}

		for _, item := range playlistResponse.Items {
			track := item.Track
			playlistIDs = append(playlistIDs, track.ID)
		}

		nextURL = playlistResponse.Next
	}

	return playlistIDs, nil
}

func UpdatePlaylistItems(client *http.Client, playlistID string) error {
	trackIDs := []string{}
	episodeIDs := []string{}
	reqURL := fmt.Sprintf("https://api.spotify.com/v1/playlists/%s/tracks", playlistID)

	// get user playlist
	playlist, err := GetPlaylistIDs(client, playlistID)
	if err != nil {
		return err
	}

	// compare if songs from user playlist in convert map
	convertMap, err := utils.LoadJSONtoMap("conversion_map.json")
	if err != nil {
		return fmt.Errorf("failed to load conversion map: %w", err)
	}

	// then, get corresponding episodeID to "IDsToAdd" from convert map
	for _, songID := range playlist {
		if episodeID, exists := convertMap[songID]; exists {
			trackIDs = append(trackIDs, songID)
			episodeIDs = append(episodeIDs, fmt.Sprintf(episodeID))
		}
	}

	// delete songs from user playlist
	for i := 0; i < len(trackIDs); i += 100 {
		// convert []string{} to []TrackURI{} to put inside req
		max_slice_size := min(i+100, len(trackIDs))
		trackURIs := make([]TrackURI, max_slice_size-i)
		for j, trackID := range trackIDs[i:max_slice_size] {
			trackURIs[j] = TrackURI{
				URI: fmt.Sprintf("spotify:track:%s", trackID),
			}
		}

		// add tracks to delete in unmarshalled req
		DeleteTracksRequest := DeleteTracksRequest{
			Items: trackURIs,
		}

		// marshal delete req
		jsonBody, err := json.Marshal(DeleteTracksRequest)
		if err != nil {
			return fmt.Errorf("failed to marshal delete request: %w", err)
		}

		req, err := http.NewRequest("DELETE", reqURL, bytes.NewBuffer(jsonBody))
		if err != nil {
			return fmt.Errorf("failed to create new delete request: %w", err)
		}

		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("failed to send delete request: %w", err)
		}
		defer resp.Body.Close()
	}

	// then add songs
	for i := 0; i < len(episodeIDs); i += 100 {
		// turn episodeIDs into URIs to put inside unmarshalled req
		max_slice_size := min(i+100, len(episodeIDs))
		PostTracksRequest := PostTracksRequest{
			URIs: []string{},
		}
		for _, episodeID := range episodeIDs[i:max_slice_size] {
			PostTracksRequest.URIs = append(PostTracksRequest.URIs,
				fmt.Sprintf("spotify:episode:%s", episodeID))
		}

		// marshal post req
		jsonBody, err := json.Marshal(PostTracksRequest)
		if err != nil {
			return fmt.Errorf("failed to marshal post request: %w", err)
		}

		req, err := http.NewRequest("POST", reqURL, bytes.NewBuffer(jsonBody))
		if err != nil {
			return fmt.Errorf("failed to create new post request: %w", err)
		}

		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("failed to send post request: %w", err)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("Error reading response body: %v", err)
			return err
		}

		log.Printf("%s\n", string(body))
	}

	return nil
}

func Bruh() error {
	episode_items, err := utils.LoadJSONtoMap("episode_items")
	if err != nil {
		return err
	}

	conversion_map, err := utils.LoadJSONtoMap("episode_items.json")
	if err != nil {
		return err
	}

	// get values of both
	episode_items_values := make([]string, 0, len(episode_items))
	for _, value := range episode_items {
		episode_items_values = append(episode_items_values, value)
	}

	conversion_item_values := make([]string, 0, len(conversion_map))
	for _, value := range conversion_map {
		conversion_item_values = append(conversion_item_values, value)
	}

	// Items in episode_items_values that are not in conversion_item_values
	diff1 := difference(episode_items_values, conversion_item_values)

	for _, item := range diff1 {
		log.Printf("%s", item)
	}

	return nil
}

func contains(slice []string, item string) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

func difference(slice1, slice2 []string) []string {
	diff := make([]string, 0)
	for _, item := range slice1 {
		if !contains(slice2, item) {
			diff = append(diff, item)
		}
	}
	return diff
}
