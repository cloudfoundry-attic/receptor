package handlers

import (
	"net/http"

	"github.com/cloudfoundry-incubator/receptor"
	Bbs "github.com/cloudfoundry-incubator/runtime-schema/bbs"
	"github.com/pivotal-golang/lager"
	"github.com/tedsuo/rata"
)

func New(bbs Bbs.ReceptorBBS, logger lager.Logger, username, password string) http.Handler {
	taskHandler := NewTaskHandler(bbs, logger)

	actions := rata.Handlers{
		receptor.CreateTask:          route(taskHandler.Create),
		receptor.GetAllTasks:         route(taskHandler.GetAll),
		receptor.GetAllTasksByDomain: route(taskHandler.GetAllByDomain),
		receptor.GetTask:             route(taskHandler.GetByGuid),
		receptor.DeleteTask:          route(taskHandler.Delete),
	}

	handler, err := rata.NewRouter(receptor.Routes, actions)
	if err != nil {
		panic("unable to create router: " + err.Error())
	}

	if username != "" {
		handler = BasicAuthWrap(handler, username, password)
	}

	handler = LogWrap(handler, logger)

	return handler
}

func route(f func(w http.ResponseWriter, r *http.Request)) http.Handler {
	return http.HandlerFunc(f)
}
