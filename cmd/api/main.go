package main

import (
	"log"

	"github.com/MohummedSoliman/social/internal/env"
)

func main() {
	app := &application{
		config: config{
			addr: env.GetString("ADDR", ":8080"),
		},
	}

	mux := app.mount()
	log.Fatal(app.run(mux))
}
