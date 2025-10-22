package main

import (
	"net/http"
	"strconv"

	"github.com/MohummedSoliman/social/internal/store"
	"github.com/go-chi/chi/v5"
)

type userContextKeys string

var USERKEY userContextKeys = "user"

func (app *application) getUserHandler(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r)

	if err := jsonResponse(w, http.StatusOK, user); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

func (app *application) followUserHandler(w http.ResponseWriter, r *http.Request) {
	followerUser := getUserFromContext(r)
	followedID, err := strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64)
	if err != nil {
		app.badRequest(w, r, err)
		return
	}

	err = app.store.Followers.Follow(r.Context(), followerUser.ID, followedID)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := jsonResponse(w, http.StatusNoContent, nil); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

func (app *application) unFollowUserHandler(w http.ResponseWriter, r *http.Request) {
	unfollowedUser := getUserFromContext(r)

	unfollowedID, err := strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64)
	if err != nil {
		app.badRequest(w, r, err)
		return
	}

	err = app.store.Followers.UnFollow(r.Context(), unfollowedUser.ID, unfollowedID)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := jsonResponse(w, http.StatusNoContent, nil); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

// func (app *application) userContextMiddleware(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		idParam := chi.URLParam(r, "userID")
// 		userID, err := strconv.Atoi(idParam)
// 		if err != nil {
// 			app.badRequest(w, r, err)
// 			return
// 		}

// 		user, err := app.store.Users.GetUserByID(r.Context(), int64(userID))
// 		if err != nil {
// 			switch {
// 			case errors.Is(err, store.ErrNotFound):
// 				app.badRequest(w, r, err)
// 			default:
// 				app.internalServerError(w, r, err)
// 			}
// 			return
// 		}

// 		ctx := context.WithValue(r.Context(), USERKEY, user)
// 		next.ServeHTTP(w, r.WithContext(ctx))
// 	})
// }

func getUserFromContext(r *http.Request) *store.User {
	user := r.Context().Value(USERKEY).(*store.User)
	return user
}
