package handlers

import (
	"errors"
	"net/http"

	"github.com/cloudfoundry/dropsonde/autowire"
	"github.com/goji/httpauth"
	"github.com/pivotal-golang/lager"
)

func LogWrap(handler http.Handler, logger lager.Logger) http.HandlerFunc {
	handler = autowire.InstrumentedHandler(handler)

	return func(w http.ResponseWriter, r *http.Request) {
		requestLog := logger.Session("request", lager.Data{
			"method":  r.Method,
			"request": r.URL.String(),
		})

		requestLog.Info("serving")
		handler.ServeHTTP(w, r)
		requestLog.Info("done")
	}
}

func BasicAuthWrap(handler http.Handler, username, password string) http.Handler {
	opts := httpauth.AuthOptions{
		Realm:               "API Authentication",
		User:                username,
		Password:            password,
		UnauthorizedHandler: http.HandlerFunc(unauthorized),
	}
	return httpauth.BasicAuth(opts)(handler)
}

func unauthorized(w http.ResponseWriter, r *http.Request) {
	status := http.StatusUnauthorized
	err := errors.New(http.StatusText(status))
	writeErrorResponse(w, status, err)
}
