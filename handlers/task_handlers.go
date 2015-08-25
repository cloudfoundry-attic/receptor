package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/cloudfoundry-incubator/bbs"
	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/receptor/serialization"
	"github.com/pivotal-golang/lager"
)

type TaskHandler struct {
	bbs    bbs.Client
	logger lager.Logger
}

func NewTaskHandler(bbs bbs.Client, logger lager.Logger) *TaskHandler {
	return &TaskHandler{
		bbs:    bbs,
		logger: logger.Session("task-handler"),
	}
}

func (h *TaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	log := h.logger.Session("create")
	taskRequest := receptor.TaskCreateRequest{}

	err := json.NewDecoder(r.Body).Decode(&taskRequest)
	if err != nil {
		log.Error("invalid-json", err)
		writeJSONResponse(w, http.StatusBadRequest, receptor.Error{
			Type:    receptor.InvalidJSON,
			Message: err.Error(),
		})
		return
	}

	task, err := serialization.TaskFromRequest(taskRequest)
	if err == nil {
		if task.GetCompletionCallbackUrl() != "" {
			_, err = url.ParseRequestURI(task.GetCompletionCallbackUrl())
		}
	}
	if err != nil {
		log.Error("task-request-invalid", err)
		writeJSONResponse(w, http.StatusBadRequest, receptor.Error{
			Type:    receptor.InvalidTask,
			Message: err.Error(),
		})
		return
	}

	log.Debug("creating-task", lager.Data{"task-guid": task.TaskGuid})

	err = h.bbs.DesireTask(task.TaskGuid, task.Domain, task.TaskDefinition)
	if err != nil {
		log.Error("failed-to-desire-task", err)
		if mErr, ok := err.(*models.Error); ok {
			if mErr.Equal(models.ErrBadRequest) {
				writeJSONResponse(w, http.StatusBadRequest, receptor.Error{
					Type:    receptor.InvalidTask,
					Message: err.Error(),
				})
				return
			} else if mErr.Equal(models.ErrResourceExists) {
				writeJSONResponse(w, http.StatusConflict, receptor.Error{
					Type:    receptor.TaskGuidAlreadyExists,
					Message: "task already exists",
				})
				return
			}
		}
		writeUnknownErrorResponse(w, err)
		return
	}

	log.Info("created", lager.Data{"task-guid": task.TaskGuid})
	w.WriteHeader(http.StatusCreated)
}

func (h *TaskHandler) GetAll(w http.ResponseWriter, req *http.Request) {
	domain := req.FormValue("domain")
	logger := h.logger.Session("get-all", lager.Data{
		"domain": domain,
	})

	var tasks []*models.Task
	var err error

	if domain == "" {
		tasks, err = h.bbs.Tasks()
	} else {
		tasks, err = h.bbs.TasksByDomain(domain)
	}

	writeTaskResponse(w, logger, tasks, err)
}

func (h *TaskHandler) GetByGuid(w http.ResponseWriter, req *http.Request) {
	guid := req.FormValue(":task_guid")
	logger := h.logger.Session("get-by-guid", lager.Data{
		"TaskGuid": guid,
	})

	if guid == "" {
		err := errors.New("task_guid missing from request")
		logger.Error("missing-task-guid", err)
		writeBadRequestResponse(w, receptor.InvalidRequest, err)
		return
	}

	task, err := h.bbs.TaskByGuid(guid)
	if models.ErrResourceNotFound.Equal(err) {
		writeTaskNotFoundResponse(w, guid)
		return
	}

	if err != nil {
		h.logger.Error("failed-to-fetch-task", err)
		writeUnknownErrorResponse(w, err)
		return
	}

	writeJSONResponse(w, http.StatusOK, serialization.TaskToResponse(task))
}

func (h *TaskHandler) Delete(w http.ResponseWriter, req *http.Request) {
	guid := req.FormValue(":task_guid")

	err := h.bbs.ResolvingTask(guid)
	if err != nil {
		if modelErr, ok := err.(*models.Error); ok {
			switch modelErr.Type {
			case models.InvalidStateTransition:
				h.logger.Error("invalid-task-state-transition", modelErr)
				writeJSONResponse(w, http.StatusConflict, receptor.Error{
					Type:    receptor.TaskNotDeletable,
					Message: "This task has not been completed. Please retry when it is completed.",
				})
				return
			case models.ResourceNotFound:
				h.logger.Error("task-not-found", modelErr)
				writeTaskNotFoundResponse(w, guid)
				return
			}
		}

		h.logger.Error("failed-to-mark-task-resolving", err)
		writeUnknownErrorResponse(w, err)
		return
	}

	err = h.bbs.DeleteTask(guid)
	if err != nil {
		h.logger.Error("failed-to-delete-task", err)
		writeUnknownErrorResponse(w, err)
	}
}

func (h *TaskHandler) Cancel(w http.ResponseWriter, req *http.Request) {
	guid := req.FormValue(":task_guid")

	err := h.bbs.CancelTask(guid)
	if err != nil {
		if models.ErrResourceNotFound.Equal(err) {
			h.logger.Error("failed-to-cancel-task", err)
			writeTaskNotFoundResponse(w, guid)
			return
		}

		h.logger.Error("failed-to-fetch-task", err)
		writeUnknownErrorResponse(w, err)
	}
}

func writeTaskResponse(w http.ResponseWriter, logger lager.Logger, tasks []*models.Task, err error) {
	if err != nil {
		logger.Error("failed-to-fetch-tasks", err)
		writeUnknownErrorResponse(w, err)
		return
	}

	taskResponses := make([]receptor.TaskResponse, 0, len(tasks))
	for _, task := range tasks {
		taskResponses = append(taskResponses, serialization.TaskToResponse(task))
	}

	writeJSONResponse(w, http.StatusOK, taskResponses)
}

func writeTaskNotFoundResponse(w http.ResponseWriter, taskGuid string) {
	writeJSONResponse(w, http.StatusNotFound, receptor.Error{
		Type:    receptor.TaskNotFound,
		Message: fmt.Sprintf("task with guid '%s' not found", taskGuid),
	})
}
