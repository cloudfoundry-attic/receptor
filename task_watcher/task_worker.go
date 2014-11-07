package task_watcher

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/cloudfoundry-incubator/receptor/serialization"
	Bbs "github.com/cloudfoundry-incubator/runtime-schema/bbs"
	"github.com/cloudfoundry-incubator/runtime-schema/models"
	"github.com/pivotal-golang/lager"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/grouper"
)

func newTaskWorkerPool(poolSize int, taskQueue <-chan models.Task, bbs Bbs.ReceptorBBS, logger lager.Logger) ifrit.Runner {
	members := make(grouper.Members, poolSize)
	for i := 0; i < poolSize; i++ {
		name := fmt.Sprintf("task-worker-%d", i)
		members[i].Name = name
		members[i].Runner = newTaskWorker(taskQueue, bbs, logger.Session(name))
	}
	return grouper.NewParallel(os.Interrupt, members)
}

func newTaskWorker(taskQueue <-chan models.Task, bbs Bbs.ReceptorBBS, logger lager.Logger) *taskWorker {
	return &taskWorker{
		taskQueue:  taskQueue,
		bbs:        bbs,
		logger:     logger,
		httpClient: http.Client{},
	}
}

type taskWorker struct {
	taskQueue  <-chan models.Task
	bbs        Bbs.ReceptorBBS
	logger     lager.Logger
	httpClient http.Client
}

func (t *taskWorker) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	t.logger.Debug("starting")
	close(ready)
	for {
		select {
		case task := <-t.taskQueue:
			t.handleCompletedTask(task)
		case <-signals:
			t.logger.Debug("exited")
			return nil
		}
	}
}

func (t *taskWorker) handleCompletedTask(task models.Task) {
	logger := t.logger.WithData(lager.Data{"task-guid": task.TaskGuid})

	if task.CompletionCallbackURL != nil {
		logger.Info("resolving-task")
		err := t.bbs.ResolvingTask(task.TaskGuid)
		if err != nil {
			logger.Error("marking-task-as-resolving-failed", err)
			return
		}

		logger = logger.WithData(lager.Data{"callback_url": task.CompletionCallbackURL.String()})

		json, err := json.Marshal(serialization.TaskToResponse(task))
		if err != nil {
			logger.Error("marshalling-task-failed", err)
			return
		}

		var statusCode int

		for i := 0; i < MAX_RETRIES; i++ {
			request, err := http.NewRequest("POST", task.CompletionCallbackURL.String(), bytes.NewReader(json))
			if err != nil {
				logger.Error("building-request-failed", err)
				return
			}

			request.Header.Set("Content-Type", "application/json")

			response, err := t.httpClient.Do(request)
			if err != nil {
				logger.Error("doing-request-failed", err)
				return
			}

			statusCode = response.StatusCode
			if shouldResolve(statusCode) {
				err = t.bbs.ResolveTask(task.TaskGuid)
				if err != nil {
					logger.Error("resolving-task-failed", err)
					return
				}

				logger.Info("resolved-task", lager.Data{"status_code": statusCode})
				return
			}
		}

		logger.Info("callback-failed", lager.Data{"status_code": statusCode})
	}
}

func shouldResolve(status int) bool {
	switch status {
	case http.StatusServiceUnavailable, http.StatusGatewayTimeout:
		return false
	default:
		return true
	}
}
