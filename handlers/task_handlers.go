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

func NewCreateTaskHandler(bbs Bbs.ReceptorBBS, logger lager.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log := logger.Session("create-task-handler")
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

		err = bbs.DesireTask(task)
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
	})
}

func NewGetAllTasksHandler(bbs Bbs.ReceptorBBS, logger lager.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		tasks, err := bbs.GetAllTasks()
		writeTaskResponse(w, logger.Session("get-all-tasks-handler"), tasks, err)
	})
}

func NewGetAllTasksByDomainHandler(bbs Bbs.ReceptorBBS, logger lager.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		tasks, err := bbs.GetAllTasksByDomain(req.FormValue(":domain"))
		writeTaskResponse(w, logger.Session("get-tasks-by-domain-handler"), tasks, err)
	})
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
