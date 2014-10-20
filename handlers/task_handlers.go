package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/cloudfoundry-incubator/receptor"
	Bbs "github.com/cloudfoundry-incubator/runtime-schema/bbs"
	"github.com/cloudfoundry/storeadapter"
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
	taskRequest := receptor.CreateTaskRequest{}

	err := json.NewDecoder(r.Body).Decode(&taskRequest)
	if err != nil {
		log.Error("invalid-json", err)
		writeJSONResponse(w, http.StatusBadRequest, receptor.Error{
			Type:    receptor.InvalidJSON,
			Message: err.Error(),
		})
		return
	}

	task, err := taskRequest.ToTask()
	if err != nil {
		log.Error("task-request-invalid", err)
		writeJSONResponse(w, http.StatusBadRequest, receptor.Error{
			Type:    receptor.InvalidTask,
			Message: err.Error(),
		})
		return
	}

	err = h.bbs.DesireTask(task)
	if err != nil {
		log.Error("desire-task-failed", err)
		if err == storeadapter.ErrorKeyExists {
			writeJSONResponse(w, http.StatusConflict, receptor.Error{
				Type:    receptor.TaskGuidAlreadyExists,
				Message: "task already exists",
			})
		} else {
			writeUnknownErrorResponse(w, err)
		}
		return
	}

	log.Info("created", lager.Data{"task-guid": task.TaskGuid})
	w.WriteHeader(http.StatusCreated)
}
