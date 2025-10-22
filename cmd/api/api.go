package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/MohummedSoliman/social/internal/auth"
	"github.com/MohummedSoliman/social/internal/mailer"
	"github.com/MohummedSoliman/social/internal/ratelimiter"
	"github.com/MohummedSoliman/social/internal/store"
	"github.com/MohummedSoliman/social/internal/store/cache"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type application struct {
	config        config
	store         store.Storage
	db            dbConfig
	mailer        mailer.Client
	authenticator auth.Authenticator
	cacheStore    cache.Storage
	ratelimiter   ratelimiter.Limiter
}

type config struct {
	addr        string
	mail        mailConfig
	frontendURL string
	auth        authConfig
	redisConfig redisConfig
	rateLimiter ratelimiter.Config
}

type redisConfig struct {
	addr     string
	password string
	db       int
	enabled  bool
}

type authConfig struct {
	basic basicConfig
	token tokenConfig
}

type tokenConfig struct {
	secret string
	exp    time.Duration
}

type basicConfig struct {
	user string
	pass string
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
	mux.Use(app.RateLimiterMiddleware)

	mux.Use(middleware.Timeout(60 * time.Second))

	mux.Route("/v1", func(r chi.Router) {
		// r.With(app.BasicAuthMiddleware()).Get("/health", app.healthCheckHandler)
		r.Get("/health", app.healthCheckHandler)

		r.Route("/users", func(r chi.Router) {
			r.Put("/activate/{token}", app.activateUserHandler)

			r.Route("/{userID}", func(r chi.Router) {
				r.Use(app.AuthTokenMiddleware())

				r.Get("/", app.getUserHandler)
				r.Put("/follow", app.followUserHandler)
				r.Put("/unfollow", app.unFollowUserHandler)
			})

			r.Group(func(r chi.Router) {
				r.Use(app.AuthTokenMiddleware())
				r.Get("/feed", app.getUserFeedhandler)
			})
		})

		r.Route("/authentication", func(r chi.Router) {
			r.Post("/user", app.registerUserHandler)
			r.Post("/token", app.createTokenHandler)
		})

		r.Route("/posts", func(r chi.Router) {
			r.Use(app.AuthTokenMiddleware())

			r.Post("/", app.createPostHandler)

			r.Route("/{postID}", func(r chi.Router) {
				r.Get("/", app.getPostHandler)
				r.Delete("/", app.checkPostOwnership("admin", app.deletePostHandler))
				r.Patch("/", app.checkPostOwnership("moderator", app.updatePostHandler))
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

	shutdown := make(chan error)

	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		s := <-quit

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		log.Println("Signal Caught: ", s.String())

		shutdown <- srv.Shutdown(ctx)
	}()

	log.Println("Server has started at: ", app.config.addr)

	err := srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	err = <-shutdown
	if err != nil {
		return err
	}

	log.Println("Server has stopped")
	return nil
}
