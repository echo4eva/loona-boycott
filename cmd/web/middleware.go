package main

import (
	"context"
	"net/http"
)

func (app *application) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// "spotify_token" is type oauth2.Token
		spotify_token := app.sessionManager.Get(r.Context(), "spotify_token")

		if spotify_token != nil {
			ctx := context.WithValue(r.Context(), isAuthenticatedContextKey, true)
			r = r.WithContext(ctx)
		}

		next.ServeHTTP(w, r)
	})
}
