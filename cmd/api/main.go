package main

import (
	"log"
	"time"

	"github.com/MohummedSoliman/social/internal/auth"
	"github.com/MohummedSoliman/social/internal/db"
	"github.com/MohummedSoliman/social/internal/env"
	"github.com/MohummedSoliman/social/internal/mailer"
	"github.com/MohummedSoliman/social/internal/ratelimiter"
	"github.com/MohummedSoliman/social/internal/store"
	"github.com/MohummedSoliman/social/internal/store/cache"
	"github.com/go-redis/redis/v8"
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

	redisConfig := redisConfig{
		addr:     env.GetString("REDIS_ADDR", "localhost:6379"),
		password: env.GetString("REDIS_PASS", ""),
		db:       env.GetInt("REDIS_DB", 0),
		enabled:  env.GetBool("REDIS_ENABLED", true),
	}

	db, err := db.New(cfg.addr, cfg.maxOpenConns, cfg.maxIdleConns, cfg.maxIdleTime)
	if err != nil {
		log.Panic(err)
	}
	defer db.Close()
	log.Println("postgres connection pool establish sucessfully!!")

	var rdsDB *redis.Client
	if redisConfig.enabled {
		rdsDB = cache.NewRedisClient(redisConfig.addr, redisConfig.password, redisConfig.db)
		log.Println("redis connection establish successfully")
	}

	rateLimiterCfg := ratelimiter.Config{
		RequestsPerTimeFrame: env.GetInt("RATELIMITER_REQUESTS_COUNT", 20),
		TimeFrame:            time.Second * 5,
		Enabled:              env.GetBool("RATELIMITER_ENABLED", true),
	}

	ratelimiter := ratelimiter.NewFixedWindowLimiter(
		rateLimiterCfg.RequestsPerTimeFrame,
		rateLimiterCfg.TimeFrame,
	)

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
			redisConfig: redisConfig,
			rateLimiter: rateLimiterCfg,
		},
		db:            cfg,
		store:         store,
		cacheStore:    cache.NewRedisStorage(rdsDB),
		mailer:        mailer,
		authenticator: jwtAuthenticator,
		ratelimiter:   ratelimiter,
	}

	mux := app.mount()
	log.Fatal(app.run(mux))
}
