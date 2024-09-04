package main

import (
	"echo4eva/loona/internal/config"
	"html/template"
	"log/slog"
	"net/http"
	"os"

	"github.com/alexedwards/scs/redisstore"
	"github.com/alexedwards/scs/v2"
	"github.com/gomodule/redigo/redis"
	"golang.org/x/oauth2"
)

type application struct {
	oAuth          *oauth2.Config
	sessionManager *scs.SessionManager
	logger         *slog.Logger
	spotifyClient  *http.Client
	templateCache  map[string]*template.Template
}

func main() {
	// init connection pool to redis
	pool := &redis.Pool{
		MaxIdle: 10,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", "redis:6379")
		},
	}

	// init session manager
	sessionManager := scs.New()
	sessionManager.Store = redisstore.New(pool)

	// init logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// init template cache
	templateCache, err := newTemplateCache()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	app := &application{
		oAuth:          config.SpotifyOAuthConfig,
		sessionManager: sessionManager,
		logger:         logger,
		spotifyClient:  &http.Client{},
		templateCache:  templateCache,
	}

	// init server
	server := &http.Server{
		Addr: "0.0.0.0:9001",
		// first handler == main router
		Handler:  app.routes(),
		ErrorLog: slog.NewLogLogger(logger.Handler(), slog.LevelError),
	}

	// starts server
	logger.Info("starting server")
	err = server.ListenAndServe()
	logger.Error(err.Error())
	os.Exit(1)
}
