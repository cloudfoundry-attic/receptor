package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/cloudfoundry-incubator/receptor"
	Bbs "github.com/cloudfoundry-incubator/runtime-schema/bbs"
	"github.com/cloudfoundry-incubator/runtime-schema/models"
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

	task, err := newTaskFromCreateRequest(taskRequest)
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

func newTaskFromCreateRequest(req receptor.CreateTaskRequest) (models.Task, error) {
	task := models.Task{
		TaskGuid:   req.TaskGuid,
		Domain:     req.Domain,
		Actions:    req.Actions,
		Stack:      req.Stack,
		MemoryMB:   req.MemoryMB,
		DiskMB:     req.DiskMB,
		CpuPercent: req.CpuPercent,
		Log:        req.Log,
		Annotation: req.Annotation,
	}

	err := task.Validate()
	if err != nil {
		return models.Task{}, err
	}
	return task, nil
}
