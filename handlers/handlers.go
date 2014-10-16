package handlers

import (
	"net/http"

	"github.com/cloudfoundry-incubator/receptor/api"
	"github.com/pivotal-golang/lager"
	"github.com/tedsuo/rata"
)

func New(logger lager.Logger) http.Handler {
	routes, err := rata.NewRouter(api.Routes, rata.Handlers{
		api.CreateTask: newCreateTaskHandler(logger),
	})
	if err != nil {
		panic("unable to create router: " + err.Error())
	}
	return logWrap(routes, logger)
}

func logWrap(handler http.Handler, logger lager.Logger) http.HandlerFunc {
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
