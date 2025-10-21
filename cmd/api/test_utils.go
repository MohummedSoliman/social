package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MohummedSoliman/social/internal/auth"
	"github.com/MohummedSoliman/social/internal/store"
	"github.com/MohummedSoliman/social/internal/store/cache"
)

func newTestApplication(t *testing.T) *application {
	t.Helper()
	mockStore := store.NewMockStore()
	mockCacheStore := cache.NewMockCacheStorage()
	testAuth := auth.NewTestAuthenticator()

	return &application{
		store:         mockStore,
		cacheStore:    mockCacheStore,
		authenticator: testAuth,
	}
}

func executeRequest(r *http.Request, mux http.Handler) *httptest.ResponseRecorder {
	reqRec := httptest.NewRecorder()
	mux.ServeHTTP(reqRec, r)
	return reqRec
}

func checkResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected Response code %d, but got %d", expected, actual)
	}
}
