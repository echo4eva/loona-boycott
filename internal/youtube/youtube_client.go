package youtube

import (
	"context"
	"echo4eva/loona/internal/utils"
	"encoding/json"
	"fmt"
	"log"
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

func GetUserPlaylistItems(client *http.Client, userPlaylistID string) (map[string][]string, error) {
	// id != videoId
	// id is the video associated with the playlist
	// videoId is the unique id of the video
	songTitleAndIDs := make(map[string][]string)

	service, err := youtube.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		return nil, err
	}

	nextPageToken := ""
	for {
		call := service.PlaylistItems.List([]string{"snippet"}).
			PlaylistId(userPlaylistID).
			MaxResults(50).
			PageToken(nextPageToken)

		response, err := call.Do()
		if err != nil {
			return nil, err
		}

		for _, item := range response.Items {
			title := item.Snippet.Title
			id := item.Id

			if _, exists := songTitleAndIDs[title]; !exists {
				songTitleAndIDs[title] = []string{id}
			} else {
				songTitleAndIDs[title] = append(songTitleAndIDs[title], id)
			}
		}

		nextPageToken = response.NextPageToken
		if nextPageToken == "" {
			break
		}
	}

	return songTitleAndIDs, nil
}

func UpdatePlaylistItems(client *http.Client, userPlaylistID string) error {
	// conversion map was manually created and edited
	convertMap, err := utils.LoadJSONtoMap("yt_conversion_map.json")
	if err != nil {
		return fmt.Errorf("failed to load conversion map: %w", err)
	}
	boycottMap, err := utils.LoadJSONtoMap("yt_playlist_boycott_items.json")
	if err != nil {
		return fmt.Errorf("failed to load boycott map: %w", err)
	}
	officialSongIDsToDelete := []string{}
	boycottSongIDsToAdd := []string{}

	service, err := youtube.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		return err
	}

	userPlaylistItemTitles, err := GetUserPlaylistItems(client, userPlaylistID)
	if err != nil {
		return err
	}

	for title, ids := range userPlaylistItemTitles {
		// if this song title exists in the conversion map
		if boycottTitle, exists := convertMap[title]; exists {
			// get the song's id to delete
			officialSongIDsToDelete = append(officialSongIDsToDelete, ids...)
			// get the corresponding boycott IDs to add
			if boycottSongID, exists := boycottMap[boycottTitle]; exists {
				boycottSongIDsToAdd = append(boycottSongIDsToAdd, boycottSongID)
			}
		}
	}

	// deletes the official songs
	for _, id := range officialSongIDsToDelete {
		deleteCall := service.PlaylistItems.Delete(id)
		err := deleteCall.Do()
		if err != nil {
			return err
		}
	}

	// insert the boycott songs
	for _, id := range boycottSongIDsToAdd {
		// creates snippet to insert into playlist
		itemSnippet := &youtube.PlaylistItem{
			Snippet: &youtube.PlaylistItemSnippet{
				PlaylistId: userPlaylistID,
				ResourceId: &youtube.ResourceId{
					Kind:    "youtube#video",
					VideoId: id,
				},
			},
		}

		insertCall := service.PlaylistItems.Insert([]string{"snippet"}, itemSnippet)
		response, err := insertCall.Do()
		if err != nil {
			return err
		}

		log.Printf("Inserted video: %s\n", response.Snippet.Title)
	}

	return nil
}

func GetBoycottPlaylistItems(client *http.Client, playlistID string, jsonName string) (map[string]string, error) {
	boycottItems := make(map[string]string)

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
			if _, exists := boycottItems[videoTitle]; !exists {
				boycottItems[videoTitle] = videoID
			}
		}

		nextPageToken = response.NextPageToken
		if nextPageToken == "" {
			break
		}
	}

	jsonData, err := json.MarshalIndent(boycottItems, "", " ")
	if err != nil {
		return nil, err
	}

	err = os.WriteFile(fmt.Sprintf("data/%s.json", jsonName), jsonData, 0644)
	if err != nil {
		return nil, err
	}

	return boycottItems, nil
}

func GetOfficialPlaylistItems(client *http.Client, playlistID string, jsonName string) (map[string][]string, error) {
	officialItems := make(map[string][]string)

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
			if _, exists := officialItems[videoTitle]; !exists {
				officialItems[videoTitle] = []string{videoID}
			} else {
				officialItems[videoTitle] = append(officialItems[videoTitle], videoID)
			}
		}

		nextPageToken = response.NextPageToken
		if nextPageToken == "" {
			break
		}
	}

	jsonData, err := json.MarshalIndent(officialItems, "", " ")
	if err != nil {
		return nil, err
	}

	err = os.WriteFile(fmt.Sprintf("data/%s.json", jsonName), jsonData, 0644)
	if err != nil {
		return nil, err
	}

	return officialItems, nil
}
