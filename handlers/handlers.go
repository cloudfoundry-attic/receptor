package handlers

import (
	"net/http"

	"github.com/cloudfoundry-incubator/bbs"
	"github.com/cloudfoundry-incubator/locket"
	"github.com/cloudfoundry-incubator/receptor"
	"github.com/pivotal-golang/lager"
	"github.com/tedsuo/rata"
)

func New(bbs bbs.Client, locketClient locket.Client, logger lager.Logger, username, password string, corsEnabled bool, artifactLocator ArtifactLocator) http.Handler {
	taskHandler := NewTaskHandler(bbs, logger)
	desiredLRPHandler := NewDesiredLRPHandler(bbs, logger)
	actualLRPHandler := NewActualLRPHandler(bbs, logger)
	cellHandler := NewCellHandler(locketClient, logger)
	domainHandler := NewDomainHandler(bbs, logger)
	syncHandler := NewSyncHandler(artifactLocator, logger)
	eventStreamHandler := NewEventStreamHandler(bbs, logger)
	authCookieHandler := NewAuthCookieHandler(logger)

	auth := func(handler func(http.ResponseWriter, *http.Request)) http.Handler {
		if username == "" {
			return http.HandlerFunc(handler)
		}
		return CookieAuthWrap(BasicAuthWrap(http.HandlerFunc(handler), username, password), receptor.AuthorizationCookieName)
	}

	actions := rata.Handlers{
		// Tasks
		receptor.CreateTaskRoute: auth(taskHandler.Create),
		receptor.TasksRoute:      auth(taskHandler.GetAll),
		receptor.GetTaskRoute:    auth(taskHandler.GetByGuid),
		receptor.DeleteTaskRoute: auth(taskHandler.Delete),
		receptor.CancelTaskRoute: auth(taskHandler.Cancel),

		// DesiredLRPs
		receptor.CreateDesiredLRPRoute: auth(desiredLRPHandler.Create),
		receptor.GetDesiredLRPRoute:    auth(desiredLRPHandler.Get),
		receptor.UpdateDesiredLRPRoute: auth(desiredLRPHandler.Update),
		receptor.DeleteDesiredLRPRoute: auth(desiredLRPHandler.Delete),
		receptor.DesiredLRPsRoute:      auth(desiredLRPHandler.GetAll),

		// ActualLRPs
		receptor.ActualLRPsRoute:                         auth(actualLRPHandler.GetAll),
		receptor.ActualLRPsByProcessGuidRoute:            auth(actualLRPHandler.GetAllByProcessGuid),
		receptor.ActualLRPByProcessGuidAndIndexRoute:     auth(actualLRPHandler.GetByProcessGuidAndIndex),
		receptor.KillActualLRPByProcessGuidAndIndexRoute: auth(actualLRPHandler.KillByProcessGuidAndIndex),

		// Cells
		receptor.CellsRoute: auth(cellHandler.GetAll),

		// Domains
		receptor.UpsertDomainRoute: auth(domainHandler.Upsert),
		receptor.DomainsRoute:      auth(domainHandler.GetAll),

		// Sync
		receptor.DownloadRoute: http.HandlerFunc(syncHandler.Download),

		// Event Streaming
		receptor.EventStream: auth(eventStreamHandler.EventStream),

		// Authentication Cookie
		receptor.GenerateCookie: auth(authCookieHandler.GenerateCookie),
	}

	handler, err := rata.NewRouter(receptor.Routes, actions)
	if err != nil {
		panic("unable to create router: " + err.Error())
	}

	if corsEnabled {
		handler = CORSWrapper(handler)
	}

	return LogWrap(handler, logger)
}
