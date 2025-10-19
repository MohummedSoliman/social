package main

import (
	"log"
	"time"

	"github.com/MohummedSoliman/social/internal/auth"
	"github.com/MohummedSoliman/social/internal/db"
	"github.com/MohummedSoliman/social/internal/env"
	"github.com/MohummedSoliman/social/internal/mailer"
	"github.com/MohummedSoliman/social/internal/store"
)

const version = "0.0.1"

func main() {
	cfg := dbConfig{
		addr:         env.GetString("DB_ADDR", "postgres://admin:adminpassword@localhost/social?sslmode=disable"),
		maxOpenConns: env.GetInt("DB_MAX_OPEN_CONNS", 30),
		maxIdleConns: env.GetInt("DB_MAX_IDLE_CONNS", 30),
		maxIdleTime:  env.GetString("DB_MAX_IDLE_TIME", "15m"),
		env:          env.GetString("ENV", "Development"),
	}

	mailCfg := mailConfig{
		expiry: time.Hour * 24 * 3,
		sendGrid: sendGridConfig{
			apiKey:    env.GetString("SENDGRID_API_KEY", ""),
			fromEmail: env.GetString("SENDGRID_FROM_EMAIL", ""),
		},
	}

	db, err := db.New(cfg.addr, cfg.maxOpenConns, cfg.maxIdleConns, cfg.maxIdleTime)
	if err != nil {
		log.Panic(err)
	}
	defer db.Close()
	log.Println("postgres connection pool establish sucessfully!!")

	store := store.NewStorage(db)

	mailer := mailer.NewSendgrid(mailCfg.sendGrid.apiKey, mailCfg.sendGrid.fromEmail)

	token := tokenConfig{
		secret: env.GetString("JWT_TOKEN_SECRET", ""),
		exp:    time.Hour * 24 * 3,
	}

	jwtAuthenticator := auth.NewJWTAuthenticator(
		token.secret, "GopherSocial", "GopherSocial",
	)

	app := &application{
		config: config{
			addr:        env.GetString("ADDR", ":8080"),
			mail:        mailCfg,
			frontendURL: env.GetString("FRONTEND_URL", "http://localhost:4000"),
			auth: authConfig{
				basic: basicConfig{
					user: env.GetString("AUTH_BASIC_USER", "admin"),
					pass: env.GetString("AUTH_BASIC_PASS", "admin"),
				},
				token: token,
			},
		},
		db:            cfg,
		store:         store,
		mailer:        mailer,
		authenticator: jwtAuthenticator,
	}

	mux := app.mount()
	log.Fatal(app.run(mux))
}
