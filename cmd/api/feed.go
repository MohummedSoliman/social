package main

import (
	"net/http"

	"github.com/MohummedSoliman/social/internal/store"
)

func (app *application) getUserFeedhandler(w http.ResponseWriter, r *http.Request) {
	fq := store.PaginatedFeedQuery{
		Limit:  20,
		Offset: 0,
		Sort:   "desc",
	}

	fq, err := fq.Parse(r)
	if err != nil {
		app.badRequest(w, r, err)
		return
	}

	err = Validate.Struct(fq)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	ctx := r.Context()

	feeds, err := app.store.Posts.GetUserFeed(ctx, int64(1), fq)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := jsonResponse(w, http.StatusOK, feeds); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}
