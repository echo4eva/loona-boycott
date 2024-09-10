package main

type contextKey string

const (
	isSpotifyAuthenticatedContextKey = contextKey("isSpotifyAuthenticated")
	isYoutubeAuthenticatedContextKey = contextKey("isYoutubeAuthenticated")
)
