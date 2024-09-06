package youtube

import (
	"context"
	"net/http"

	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

func Test(client *http.Client) (interface{}, error) {
	service, err := youtube.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		return nil, err
	}

	channelCall := service.Playlists.List([]string{"snippet", "contentDetails"}).Mine(true)
	channelResponse, err := channelCall.Do()
	if err != nil {
		return nil, err
	}

	playlistStuff := []string{}

	for _, playlist := range channelResponse.Items {
		playlistStuff = append(playlistStuff, playlist.Id)
	}

	return playlistStuff, nil
}
