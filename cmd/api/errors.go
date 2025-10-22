package main

import (
	"log"
	"net/http"
)

func (app *application) internalServerError(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("internal server error : %s path: %s error: %s", r.Method, r.URL.Path, err)
	writeJSONError(w, http.StatusInternalServerError, "Server Encounter a Problem!!")
}

func (app *application) badRequest(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("bad request error : %s path: %s error: %s", r.Method, r.URL.Path, err)
	writeJSONError(w, http.StatusBadRequest, "Bad Request Data")
}

func (app *application) notFoundError(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("not found error : %s path: %s error: %s", r.Method, r.URL.Path, err)
	writeJSONError(w, http.StatusNotFound, "Bad Request Data")
}

func (app *application) unauthorizedError(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("UnAuthorized Error : %s path: %s error: %s", r.Method, r.URL.Path, err)
	writeJSONError(w, http.StatusUnauthorized, "unauthorized")
}

func (app *application) unauthorizedBasicError(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("UnAuthorized Basic Error :  %s path: %s error: %s", r.Method, r.URL.Path, err)
	w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
	writeJSONError(w, http.StatusUnauthorized, "unauthorized")
}

func (app *application) forbiddenResponse(w http.ResponseWriter, r *http.Request) {
	log.Printf("Forbidden Error :  %s path: %s ", r.Method, r.URL.Path)
	writeJSONError(w, http.StatusForbidden, "Forbidden")
}

func (app *application) rateLimiterExceededResponse(w http.ResponseWriter, r *http.Request, retryAfter string) {
	log.Printf("Rate Limiter Exceeded for path %s", r.URL.Path)
	w.Header().Set("Retry-After", retryAfter)
	writeJSONError(w, http.StatusTooManyRequests, "rate limit exceeded, retry after: "+retryAfter)
}
