package youtube

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

func GetPlaylistIDs(client *http.Client) (interface{}, error) {
	service, err := youtube.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		return nil, err
	}

	channelCall := service.Playlists.List([]string{"snippet"}).Mine(true).MaxResults(50)
	channelResponse, err := channelCall.Do()
	if err != nil {
		return nil, err
	}

	playlistStuff := []string{}

	for _, playlist := range channelResponse.Items {
		println(playlist.Snippet.Title)
		println(playlist.Id)
	}

	return playlistStuff, nil
}

func GetPlaylistItems(client *http.Client, playlistID string, jsonName string) (map[string][]string, error) {
	playlistItems := make(map[string][]string)

	service, err := youtube.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		return nil, err
	}

	nextPageToken := ""
	for {
		call := service.PlaylistItems.List([]string{"snippet"}).
			PlaylistId(playlistID).
			MaxResults(50).
			PageToken(nextPageToken)

		response, err := call.Do()
		if err != nil {
			return nil, err
		}

		for _, item := range response.Items {
			videoTitle := item.Snippet.Title
			videoID := item.Snippet.ResourceId.VideoId
			if _, exists := playlistItems[videoTitle]; !exists {
				playlistItems[videoTitle] = []string{videoID}
			} else {
				playlistItems[videoTitle] = append(playlistItems[videoTitle], videoID)
			}
		}

		nextPageToken = response.NextPageToken
		if nextPageToken == "" {
			break
		}
	}

	jsonData, err := json.MarshalIndent(playlistItems, "", " ")
	if err != nil {
		return nil, err
	}

	err = os.WriteFile(fmt.Sprintf("%s.json", jsonName), jsonData, 0644)
	if err != nil {
		return nil, err
	}

	return playlistItems, nil
}
