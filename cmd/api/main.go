package main

import (
	"log"

	"github.com/MohummedSoliman/social/internal/db"
	"github.com/MohummedSoliman/social/internal/env"
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

	db, err := db.New(cfg.addr, cfg.maxOpenConns, cfg.maxIdleConns, cfg.maxIdleTime)
	if err != nil {
		log.Panic(err)
	}
	defer db.Close()
	log.Println("postgres connection pool establish sucessfully!!")

	store := store.NewStorage(db)

	app := &application{
		config: config{
			addr: env.GetString("ADDR", ":8080"),
		},
		db:    cfg,
		store: store,
	}

	mux := app.mount()
	log.Fatal(app.run(mux))
}
