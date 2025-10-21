package main

import (
	"net/http"
	"testing"
)

func TestGetUser(t *testing.T) {
	app := newTestApplication(t)
	mux := app.mount()
	testToken, _ := app.authenticator.GenerateToken(nil)
	t.Run("Should not allow unauthenticated requests", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "/v1/users/1", nil)
		if err != nil {
			t.Fatal(err)
		}

		reqRec := executeRequest(req, mux)
		checkResponseCode(t, http.StatusUnauthorized, reqRec.Code)
	})

	t.Run("Should allow authenticated requests", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "/v1/users/1", nil)
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Authorization", "Bearer "+testToken)
		reqRec := executeRequest(req, mux)
		checkResponseCode(t, http.StatusOK, reqRec.Code)
	})
}
