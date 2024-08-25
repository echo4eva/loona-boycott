package main

import "net/http"

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /{$}", app.helloWorld)
	mux.HandleFunc("GET /auth/callback", app.callback)
	mux.HandleFunc("GET /auth/login", app.login)
	mux.HandleFunc("GET /me", app.me)
	// mux.HandleFunc("GET /playlist", app.getConversionMap)
	// mux.HandleFunc("GET /test", app.test)
	mux.HandleFunc("GET /replace", app.replace)
	mux.HandleFunc("GET /bruh", app.bruh)

	return app.sessionManager.LoadAndSave(mux)
}
