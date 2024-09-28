package main

import (
	"context"
	"net/http"
)

func (app *application) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// "spotify_token" is type oauth2.Token
		spotify_token := app.sessionManager.Get(r.Context(), "spotify_token")
		youtube_token := app.sessionManager.Get(r.Context(), "youtube_token")

		ctx := r.Context()
		if spotify_token != nil {
			ctx = context.WithValue(ctx, isSpotifyAuthenticatedContextKey, true)
			ctx = context.WithValue(ctx, isYoutubeAuthenticatedContextKey, false)
		} else if youtube_token != nil {
			ctx = context.WithValue(ctx, isYoutubeAuthenticatedContextKey, true)
			ctx = context.WithValue(ctx, isSpotifyAuthenticatedContextKey, false)
		} else {
			ctx = context.WithValue(ctx, isYoutubeAuthenticatedContextKey, false)
			ctx = context.WithValue(ctx, isSpotifyAuthenticatedContextKey, false)
		}
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

func commonHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Security-Policy",
			"default-src 'self'; style-src 'self' https://fonts.googleapis.com; font-src 'self' https://fonts.gstatic.com;")
		w.Header().Set("Referrer-Policy", "origin-when-cross-origin")
		w.Header().Set("X-Content-Tyoe-Options", "no-sniff")
		w.Header().Set("X-Frame-Options", "deny")
		w.Header().Set("X-XSS-Protection", "0")

		w.Header().Set("Server", "Go")
		w.Header().Add("Cache-Control", "no-store")

		next.ServeHTTP(w, r)
	})
}
