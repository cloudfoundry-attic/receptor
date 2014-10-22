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

	writeJSONResponse(w, http.StatusOK, responseFromTask(task))
}

func writeTaskResponse(w http.ResponseWriter, logger lager.Logger, tasks []models.Task, err error) {
	if err != nil {
		logger.Error("failed-to-fetch-tasks", err)
		writeUnknownErrorResponse(w, err)
		return
	}

	taskResponses := make([]receptor.TaskResponse, 0, len(tasks))
	for _, task := range tasks {
		taskResponses = append(taskResponses, responseFromTask(task))
	}

	writeJSONResponse(w, http.StatusOK, taskResponses)
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
		ResultFile: req.ResultFile,
		Annotation: req.Annotation,
	}

	err := task.Validate()
	if err != nil {
		return models.Task{}, err
	}
	return task, nil
}

func responseFromTask(task models.Task) receptor.TaskResponse {
	return receptor.TaskResponse{
		TaskGuid:   task.TaskGuid,
		Domain:     task.Domain,
		Actions:    task.Actions,
		Stack:      task.Stack,
		MemoryMB:   task.MemoryMB,
		DiskMB:     task.DiskMB,
		CpuPercent: task.CpuPercent,
		Log:        task.Log,
		Annotation: task.Annotation,
	}
}
