package task_handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/cloudfoundry-incubator/runtime-schema/models"
	"github.com/pivotal-golang/lager"
)

var ErrWatchFailed = errors.New("watching for completed tasks failed")

func NewHandler(enqueue chan<- models.Task, logger lager.Logger) http.Handler {
	return &TaskHandler{
		enqueue: enqueue,
		logger:  logger.Session("task-watcher"),
	}
}

type TaskHandler struct {
	enqueue chan<- models.Task
	logger  lager.Logger
}

func (t *TaskHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var task models.Task
	err := json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	t.enqueue <- task

	w.WriteHeader(http.StatusAccepted)
}
