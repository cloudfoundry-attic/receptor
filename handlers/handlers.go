package handlers

import (
	"net/http"

	"github.com/cloudfoundry-incubator/receptor"
	Bbs "github.com/cloudfoundry-incubator/runtime-schema/bbs"
	"github.com/pivotal-golang/lager"
	"github.com/tedsuo/rata"
)

func New(bbs Bbs.ReceptorBBS, logger lager.Logger, username, password string, corsEnabled bool) http.Handler {
	taskHandler := NewTaskHandler(bbs, logger)
	desiredLRPHandler := NewDesiredLRPHandler(bbs, logger)
	actualLRPHandler := NewActualLRPHandler(bbs, logger)
	cellHandler := NewCellHandler(bbs, logger)
	freshDomainHandler := NewFreshDomainHandler(bbs, logger)

	actions := rata.Handlers{
		// Tasks
		receptor.CreateTaskRoute:    route(taskHandler.Create),
		receptor.TasksRoute:         route(taskHandler.GetAll),
		receptor.TasksByDomainRoute: route(taskHandler.GetAllByDomain),
		receptor.GetTaskRoute:       route(taskHandler.GetByGuid),
		receptor.DeleteTaskRoute:    route(taskHandler.Delete),
		receptor.CancelTaskRoute:    route(taskHandler.Cancel),

		// DesiredLRPs
		receptor.CreateDesiredLRPRoute:    route(desiredLRPHandler.Create),
		receptor.GetDesiredLRPRoute:       route(desiredLRPHandler.Get),
		receptor.UpdateDesiredLRPRoute:    route(desiredLRPHandler.Update),
		receptor.DeleteDesiredLRPRoute:    route(desiredLRPHandler.Delete),
		receptor.DesiredLRPsRoute:         route(desiredLRPHandler.GetAll),
		receptor.DesiredLRPsByDomainRoute: route(desiredLRPHandler.GetAllByDomain),

		// ActualLRPs
		receptor.ActualLRPsRoute:                         route(actualLRPHandler.GetAll),
		receptor.ActualLRPsByDomainRoute:                 route(actualLRPHandler.GetAllByDomain),
		receptor.ActualLRPsByProcessGuidRoute:            route(actualLRPHandler.GetAllByProcessGuid),
		receptor.ActualLRPByProcessGuidAndIndexRoute:     route(actualLRPHandler.GetByProcessGuidAndIndex),
		receptor.KillActualLRPByProcessGuidAndIndexRoute: route(actualLRPHandler.KillByProcessGuidAndIndex),

		// Cells
		receptor.CellsRoute: route(cellHandler.GetAll),

		// Fresh domains
		receptor.BumpFreshDomainRoute: route(freshDomainHandler.Bump),
		receptor.FreshDomainsRoute:    route(freshDomainHandler.GetAll),
	}

	handler, err := rata.NewRouter(receptor.Routes, actions)
	if err != nil {
		panic("unable to create router: " + err.Error())
	}

	if username != "" {
		handler = BasicAuthWrap(handler, username, password)
	}

	if corsEnabled {
		handler = CORSWrapper(handler)
	}

	handler = LogWrap(handler, logger)

	return handler
}

func route(f func(w http.ResponseWriter, r *http.Request)) http.Handler {
	return http.HandlerFunc(f)
}
