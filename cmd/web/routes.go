package main

import "net/http"

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()

	fileServer := http.FileServer(http.Dir("./ui/static/"))
	mux.Handle("GET /static/", http.StripPrefix("/static", fileServer))

	mux.HandleFunc("GET /{$}", app.helloWorld)
	mux.HandleFunc("GET /auth/callback", app.callback)
	mux.HandleFunc("GET /auth/yt/callback", app.youtubeCallback)
	mux.HandleFunc("GET /auth/login", app.login)
	mux.HandleFunc("GET /auth/yt/login", app.youtubeLogin)
	mux.HandleFunc("GET /me", app.me)
	// mux.HandleFunc("GET /playlist", app.getConversionMap)
	// mux.HandleFunc("GET /test", app.test)
	mux.HandleFunc("POST /replace", app.replacePost)

	return app.sessionManager.LoadAndSave(app.authenticate(mux))
}
