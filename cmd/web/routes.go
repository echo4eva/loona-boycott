package main

import "net/http"

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()

	fileServer := http.FileServer(http.Dir("./ui/static/"))
	mux.Handle("GET /static/", http.StripPrefix("/static", fileServer))

	mux.HandleFunc("GET /{$}", app.homePage)
	mux.HandleFunc("GET /privacy-policy", app.privacyPage)
	mux.HandleFunc("GET /auth/spotify/callback", app.spotifyCallback)
	mux.HandleFunc("GET /auth/spotify/login", app.spotifyLogin)
	mux.HandleFunc("GET /auth/spotify/logout", app.spotifyLogout)
	mux.HandleFunc("GET /auth/youtube/callback", app.youtubeCallback)
	mux.HandleFunc("GET /auth/youtube/login", app.youtubeLogin)
	mux.HandleFunc("GET /auth/youtube/logout", app.youtubeLogout)
	// mux.HandleFunc("GET /playlist", app.getSpotifyConversionMap)
	mux.HandleFunc("POST /replace/spotify", app.spotifyReplacePost)
	mux.HandleFunc("POST /replace/youtube", app.youtubeReplacePost)
	// mux.HandleFunc("GET /dev/youtube/update", app.youtubeUpdateJSONItems)

	return app.sessionManager.LoadAndSave(app.authenticate(commonHeaders(mux)))
}
