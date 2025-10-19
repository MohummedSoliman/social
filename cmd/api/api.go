package main

import (
	"log"
	"net/http"
	"time"

	"github.com/MohummedSoliman/social/internal/mailer"
	"github.com/MohummedSoliman/social/internal/store"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type application struct {
	config config
	store  store.Storage
	db     dbConfig
	mailer mailer.Client
}

type config struct {
	addr        string
	mail        mailConfig
	frontendURL string
}

type mailConfig struct {
	expiry   time.Duration
	sendGrid sendGridConfig
}

type sendGridConfig struct {
	apiKey    string
	fromEmail string
}

type dbConfig struct {
	addr         string
	maxOpenConns int
	maxIdleConns int
	maxIdleTime  string
	env          string
}

func (app *application) mount() http.Handler {
	mux := chi.NewRouter()

	mux.Use(middleware.RealIP)
	mux.Use(middleware.RequestID)
	mux.Use(middleware.Recoverer)
	mux.Use(middleware.Logger)

	mux.Use(middleware.Timeout(60 * time.Second))

	mux.Route("/v1", func(r chi.Router) {
		r.Get("/health", app.healthCheckHandler)

		r.Route("/users", func(r chi.Router) {
			r.Put("/activate/{token}", app.activateUserHandler)

			r.Route("/{userID}", func(r chi.Router) {
				r.Use(app.userContextMiddleware)
				r.Get("/", app.getUserHandler)
				r.Put("/follow", app.followUserHandler)
				r.Put("/unfollow", app.unFollowUserHandler)
			})

			r.Group(func(r chi.Router) {
				r.Get("/feed", app.getUserFeedhandler)
			})
		})

		r.Route("/authentication", func(r chi.Router) {
			r.Post("/user", app.registerUserHandler)
		})

		r.Route("/posts", func(r chi.Router) {
			r.Post("/", app.createPostHandler)

			r.Route("/{postID}", func(r chi.Router) {
				r.Use(app.postContextMiddleware)
				r.Get("/", app.getPostHandler)
				r.Delete("/", app.deletePostHandler)
				r.Patch("/", app.updatePostHandler)
			})
		})
	})

	return mux
}

func (app *application) run(mux http.Handler) error {
	srv := &http.Server{
		Addr:         app.config.addr,
		Handler:      mux,
		WriteTimeout: time.Second * 30,
		ReadTimeout:  time.Second * 10,
		IdleTimeout:  time.Minute,
	}

	log.Println("Server has started at: ", app.config.addr)

	return srv.ListenAndServe()
}
