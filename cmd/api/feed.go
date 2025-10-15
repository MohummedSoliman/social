package main

import "net/http"

func (app *application) getUserFeedhandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	feeds, err := app.store.Posts.GetUserFeed(ctx, int64(1))
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := jsonResponse(w, http.StatusOK, feeds); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}
