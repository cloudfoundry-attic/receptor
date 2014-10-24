package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/receptor/serialization"
	Bbs "github.com/cloudfoundry-incubator/runtime-schema/bbs"
	"github.com/cloudfoundry-incubator/runtime-schema/models"
	"github.com/cloudfoundry/storeadapter"
	"github.com/pivotal-golang/lager"
)

type TaskHandler struct {
	bbs    Bbs.ReceptorBBS
	logger lager.Logger
}

func NewTaskHandler(bbs Bbs.ReceptorBBS, logger lager.Logger) *TaskHandler {
	return &TaskHandler{
		bbs:    bbs,
		logger: logger,
	}
}

func (h *TaskHandler) Create(w http.ResponseWriter, r *http.Request) {
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

	task, err := serialization.TaskFromRequest(taskRequest)
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

func (h *TaskHandler) GetAll(w http.ResponseWriter, req *http.Request) {
	tasks, err := h.bbs.GetAllTasks()
	writeTaskResponse(w, h.logger.Session("get-all-tasks-handler"), tasks, err)
}

func (h *TaskHandler) GetAllByDomain(w http.ResponseWriter, req *http.Request) {
	tasks, err := h.bbs.GetAllTasksByDomain(req.FormValue(":domain"))
	writeTaskResponse(w, h.logger.Session("get-tasks-by-domain-handler"), tasks, err)
}

func (h *TaskHandler) GetByGuid(w http.ResponseWriter, req *http.Request) {
	task, err := h.bbs.GetTaskByGuid(req.FormValue(":task_guid"))
	if err == storeadapter.ErrorKeyNotFound {
		h.logger.Error("failed-to-fetch-task", err)
		writeJSONResponse(w, http.StatusNotFound, receptor.Error{
			Type:    receptor.TaskNotFound,
			Message: "task guid not found",
		})
		return
	} else if err != nil {
		h.logger.Error("failed-to-fetch-task", err)
		writeUnknownErrorResponse(w, err)
		return
	}

	writeJSONResponse(w, http.StatusOK, serialization.TaskToResponse(task))
}

func (h *TaskHandler) Delete(w http.ResponseWriter, req *http.Request) {
	guid := req.FormValue(":task_guid")

	err := h.bbs.ResolvingTask(guid)
	switch err {
	case nil:
	case Bbs.ErrTaskNotFound:
		h.logger.Error("task-not-found", err)
		writeJSONResponse(w, http.StatusConflict, receptor.Error{
			Type:    receptor.TaskNotFound,
			Message: "stuff",
		})
	case Bbs.ErrTaskNotResolvable:
		h.logger.Error("task-not-completed", err)
		writeJSONResponse(w, http.StatusConflict, receptor.Error{
			Type:    receptor.TaskNotDeletable,
			Message: "This task has not been completed. Please retry when it is completed.",
		})
	default:
		h.logger.Error("failed-to-mark-task-resolving", err)
		writeUnknownErrorResponse(w, err)
	}

	err = h.bbs.ResolveTask(guid)
	if err != nil {
		h.logger.Error("failed-to-resolve-task", err)
		writeUnknownErrorResponse(w, err)
	}
}

func writeTaskResponse(w http.ResponseWriter, logger lager.Logger, tasks []models.Task, err error) {
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
