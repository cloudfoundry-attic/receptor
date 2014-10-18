package handlers

import (
	"net/http"

	Bbs "github.com/cloudfoundry-incubator/runtime-schema/bbs"
	"github.com/pivotal-golang/lager"
	"github.com/tedsuo/rata"
)

const (
	CreateTask = "CreateTask"
)

var Routes = rata.Routes{
	{Path: "/tasks", Method: "POST", Name: CreateTask},
}

func New(bbs Bbs.ReceptorBBS, logger lager.Logger, username, password string) http.Handler {
	actions := rata.Handlers{
		CreateTask: NewCreateTaskHandler(bbs, logger),
	}

	handler, err := rata.NewRouter(Routes, actions)
	if err != nil {
		panic("unable to create router: " + err.Error())
	}

	if username != "" {
		handler = BasicAuthWrap(handler, username, password)
	}

	handler = LogWrap(handler, logger)

	return handler
}
