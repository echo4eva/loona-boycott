package main

import "net/http"

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()

	fileServer := http.FileServer(http.Dir("./ui/static/"))
	mux.Handle("GET /static/", http.StripPrefix("/static", fileServer))

	mux.HandleFunc("GET /{$}", app.helloWorld)
	mux.HandleFunc("GET /auth/spotify/callback", app.spotifyCallback)
	mux.HandleFunc("GET /auth/spotify/login", app.spotifyLogin)
	mux.HandleFunc("GET /auth/spotify/logout", app.spotifyLogout)
	mux.HandleFunc("GET /auth/youtube/callback", app.youtubeCallback)
	mux.HandleFunc("GET /auth/youtube/login", app.youtubeLogin)
	mux.HandleFunc("GET /auth/youtube/logout", app.youtubeLogout)
	// mux.HandleFunc("GET /me", app.me)
	mux.HandleFunc("GET /playlist", app.getConversionMap)
	// mux.HandleFunc("GET /test", app.test)
	mux.HandleFunc("POST /replace/spotify", app.spotifyReplacePost)
	// mux.HandleFunc("GET /test", app.youtubeTest)
	// mux.HandleFunc("GET /test2", app.youtubeTest2)
	mux.HandleFunc("POST /replace/youtube", app.youtubeReplacePost)

	return app.sessionManager.LoadAndSave(app.authenticate(commonHeaders(mux)))
}
