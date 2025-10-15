package main

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/MohummedSoliman/social/internal/store"
	"github.com/go-chi/chi/v5"
)

func (app *application) getUserHandler(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "userID")
	userID, err := strconv.Atoi(idParam)
	if err != nil {
		app.badRequest(w, r, err)
		return
	}

	user, err := app.store.Users.GetUserByID(r.Context(), int64(userID))
	if err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			app.badRequest(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	if err := jsonResponse(w, http.StatusOK, user); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}
