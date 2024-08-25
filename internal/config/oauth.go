package config

import (
	"encoding/gob"
	"log"
	"os"

	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/endpoints"
)

var SpotifyOAuthConfig *oauth2.Config

func init() {
	gob.Register(&oauth2.Token{})

	err := godotenv.Load(".env")
	if err != nil {
		log.Printf("Error loading .env file: %v", err)
	}

	SpotifyOAuthConfig = &oauth2.Config{
		ClientID:     os.Getenv("SPOTIFY_CLIENT_ID"),
		ClientSecret: os.Getenv("SPOTIFY_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("SPOTIFY_REDIRECT_URL"),
		Scopes: []string{
			"playlist-modify-public",
			"playlist-modify-private",
		},
		Endpoint: endpoints.Spotify,
	}
}
