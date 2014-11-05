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
	desiredLRPHandler := NewDesiredLRPHandler(bbs, logger)

	actions := rata.Handlers{
		// Tasks
		receptor.CreateTaskRoute:          route(taskHandler.Create),
		receptor.GetAllTasksRoute:         route(taskHandler.GetAll),
		receptor.GetAllTasksByDomainRoute: route(taskHandler.GetAllByDomain),
		receptor.GetTaskRoute:             route(taskHandler.GetByGuid),
		receptor.DeleteTaskRoute:          route(taskHandler.Delete),

		// DesiredLRPs
		receptor.CreateDesiredLRPRoute:           route(desiredLRPHandler.Create),
		receptor.GetDesiredLRPByProcessGuidRoute: route(desiredLRPHandler.GetByProcessGuid),
		receptor.UpdateDesiredLRPRoute:           route(desiredLRPHandler.Update),
		receptor.DeleteDesiredLRPRoute:           route(desiredLRPHandler.Delete),
		receptor.GetAllDesiredLRPsRoute:          route(desiredLRPHandler.GetAll),
		receptor.GetAllDesiredLRPsByDomainRoute:  route(desiredLRPHandler.GetAllByDomain),
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
