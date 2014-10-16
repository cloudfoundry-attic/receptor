package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/cloudfoundry-incubator/receptor/api"
	Bbs "github.com/cloudfoundry-incubator/runtime-schema/bbs"
	"github.com/pivotal-golang/lager"
)

type createTaskHandler struct {
	bbs    Bbs.ReceptorBBS
	logger lager.Logger
}

func NewCreateTaskHandler(bbs Bbs.ReceptorBBS, logger lager.Logger) http.Handler {
	return &createTaskHandler{
		bbs:    bbs,
		logger: logger,
	}
}

func (h *createTaskHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log := h.logger.Session("create-task-handler")
	taskRequest := api.CreateTaskRequest{}

	err := json.NewDecoder(r.Body).Decode(&taskRequest)
	if err != nil {
		log.Error("invalid-json", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write(api.NewErrorResponse(err).JSONReader().Bytes())
		return
	}

	task, err := taskRequest.ToTask()
	if err != nil {
		log.Error("task-request-invalid", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write(api.NewErrorResponse(err).JSONReader().Bytes())
		return
	}

	err = h.bbs.DesireTask(task)
	if err != nil {
		log.Error("desire-task-failed", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(api.NewErrorResponse(err).JSONReader().Bytes())
		return
	}

	log.Info("created", lager.Data{"task-guid": task.TaskGuid})
	w.WriteHeader(http.StatusCreated)
}
