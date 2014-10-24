package task_watcher

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"

	"github.com/cloudfoundry-incubator/receptor/serialization"
	Bbs "github.com/cloudfoundry-incubator/runtime-schema/bbs"
	"github.com/cloudfoundry-incubator/runtime-schema/models"
	"github.com/pivotal-golang/lager"
	"github.com/tedsuo/ifrit"
)

const MAX_RETRIES = 3

func New(bbs Bbs.ReceptorBBS, logger lager.Logger) ifrit.Runner {
	return &taskWatcher{
		bbs:        bbs,
		logger:     logger,
		httpClient: http.Client{},
	}
}

type taskWatcher struct {
	bbs        Bbs.ReceptorBBS
	logger     lager.Logger
	httpClient http.Client
}

func (t *taskWatcher) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	close(ready)

	tasks, stopChan, _ := t.bbs.WatchForCompletedTask()

loop:
	for {
		select {
		case task := <-tasks:
			t.handleCompletedTask(task)
		case <-signals:
			break loop
		}
	}

	stopChan <- true
	return nil
}

func (t *taskWatcher) handleCompletedTask(task models.Task) {
	err := t.bbs.ResolvingTask(task.TaskGuid)
	if err != nil {
		t.logger.Error("marking-task-as-resolving-failed", err)
		return
	}

	if task.CompletionCallbackURL != nil {
		for i := 0; i < MAX_RETRIES; i++ {
			json, err := json.Marshal(serialization.TaskToResponse(task))
			if err != nil {
				t.logger.Error("marshalling-task-failed", err)
				return
			}

			request, err := http.NewRequest("POST", task.CompletionCallbackURL.String(), bytes.NewReader(json))
			if err != nil {
				t.logger.Error("building-request-failed", err)
				return
			}

			request.Header.Set("Content-Type", "application/json")

			response, err := t.httpClient.Do(request)
			if err != nil {
				t.logger.Error("doing-request-failed", err)
				return
			}

			if isSuccess(response.StatusCode) || isBadRequest(response.StatusCode) {
				err = t.bbs.ResolveTask(task.TaskGuid)
				if err != nil {
					t.logger.Error("resolving-task-failed", err)
					return
				}

				t.logger.Info("resolved-task")
				return
			}
		}
	}
}

func isSuccess(status int) bool {
	return 200 <= status && status < 300
}

func isBadRequest(status int) bool {
	return 400 <= status && status < 500
}
